variable "name_prefix" {
  type        = string
  default     = "hivemind-registry"
  description = "Prefix for IAM resources"
}

variable "s3_packages_bucket_arn" {
  type        = string
  description = "ARN of the S3 packages bucket"
}

variable "cloudfront_distribution_arn" {
  type        = string
  description = "ARN of the CloudFront distribution (for invalidation)"
}

variable "ecr_repository_arn" {
  type        = string
  description = "ARN of the ECR repository"
}

variable "github_org_repo" {
  type        = string
  description = "GitHub org/repo (e.g. rithulkamesh/hivemind)"
}

variable "github_branch" {
  type        = string
  default     = "main"
  description = "Branch allowed to assume deploy role"
}

variable "terraform_state_bucket_arn" {
  type        = string
  default     = null
  description = "ARN of S3 bucket for Terraform state (for infra workflow)"
}

variable "terraform_lock_table_arn" {
  type        = string
  default     = null
  description = "ARN of DynamoDB table for state locking (for infra workflow)"
}
