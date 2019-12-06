# Packer template for AWS Ami with Vault

A Packer project to build an AWS AMI with Vault installed.

The the ami will have scripts that can be used to initialize and unseal vault located in `/etc/vault.d/scripts/`.

**Note** that the initialization script will write out the unseal keys and the initial Vault root token to files in `/etc/vault.d/` so that the unseal script can use them afterwards which is absolutely insecure. The script's purpose is to automate these actions which are intended to be carried out manually but at the cost of greatly reducing security.

## Prerequisites

* Install [packer](https://www.packer.io/downloads.html).
* Configure [AWS SDK credentials](https://docs.aws.amazon.com/sdk-for-java/v1/developer-guide/credentials.html) using environment variables and also set `AWS_REGION` environment variable to the region you want to create your ami in.

## Building the box with Packer

The packer template is in `template.json`. In the `variables` section you can set parameters to customize the build. Help on setting, overriding variables in packer can be found [here](https://www.packer.io/docs/templates/user-variables.html#setting-variables).

* `vault_ver` - the version of Vault to install. If it is set to an empty string, the latest version will be installed.
* `aws_region` - the AWS region in which to build the AMI. Will default to the value of `AWS_REGION` environment variable
* `base_ami_id`  - the base ami to use. Needs to be the in the region configured with `aws_region` variable.
* `tag_owner` - set the value of an AWS tag named `owner` which will be applied to the ami.
* `build_name` - used internally to set parameters of the packer builder. Usually no need to change it.

Run `packer validate template.json` - to make basic template validation.

Run `packer build -var "aws_region=eu-central-1" -var "base_ami_id=ami-09356619876445425" template.json` - to build the Vagrant box with packer.

## Testing [terratest](https://github.com/gruntwork-io/terratest/)

The project includes a test using the [terratest](https://github.com/gruntwork-io/terratest/) library with the Golang test framework.

The test will: 

1. If `AWS_REGION` is not set choose a random AWS Region to run the test in.
2. Create the AMI using the packer template.
3. Will provision an AWS instance with this AMI using the terraform configuration located in `test/terraform` 
4. Will run tests against this instance.
5. Will clean-up the created AWS resources.

**Note:** Currently the test will run in a subnet from the Default VPC in the selected region.

## Prerequisites

1. Install [terraform](https://www.terraform.io/downloads.html) >= 12.0.
2. Install [Golang](https://golang.org/dl/) >= 1.13 if not already installed.
3. Install dependency golang packages - `go get -v -d -t ./test/...`.
4. Configure AWS SDK credentials.

## Running the test

Run `go test -v -timeout=60m ./test/`.
