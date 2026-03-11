resource "tls_private_key" "ec2" {
  count = var.create_key_pair ? 1 : 0

  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "ec2" {
  count = var.create_key_pair ? 1 : 0

  key_name   = "${var.name_prefix}-deploy-key"
  public_key = tls_private_key.ec2[0].public_key_openssh
}

locals {
  key_name = var.key_pair_name != null ? var.key_pair_name : (var.create_key_pair ? aws_key_pair.ec2[0].key_name : null)
  # User data: install docker, docker-compose, clone repo, optional run compose
  user_data_base = <<-EOT
#!/bin/bash
set -e
dnf update -y
dnf install -y docker git
systemctl enable --now docker
usermod -aG docker ec2-user
curl -sL "https://github.com/docker/compose/releases/latest/download/docker-compose-linux-aarch64" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
mkdir -p /opt/hivemind-registry
sudo -u ec2-user git clone --branch ${var.repo_branch} --single-branch --depth 1 ${var.repo_clone_url} /tmp/hivemind-repo
cp -r /tmp/hivemind-repo/registry/deploy/* /opt/hivemind-registry/ || true
chown -R ec2-user:ec2-user /opt/hivemind-registry
rm -rf /tmp/hivemind-repo
EOT
  user_data = var.user_data != "" ? var.user_data : local.user_data_base
}

data "aws_ami" "al2023" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-*-kernel-*-arm64"]
  }

  filter {
    name   = "state"
    values = ["available"]
  }
}

resource "aws_instance" "main" {
  ami                    = data.aws_ami.al2023.id
  instance_type           = "t4g.small"
  subnet_id               = var.subnet_id
  vpc_security_group_ids  = var.security_group_ids
  iam_instance_profile    = var.iam_instance_profile_name
  key_name                = local.key_name
  user_data               = base64encode(local.user_data)
  user_data_replace_on_change = true

  root_block_device {
    volume_size = 20
    volume_type = "gp3"
  }

  tags = {
    Name = "${var.name_prefix}-ec2"
  }
}

resource "aws_eip" "main" {
  domain = "vpc"
  instance = aws_instance.main.id
  tags = {
    Name = "${var.name_prefix}-eip"
  }
}
