output "ec2_instance_profile_name" {
  value = aws_iam_instance_profile.ec2.name
}

output "ec2_instance_profile_arn" {
  value = aws_iam_instance_profile.ec2.arn
}

output "github_deploy_role_arn" {
  value = aws_iam_role.github_deploy.arn
}
