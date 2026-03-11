variable "aws_region" {
  type        = string
  default     = "us-east-1"
  description = "AWS region for all resources"
}

variable "environment" {
  type        = string
  default     = "production"
  description = "Environment name"
}

variable "domain" {
  type        = string
  description = "Root domain (e.g. hivemind.rithul.dev)"
}

variable "registry_fqdn" {
  type        = string
  description = "Registry app FQDN (e.g. registry.hivemind.rithul.dev)"
}

variable "packages_fqdn" {
  type        = string
  description = "Packages/CloudFront FQDN (e.g. packages.hivemind.rithul.dev)"
}

variable "ssh_allowed_cidrs" {
  type        = list(string)
  description = "CIDR blocks allowed to SSH to EC2 (e.g. [\"1.2.3.4/32\"])"
}

variable "key_pair_name" {
  type        = string
  default     = null
  description = "Existing EC2 key pair name. If null, a key is generated and private key output (add to GitHub REGISTRY_DEPLOY_KEY)"
}

variable "create_key_pair" {
  type        = bool
  default     = true
  description = "If true and key_pair_name is null, create a key pair and output private key"
}

variable "github_org_repo" {
  type        = string
  description = "GitHub org/repo for OIDC trust (e.g. rithulkamesh/hivemind)"
}

variable "github_branch" {
  type        = string
  default     = "main"
  description = "Branch allowed to assume deploy role"
}

variable "ecr_api_repository_name" {
  type        = string
  default     = "hivemind-registry-api"
  description = "ECR repository name for API image"
}

variable "s3_packages_bucket" {
  type        = string
  default     = "hivemind-registry-packages"
  description = "S3 bucket name for package artifacts"
}

variable "ses_config_set" {
  type        = string
  default     = "hivemind-registry"
  description = "SES configuration set name"
}

variable "route53_zone_id" {
  type        = string
  default     = null
  description = "Route53 hosted zone ID for domain. If null, zone is looked up by domain name"
}

variable "acm_certificate_arn" {
  type        = string
  default     = null
  description = "ACM certificate ARN (us-east-1) for CloudFront. If null, looked up by domain"
}

variable "terraform_state_bucket_name" {
  type        = string
  default     = "hivemind-terraform-state"
  description = "S3 bucket name for Terraform state (GitHub Actions needs access)"
}

variable "terraform_lock_table_name" {
  type        = string
  default     = "hivemind-terraform-locks"
  description = "DynamoDB table name for state locking (GitHub Actions needs access)"
}
