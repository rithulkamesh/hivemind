#!/usr/bin/env bash
set -e

echo "Bootstrapping EC2 instance for Hivemind Registry..."

sudo apt-get update -y
sudo apt-get upgrade -y
sudo apt-get install -y apt-transport-https ca-certificates curl software-properties-common unzip

# Install Docker
if ! command -v docker &> /dev/null; then
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
    sudo usermod -aG docker ubuntu
    rm get-docker.sh
fi

# Install AWS CLI
if ! command -v aws &> /dev/null; then
    curl "https://awscli.amazonaws.com/awscli-exe-linux-aarch64.zip" -o "awscliv2.zip"
    unzip awscliv2.zip
    sudo ./aws/install
    rm -rf aws awscliv2.zip
fi

# Setup directories
sudo mkdir -p /opt/registry
sudo chown ubuntu:ubuntu /opt/registry

echo "Bootstrap complete! Remember to log out and log back in for docker group changes to take effect."
