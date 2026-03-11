variable "zone_id" {
  type        = string
  description = "Route53 hosted zone ID"
}

variable "domain" {
  type        = string
  description = "Root domain (e.g. hivemind.rithul.dev)"
}

variable "registry_fqdn" {
  type        = string
  description = "Registry app FQDN"
}

variable "packages_fqdn" {
  type        = string
  description = "Packages/CloudFront FQDN"
}

variable "registry_ip" {
  type        = string
  description = "EC2 Elastic IP for registry A record"
}

variable "cloudfront_domain_name" {
  type        = string
  description = "CloudFront distribution domain name (for packages CNAME/alias)"
}

variable "cloudfront_hosted_zone_id" {
  type        = string
  description = "CloudFront distribution hosted zone ID"
}

variable "ses_dkim_tokens" {
  type        = list(string)
  default     = []
  description = "SES DKIM tokens for CNAME records"
}

variable "dmarc_rua" {
  type        = string
  default     = "mailto:dmarc@hivemind.rithul.dev"
  description = "DMARC RUA address"
}
