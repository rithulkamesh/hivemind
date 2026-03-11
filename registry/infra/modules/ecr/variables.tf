variable "repository_name" {
  type        = string
  description = "ECR repository name"
}

variable "name_prefix" {
  type        = string
  default     = "hivemind-registry"
  description = "Prefix for resource names"
}
