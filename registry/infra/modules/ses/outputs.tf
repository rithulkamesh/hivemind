output "domain_identity_arn" {
  value = aws_ses_domain_identity.main.arn
}

output "domain_verification_token" {
  value = aws_ses_domain_identity.main.verification_token
}

output "dkim_tokens" {
  value     = aws_ses_domain_dkim.main.dkim_tokens
  sensitive = false
}

output "config_set_name" {
  value = aws_ses_configuration_set.main.name
}
