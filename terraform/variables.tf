variable "aws_region" {
  description = "AWS region to deploy resources"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "project_name" {
  description = "Project name for resource naming"
  type        = string
  default     = "logstream"
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.small" # 2GB RAM needed for Kafka + builds
}

variable "allowed_ssh_cidr" {
  description = "CIDR block allowed to SSH into the instance"
  type        = string
  default     = "0.0.0.0/0" # In practice, this should be a specific IP address for security
}

variable "key_pair_name" {
  description = "Name of the AWS key pair for SSH access"
  type        = string
  default     = "logstream-key" # Created in AWS Console
}

variable "volume_size" {
  description = "Size of the EBS volume in GB"
  type        = number
  default     = 30 # Free tier includes 30GB
}

variable "git_repo_url" {
  description = "Git repository URL to clone"
  type        = string
  default     = "https://github.com/nhulston/logstream.git"
}
