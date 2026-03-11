output "instance_id" {
  value = aws_instance.main.id
}

output "public_ip" {
  value = aws_eip.main.public_ip
}

output "private_key_openssh" {
  value     = var.create_key_pair ? tls_private_key.ec2[0].private_key_openssh : null
  sensitive = true
}

output "key_name" {
  value = local.key_name
}
