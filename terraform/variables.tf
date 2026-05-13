variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "ap-northeast-1"
}

variable "domain_name" {
  description = "Root domain name"
  type        = string
  default     = "todo-app.dev"
}

variable "db_name" {
  description = "Database name"
  type        = string
  default     = "todo"
}

variable "db_username" {
  description = "Database master username"
  type        = string
  default     = "admin"
}

variable "db_password" {
  description = "Database master password"
  type        = string
  sensitive   = true
}

variable "ses_sender_email" {
  description = "SES verified sender email address"
  type        = string
  default     = "noreply@todo-app.dev"
}

variable "firebase_project_id" {
  description = "Firebase project ID"
  type        = string
}
