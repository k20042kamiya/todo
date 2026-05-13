locals {
  name_prefix = "todo-app"
  api_domain  = "api.${var.domain_name}"

  common_tags = {
    Project   = "todo-app"
    ManagedBy = "terraform"
  }
}
