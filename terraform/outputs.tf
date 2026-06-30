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
  value       = var.create_rds ? aws_db_instance.main[0].address : "RDS未作成"
  sensitive   = true
}

output "notification_ecr_repository_url" {
  description = "通知バッチ用ECRリポジトリURL (CI/CDでdocker pushに使用)"
  value       = aws_ecr_repository.notification.repository_url
}
