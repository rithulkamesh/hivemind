output "ec2_elastic_ip" {
  value       = module.ec2.public_ip
  description = "Elastic IP for EC2 — set as GitHub secret REGISTRY_EC2_HOST"
}

output "ec2_instance_id" {
  value = module.ec2.instance_id
}

output "ecr_registry_url" {
  value       = module.ecr.repository_url
  description = "ECR repository URL for API image (e.g. 123456789.dkr.ecr.us-east-1.amazonaws.com/hivemind-registry-api)"
}

output "cloudfront_domain_name" {
  value       = module.s3.cloudfront_domain_name
  description = "CloudFront distribution domain (for packages CNAME target)"
}

output "cloudfront_distribution_id" {
  value = module.s3.cloudfront_distribution_id
}

output "github_oidc_role_arn" {
  value       = module.iam.github_deploy_role_arn
  description = "ARN of the IAM role for GitHub Actions OIDC — use in workflow role-to-assume"
}

output "aws_account_id" {
  value       = data.aws_caller_identity.current.account_id
  description = "AWS account ID (for workflow role ARN)"
}

output "deploy_private_key" {
  value     = module.ec2.private_key_openssh
  sensitive = true
  description = "SSH private key for ec2-user — set as GitHub secret REGISTRY_DEPLOY_KEY (only if create_key_pair = true)"
}

output "ses_dkim_records" {
  value       = module.ses.dkim_tokens
  description = "SES DKIM tokens (CNAME names); Route53 module creates these records"
}

output "s3_packages_bucket" {
  value = module.s3.bucket_id
}
