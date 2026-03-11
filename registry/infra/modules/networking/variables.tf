variable "vpc_cidr" {
  type        = string
  default     = "10.0.0.0/16"
  description = "VPC CIDR"
}

variable "public_subnet_cidr" {
  type        = string
  default     = "10.0.1.0/24"
  description = "Public subnet CIDR"
}

variable "availability_zone" {
  type        = string
  description = "AZ for public subnet (e.g. us-east-1a)"
}

variable "ssh_allowed_cidrs" {
  type        = list(string)
  description = "CIDR blocks allowed to SSH"
}

variable "name_prefix" {
  type        = string
  default     = "hivemind-registry"
  description = "Prefix for resource names"
}
