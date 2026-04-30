output "instance_public_ip" {
  description = "Public IP address of the provisioned BetKZ EC2 instance."
  value       = aws_instance.betkz.public_ip
}

output "security_group_id" {
  description = "Security group ID for the BetKZ instance."
  value       = aws_security_group.betkz.id
}
