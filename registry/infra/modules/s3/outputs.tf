output "bucket_id" {
  value = aws_s3_bucket.packages.id
}

output "bucket_arn" {
  value = aws_s3_bucket.packages.arn
}

output "cloudfront_distribution_id" {
  value = aws_cloudfront_distribution.packages.id
}

output "cloudfront_distribution_arn" {
  value = aws_cloudfront_distribution.packages.arn
}

output "cloudfront_domain_name" {
  value = aws_cloudfront_distribution.packages.domain_name
}

output "cloudfront_hosted_zone_id" {
  value = aws_cloudfront_distribution.packages.hosted_zone_id
}
