output "route53_nameservers" {
  description = "ドメインレジストラ(Route 53)に設定するネームサーバー"
  value       = aws_route53_zone.main.name_servers
}

output "cloudfront_domain" {
  description = "CloudFrontのドメイン名"
  value       = aws_cloudfront_distribution.frontend.domain_name
}

output "alb_dns_name" {
  description = "ALBのDNS名"
  value       = aws_lb.main.dns_name
}

output "ecr_repository_url" {
  description = "ECRリポジトリURL (CI/CDでdocker pushに使用)"
  value       = aws_ecr_repository.api.repository_url
}

output "rds_endpoint" {
  description = "RDSエンドポイント"
  value       = aws_db_instance.main.address
  sensitive   = true
}

output "lambda_function_name" {
  description = "Lambda関数名 (CI/CDでupdate-function-codeに使用)"
  value       = aws_lambda_function.notification.function_name
}
