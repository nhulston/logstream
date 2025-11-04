# IAM role for EC2 instance
resource "aws_iam_role" "logstream" {
  name = "${var.project_name}-${var.environment}-ec2-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name = "${var.project_name}-${var.environment}-ec2-role"
  }
}

# IAM policy for CloudWatch logs (optional)
resource "aws_iam_role_policy_attachment" "cloudwatch_logs" {
  role       = aws_iam_role.logstream.name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchAgentServerPolicy"
}

# IAM instance profile
resource "aws_iam_instance_profile" "logstream" {
  name = "${var.project_name}-${var.environment}-instance-profile"
  role = aws_iam_role.logstream.name
}

# EC2 instance
resource "aws_instance" "logstream" {
  ami                    = data.aws_ami.amazon_linux_2023.id
  instance_type          = var.instance_type
  key_name               = var.key_pair_name
  vpc_security_group_ids = [aws_security_group.logstream.id]
  iam_instance_profile   = aws_iam_instance_profile.logstream.name

  root_block_device {
    volume_type           = "gp3"
    volume_size           = var.volume_size
    delete_on_termination = true
    encrypted             = true

    tags = {
      Name = "${var.project_name}-${var.environment}-root-volume"
    }
  }

  user_data = templatefile("${path.module}/user_data.sh", {
    environment = var.environment
  })

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required" # IMDSv2 for security
    http_put_response_hop_limit = 1
  }

  tags = {
    Name = "${var.project_name}-${var.environment}-instance"
  }

  lifecycle {
    ignore_changes = [ami]
  }
}

# Elastic IP for consistent public IP
resource "aws_eip" "logstream" {
  domain = "vpc"

  tags = {
    Name = "${var.project_name}-${var.environment}-eip"
  }
}

# Elastic IP association
resource "aws_eip_association" "logstream" {
  instance_id   = aws_instance.logstream.id
  allocation_id = aws_eip.logstream.id
}
