variable "name_prefix" {
  type        = string
  default     = "hivemind-registry"
  description = "Prefix for resource names"
}

variable "subnet_id" {
  type        = string
  description = "Public subnet ID"
}

variable "security_group_ids" {
  type        = list(string)
  description = "Security group IDs for the instance"
}

variable "iam_instance_profile_name" {
  type        = string
  description = "IAM instance profile name"
}

variable "key_pair_name" {
  type        = string
  default     = null
  description = "Existing key pair name. If null, use key_name from created key"
}

variable "public_key_openssh" {
  type        = string
  default     = null
  description = "Public key (OpenSSH) for created key pair; required if key_pair_name is null"
}

variable "user_data" {
  type        = string
  default     = ""
  description = "User data script for first boot"
}

variable "repo_clone_url" {
  type        = string
  default     = "https://github.com/rithulkamesh/hivemind.git"
  description = "Git clone URL for deploy files (public read)"
}

variable "repo_branch" {
  type        = string
  default     = "main"
  description = "Branch to clone for deploy"
}

variable "create_key_pair" {
  type        = bool
  default     = false
  description = "Create a key pair and output private key"
}
