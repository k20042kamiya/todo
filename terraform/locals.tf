locals {
  name_prefix = "todo-app"
  api_domain  = "api.${var.domain_name}"
  db_host     = var.create_rds ? aws_db_instance.main[0].address : ""

  common_tags = {
    Project   = "todo-app"
    ManagedBy = "terraform"
  }

  ecr_lifecycle_policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Delete untagged images after 1 day"
        selection = {
          tagStatus   = "untagged"
          countType   = "sinceImagePushed"
          countUnit   = "days"
          countNumber = 1
        }
        action = { type = "expire" }
      },
      {
        rulePriority = 2
        description  = "Keep last 10 images"
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = 10
        }
        action = { type = "expire" }
      }
    ]
  })
}
