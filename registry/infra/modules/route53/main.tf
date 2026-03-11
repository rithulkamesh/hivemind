resource "aws_route53_record" "registry" {
  zone_id = var.zone_id
  name    = var.registry_fqdn
  type    = "A"
  ttl     = 300
  records = [var.registry_ip]
}

resource "aws_route53_record" "packages" {
  zone_id = var.zone_id
  name    = var.packages_fqdn
  type    = "A"

  alias {
    name                   = var.cloudfront_domain_name
    zone_id                = var.cloudfront_hosted_zone_id
    evaluate_target_health = false
  }
}

# SES DKIM (3 CNAME records)
resource "aws_route53_record" "ses_dkim" {
  for_each = toset(var.ses_dkim_tokens)

  zone_id = var.zone_id
  name    = "${each.value}._domainkey.${var.domain}"
  type    = "CNAME"
  ttl     = 600
  records = ["${each.value}.dkim.amazonses.com"]
}

# SPF
resource "aws_route53_record" "spf" {
  zone_id = var.zone_id
  name    = var.domain
  type    = "TXT"
  ttl     = 600
  records = ["v=spf1 include:amazonses.com ~all"]
}

# DMARC
resource "aws_route53_record" "dmarc" {
  zone_id = var.zone_id
  name    = "_dmarc.${var.domain}"
  type    = "TXT"
  ttl     = 600
  records = ["v=DMARC1; p=quarantine; rua=${var.dmarc_rua}"]
}
