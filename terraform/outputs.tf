output "instance_id" {
  description = "ID of the EC2 instance"
  value       = aws_instance.logstream.id
}

output "instance_public_ip" {
  description = "Public IP address of the EC2 instance"
  value       = aws_eip.logstream.public_ip
}

output "instance_public_dns" {
  description = "Public DNS name of the EC2 instance"
  value       = aws_instance.logstream.public_dns
}

output "grpc_endpoint" {
  description = "gRPC endpoint for log ingestion"
  value       = "${aws_eip.logstream.public_ip}:50051"
}

output "query_api_endpoint" {
  description = "HTTP endpoint for query API"
  value       = "http://${aws_eip.logstream.public_ip}:8080"
}

output "ssh_command" {
  description = "SSH command to connect to the instance"
  value       = "ssh -i ~/.ssh/${var.key_pair_name}.pem ec2-user@${aws_eip.logstream.public_ip}"
}

output "security_group_id" {
  description = "ID of the security group"
  value       = aws_security_group.logstream.id
}
