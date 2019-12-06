variable "aws_region" {
  type = string
}

variable "key_pair" {
  type = string
}

variable "vault_ami_id" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "ssh_private_key" {
  type = string
}

variable "common_tags" {
  type = map(string)
  default = {
    owner   = "terratest"
    project = "terratest-ami-vault"
  }
}
