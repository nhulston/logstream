# Security group for LogStream EC2 instance
resource "aws_security_group" "logstream" {
  name        = "${var.project_name}-${var.environment}-sg"
  description = "Security group for LogStream distributed log ingestion system"

  tags = {
    Name = "${var.project_name}-${var.environment}-sg"
  }
}

# SSH access
resource "aws_vpc_security_group_ingress_rule" "ssh" {
  security_group_id = aws_security_group.logstream.id
  description       = "SSH access"

  cidr_ipv4   = var.allowed_ssh_cidr
  from_port   = 22
  to_port     = 22
  ip_protocol = "tcp"

  tags = {
    Name = "ssh-access"
  }
}

# gRPC ingestion API
resource "aws_vpc_security_group_ingress_rule" "grpc" {
  security_group_id = aws_security_group.logstream.id
  description       = "gRPC log ingestion API"

  cidr_ipv4   = "0.0.0.0/0"
  from_port   = 50051
  to_port     = 50051
  ip_protocol = "tcp"

  tags = {
    Name = "grpc-ingestion"
  }
}

# HTTP Query API
resource "aws_vpc_security_group_ingress_rule" "http_query" {
  security_group_id = aws_security_group.logstream.id
  description       = "HTTP query API"

  cidr_ipv4   = "0.0.0.0/0"
  from_port   = 8080
  to_port     = 8080
  ip_protocol = "tcp"

  tags = {
    Name = "http-query-api"
  }
}

# Allow all outbound traffic
resource "aws_vpc_security_group_egress_rule" "all_outbound" {
  security_group_id = aws_security_group.logstream.id
  description       = "Allow all outbound traffic"

  cidr_ipv4   = "0.0.0.0/0"
  ip_protocol = "-1"

  tags = {
    Name = "all-outbound"
  }
}
