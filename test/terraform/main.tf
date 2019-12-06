provider "aws" {
  region  = var.aws_region
  version = "~> 2.41"
}

resource "aws_security_group" "allow-all" {
  name   = "terratest-ami-vault-${formatdate("YYYYMMDDHHmmss", timestamp())}"
  vpc_id = var.vpc_id
  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "vault" {
  ami                         = var.vault_ami_id
  vpc_security_group_ids      = [aws_security_group.allow-all.id]
  key_name                    = var.key_pair
  associate_public_ip_address = true
  instance_type               = "t2.micro"
  tags                        = var.common_tags

  // use ssh provisioner to confirm that ssh is ready
  connection {
    host        = self.public_dns
    user        = "ubuntu"
    private_key = var.ssh_private_key
  }

  provisioner "local-exec" {
    command = "echo 'ssh is available'"
  }
}

output "vault_public_dns" {
  value = aws_instance.vault.public_dns
}

output "vault_public_ip" {
  value = aws_instance.vault.public_ip
}
