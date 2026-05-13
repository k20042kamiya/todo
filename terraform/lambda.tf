# 初回apply用のプレースホルダー。実際のデプロイはCI/CDでaws lambda update-function-codeを実行
data "archive_file" "lambda_placeholder" {
  type        = "zip"
  output_path = "${path.module}/lambda_placeholder.zip"
  source {
    content  = "placeholder"
    filename = "bootstrap"
  }
}

resource "aws_lambda_function" "notification" {
  function_name    = "${local.name_prefix}-notification"
  role             = aws_iam_role.lambda.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  filename         = data.archive_file.lambda_placeholder.output_path
  source_code_hash = data.archive_file.lambda_placeholder.output_base64sha256
  timeout          = 300
  memory_size      = 128

  vpc_config {
    subnet_ids         = aws_subnet.private[*].id
    security_group_ids = [aws_security_group.lambda.id]
  }

  environment {
    variables = {
      DB_HOST          = aws_db_instance.main.address
      DB_PORT          = "3306"
      DB_NAME          = var.db_name
      DB_USER          = var.db_username
      DB_PASSWORD_PATH = aws_ssm_parameter.db_password.name
      SES_SENDER_EMAIL = var.ses_sender_email
      AWS_REGION_NAME  = var.aws_region
    }
  }

  lifecycle {
    ignore_changes = [filename, source_code_hash]
  }

  tags = local.common_tags
}

resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/${aws_lambda_function.notification.function_name}"
  retention_in_days = 30
  tags              = local.common_tags
}

# EventBridge: 毎日09:00 JST (= 00:00 UTC)
resource "aws_cloudwatch_event_rule" "daily_notification" {
  name                = "${local.name_prefix}-daily-notification"
  schedule_expression = "cron(0 0 * * ? *)"
  tags                = local.common_tags
}

resource "aws_cloudwatch_event_target" "notification_lambda" {
  rule      = aws_cloudwatch_event_rule.daily_notification.name
  target_id = "NotificationLambda"
  arn       = aws_lambda_function.notification.arn
}

resource "aws_lambda_permission" "allow_eventbridge" {
  statement_id  = "AllowExecutionFromEventBridge"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.notification.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.daily_notification.arn
}
