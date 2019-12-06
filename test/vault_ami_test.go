package test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	terratest_aws "github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/packer"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/ssh"
	"github.com/gruntwork-io/terratest/modules/terraform"
)

// Occasionally, a Packer build may fail due to intermittent issues (e.g., brief network outage or EC2 issue). We try
// to make our tests resilient to that by specifying those known common errors here and telling our builds to retry if
// they hit those errors.
var DefaultRetryablePackerErrors = map[string]string{
	"Script disconnected unexpectedly": "Occasionally, Packer seems to lose connectivity to AWS, perhaps due to a brief network outage",
	"exit status 1":                    "Occasionally, provision scripts will fail because of the apt-get",
}

var DefaultTimeBetweenPackerRetries = 15 * time.Second

const DefaultMaxPackerRetries = 3

// Timestamp based string to be used in generating unique resource names
var timestamp = time.Now().Format("200601021504")

// TestPackerTemplate will test the packer template by
// 1. Build the AMI using packer
// 2. Provision an AWS Instance from this AMI
// 3. Run tests agains the AWS Instance
// 4. Clean-up all created AWS resources
func TestPackerTemplate(t *testing.T) {

	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = terratest_aws.GetRandomStableRegion(t, nil, nil)
		t.Logf("Random AWS REGION selected: %q", awsRegion)
	}

	awsDefultVPC, err := terratest_aws.GetDefaultVpcE(t, awsRegion)
	if err != nil {
		t.Fatal(err)
	}

	// build AMI with packer
	amiID, err := BuildAmi(t, awsRegion)
	if err != nil {
		t.Fatal(err)
	}

	defer terratest_aws.DeleteAmiAndAllSnapshots(t, awsRegion, amiID)
	t.Logf("created ami: %s", amiID)

	// create a temporary AWS key pair
	keyPair, err := terratest_aws.CreateAndImportEC2KeyPairE(t, awsRegion, fmt.Sprintf("terratest-ami-vault-%s", timestamp))
	if err != nil {
		t.Fatal(err)
	}
	defer terratest_aws.DeleteEC2KeyPair(t, keyPair)
	t.Logf("created KeyPair: %s", keyPair.Name)

	// create AWS instance with terraform based on amiID.
	tfOpt := &terraform.Options{
		TerraformDir: "./terraform",
		NoColor:      true,
		Vars: map[string]interface{}{
			"key_pair":        keyPair.Name,
			"vault_ami_id":    amiID,
			"aws_region":      awsRegion,
			"vpc_id":          awsDefultVPC.Id,
			"ssh_private_key": keyPair.PrivateKey,
		},
	}

	_, err = terraform.InitAndApplyE(t, tfOpt)
	defer terraform.Destroy(t, tfOpt)
	if err != nil {
		t.Fatal(err)
	}

	// give some time to the Vault service to start
	time.Sleep(30 * time.Second)

	// run test on the instance created by terraform
	host := ssh.Host{
		Hostname:    terraform.Output(t, tfOpt, "vault_public_ip"),
		SshUserName: "ubuntu",
		SshKeyPair:  keyPair.KeyPair,
	}

	// Some times ssh fails intermittently so retrying a few times
	maxRetries := 5
	timeBetweenRetries := 5 * time.Second
	description := fmt.Sprintf("SSH to public host %s", host.Hostname)

	// execute 'vault_init.sh' and check output
	retry.DoWithRetry(t, description, maxRetries, timeBetweenRetries, func() (string, error) {
		t.Log("Running SSH: '/etc/vault.d/scripts/vault_init.sh'")
		runVaultInit, err := ssh.CheckSshCommandE(t, host, "/etc/vault.d/scripts/vault_init.sh")
		if err != nil {
			if strings.Contains(err.Error(), "Process exited with status") {
				t.Logf("'vault_init.sh' got stdout/stderr:\n%s", runVaultInit)
				t.Fatal(err)
			}
			return "", err
		}

		if strings.TrimSpace(runVaultInit) != "" {
			t.Fatalf("Unexpected 'vault_init.sh' stdout/stderr, want: '', got: %q", runVaultInit)
		}

		return "", nil
	})

	// execute 'vault_unseal.sh'
	retry.DoWithRetry(t, description, maxRetries, timeBetweenRetries, func() (string, error) {
		t.Log("Running SSH: '/etc/vault.d/scripts/vault_unseal.sh'")
		runVaultUnseal, err := ssh.CheckSshCommandE(t, host, "/etc/vault.d/scripts/vault_unseal.sh")
		if err != nil {
			if strings.Contains(err.Error(), "Process exited with status") {
				t.Logf("'vault_unseal.sh' got stdout/stderr:\n%s", runVaultUnseal)
				t.Fatal(err)
			}
			return "", err
		}

		return "", nil
	})

	// execute 'vault login' and confirm success
	retry.DoWithRetry(t, description, maxRetries, timeBetweenRetries, func() (string, error) {
		t.Log("Running SSH: 'VAULT_ADDR='http://127.0.0.1:8200' vault login $(sudo cat /etc/vault.d/.vault-token)'")
		runVaultLogin, err := ssh.CheckSshCommandE(t, host, "VAULT_ADDR='http://127.0.0.1:8200' vault login $(sudo cat /etc/vault.d/.vault-token)")
		if err != nil {
			if strings.Contains(err.Error(), "Process exited with status") {
				t.Logf("'vault login' got stdout/stderr:\n%s", runVaultLogin)
				t.Fatal(err)
			}
			return "", err
		}

		if !strings.Contains(runVaultLogin, "Success! You are now authenticated.") {
			t.Fatalf("Unexpected 'vault login' stdout/stderr, want: 'Success! You are now authenticated.', got: %q", runVaultLogin)
		}

		return "", nil
	})
}

// BuildAmi will build the AMI in the provided awsRegion using packer
func BuildAmi(t *testing.T, awsRegion string) (string, error) {

	// Find latest ubuntu Bionic AMI in the provided region
	filters := map[string][]string{
		"name":                             {"*ubuntu-bionic-18.04-amd64-server-*"},
		"virtualization-type":              {"hvm"},
		"architecture":                     {"x86_64"},
		"root-device-type":                 {"ebs"},
		"block-device-mapping.volume-type": {"gp2"},
	}

	amiID, err := terratest_aws.GetMostRecentAmiIdE(t, awsRegion, terratest_aws.CanonicalAccountId, filters)
	if err != nil {
		return "", err
	}

	packerOptions := &packer.Options{
		// The path to where the Packer template is located
		Template: "../template.json",

		// Variables to pass to our Packer build using -var options
		Vars: map[string]string{
			"tag_owner":   fmt.Sprintf("terratest-packer-vault-%s", timestamp),
			"base_ami_id": amiID,
			"aws_region":  awsRegion,
		},

		// Configure retries for intermittent errors
		RetryableErrors:    DefaultRetryablePackerErrors,
		TimeBetweenRetries: DefaultTimeBetweenPackerRetries,
		MaxRetries:         DefaultMaxPackerRetries,
	}

	return packer.BuildArtifactE(t, packerOptions)
}
