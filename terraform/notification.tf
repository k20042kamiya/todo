resource "aws_ecr_repository" "notification" {
  name                 = "${local.name_prefix}-notification"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = local.common_tags
}

resource "aws_ecs_task_definition" "notification" {
  family                   = "${local.name_prefix}-notification"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([{
    name    = "notification"
    image   = "${aws_ecr_repository.notification.repository_url}:latest"
    command = ["./notification"]
    environment = [
      { name = "DB_HOST",         value = local.db_host },
      { name = "DB_PORT",         value = "3306" },
      { name = "DB_NAME",         value = var.db_name },
      { name = "DB_USER",         value = var.db_username },
      { name = "SES_FROM_EMAIL",  value = var.ses_sender_email },
      { name = "AWS_REGION_NAME", value = var.aws_region },
    ]
    secrets = [
      { name = "DB_PASSWORD", valueFrom = aws_ssm_parameter.db_password.arn },
    ]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.notification.name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "notification"
      }
    }
  }])

  tags = local.common_tags
}

resource "aws_cloudwatch_log_group" "notification" {
  name              = "/ecs/${local.name_prefix}-notification"
  retention_in_days = 30
  tags              = local.common_tags
}

# EventBridge Scheduler: 毎日09:00 JST
resource "aws_scheduler_schedule" "daily_notification" {
  name       = "${local.name_prefix}-daily-notification"
  group_name = "default"

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression          = "cron(0 9 * * ? *)"
  schedule_expression_timezone = "Asia/Tokyo"

  target {
    arn      = aws_ecs_cluster.main.arn
    role_arn = aws_iam_role.scheduler.arn

    ecs_parameters {
      task_definition_arn = aws_ecs_task_definition.notification.arn
      launch_type         = "FARGATE"

      network_configuration {
        subnets          = aws_subnet.public[*].id
        security_groups  = [aws_security_group.notification.id]
        assign_public_ip = true
      }
    }
  }
}
