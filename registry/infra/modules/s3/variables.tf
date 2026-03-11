variable "bucket_name" {
  type        = string
  description = "S3 bucket name for package artifacts"
}

variable "packages_fqdn" {
  type        = string
  description = "FQDN for CloudFront (e.g. packages.hivemind.rithul.dev)"
}

variable "acm_certificate_arn" {
  type        = string
  description = "ACM certificate ARN (us-east-1, for CloudFront)"
}

variable "name_prefix" {
  type        = string
  default     = "hivemind-registry"
  description = "Prefix for resource names"
}
