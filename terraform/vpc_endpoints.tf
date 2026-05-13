# Interface型エンドポイント共通セキュリティグループ
resource "aws_security_group" "vpc_endpoint" {
  name   = "${local.name_prefix}-vpc-endpoint-sg"
  vpc_id = aws_vpc.main.id

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = [aws_vpc.main.cidr_block]
  }

  tags = merge(local.common_tags, { Name = "${local.name_prefix}-vpc-endpoint-sg" })
}

# S3 Gateway Endpoint: 無料。プライベートサブネット内からS3へのアクセスに使用
resource "aws_vpc_endpoint" "s3" {
  vpc_id            = aws_vpc.main.id
  service_name      = "com.amazonaws.${var.aws_region}.s3"
  vpc_endpoint_type = "Gateway"
  route_table_ids   = [aws_route_table.private.id]
  tags              = merge(local.common_tags, { Name = "${local.name_prefix}-s3-endpoint" })
}

# 以下3つはLambda (プライベートサブネット) 専用
# NATなしでAWSサービスに到達するために必要
# コスト削減のため1AZのみ (個人プロジェクトのためHA不要)

resource "aws_vpc_endpoint" "ssm" {
  vpc_id              = aws_vpc.main.id
  service_name        = "com.amazonaws.${var.aws_region}.ssm"
  vpc_endpoint_type   = "Interface"
  subnet_ids          = [aws_subnet.private[0].id]
  security_group_ids  = [aws_security_group.vpc_endpoint.id]
  private_dns_enabled = true
  tags                = merge(local.common_tags, { Name = "${local.name_prefix}-ssm-endpoint" })
}

resource "aws_vpc_endpoint" "logs" {
  vpc_id              = aws_vpc.main.id
  service_name        = "com.amazonaws.${var.aws_region}.logs"
  vpc_endpoint_type   = "Interface"
  subnet_ids          = [aws_subnet.private[0].id]
  security_group_ids  = [aws_security_group.vpc_endpoint.id]
  private_dns_enabled = true
  tags                = merge(local.common_tags, { Name = "${local.name_prefix}-logs-endpoint" })
}

# SES: ap-northeast-1でエンドポイントが利用可能か terraform plan で確認すること
# 利用不可の場合は com.amazonaws.${var.aws_region}.email-smtp (SMTP) を検討
resource "aws_vpc_endpoint" "ses" {
  vpc_id              = aws_vpc.main.id
  service_name        = "com.amazonaws.${var.aws_region}.email"
  vpc_endpoint_type   = "Interface"
  subnet_ids          = [aws_subnet.private[0].id]
  security_group_ids  = [aws_security_group.vpc_endpoint.id]
  private_dns_enabled = true
  tags                = merge(local.common_tags, { Name = "${local.name_prefix}-ses-endpoint" })
}
