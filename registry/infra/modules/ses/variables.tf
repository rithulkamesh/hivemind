variable "domain" {
  type        = string
  description = "Domain for SES identity (e.g. hivemind.rithul.dev)"
}

variable "config_set_name" {
  type        = string
  default     = "hivemind-registry"
  description = "SES configuration set name"
}

variable "dmarc_rua" {
  type        = string
  default     = "mailto:dmarc@hivemind.rithul.dev"
  description = "DMARC RUA reporting address"
}
