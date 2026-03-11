data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# Optional: lookup Route53 zone by domain name if zone_id not provided
data "aws_route53_zone" "main" {
  count        = var.route53_zone_id != null ? 0 : 1
  name         = "${var.domain}."
  private_zone = false
}

# ACM certificate (must exist in us-east-1 for CloudFront)
data "aws_acm_certificate" "cloudfront" {
  count       = var.acm_certificate_arn == null ? 1 : 0
  provider    = aws.us_east_1
  domain      = var.domain
  statuses    = ["ISSUED"]
  most_recent = true
}

locals {
  route53_zone_id = var.route53_zone_id != null ? var.route53_zone_id : data.aws_route53_zone.main[0].zone_id
  acm_cert_arn    = var.acm_certificate_arn != null ? var.acm_certificate_arn : data.aws_acm_certificate.cloudfront[0].arn
}

# Provider alias for ACM/CloudFront (cert must be in us-east-1)
provider "aws" {
  alias  = "us_east_1"
  region = "us-east-1"
  default_tags {
    tags = { Project = "hivemind-registry" }
  }
}

# Networking
module "networking" {
  source = "./modules/networking"

  vpc_cidr             = "10.0.0.0/16"
  public_subnet_cidr   = "10.0.1.0/24"
  availability_zone    = "${var.aws_region}a"
  ssh_allowed_cidrs    = var.ssh_allowed_cidrs
  name_prefix          = "hivemind-registry"
}

# ECR (no dependency on S3/CloudFront)
module "ecr" {
  source = "./modules/ecr"
  repository_name = var.ecr_api_repository_name
  name_prefix     = "hivemind-registry"
}

# S3 + CloudFront (needs ACM cert in us-east-1)
module "s3" {
  source = "./modules/s3"

  bucket_name        = var.s3_packages_bucket
  packages_fqdn      = var.packages_fqdn
  acm_certificate_arn = local.acm_cert_arn
  name_prefix        = "hivemind-registry"
}

# SES
module "ses" {
  source = "./modules/ses"
  domain         = var.domain
  config_set_name = var.ses_config_set
  dmarc_rua      = "mailto:dmarc@${var.domain}"
}

# IAM (needs ECR and S3/CloudFront ARNs)
module "iam" {
  source = "./modules/iam"

  name_prefix                 = "hivemind-registry"
  s3_packages_bucket_arn      = module.s3.bucket_arn
  cloudfront_distribution_arn = module.s3.cloudfront_distribution_arn
  ecr_repository_arn          = module.ecr.repository_arn
  github_org_repo             = var.github_org_repo
  github_branch               = var.github_branch
  terraform_state_bucket_arn  = "arn:aws:s3:::${var.terraform_state_bucket_name}"
  terraform_lock_table_arn     = "arn:aws:dynamodb:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:table/${var.terraform_lock_table_name}"
}

# EC2 (needs networking, IAM, optional key)
module "ec2" {
  source = "./modules/ec2"

  name_prefix              = "hivemind-registry"
  subnet_id                = module.networking.public_subnet_id
  security_group_ids        = [module.networking.security_group_id]
  iam_instance_profile_name = module.iam.ec2_instance_profile_name
  key_pair_name            = var.key_pair_name
  create_key_pair          = var.create_key_pair
  repo_clone_url           = "https://github.com/${var.github_org_repo}.git"
  repo_branch              = "main"
}

# Route53 records (needs EC2 IP, CloudFront, SES DKIM)
module "route53" {
  source = "./modules/route53"

  zone_id                  = local.route53_zone_id
  domain                   = var.domain
  registry_fqdn            = var.registry_fqdn
  packages_fqdn            = var.packages_fqdn
  registry_ip              = module.ec2.public_ip
  cloudfront_domain_name   = module.s3.cloudfront_domain_name
  cloudfront_hosted_zone_id = module.s3.cloudfront_hosted_zone_id
  ses_dkim_tokens          = module.ses.dkim_tokens
  dmarc_rua                = "mailto:dmarc@${var.domain}"
}
