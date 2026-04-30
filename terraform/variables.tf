variable "aws_region" {
  type        = string
  description = "AWS region where the BetKZ instance will be created."
  default     = "us-east-1"
}

variable "instance_type" {
  type        = string
  description = "EC2 instance type for the BetKZ VM."
  default     = "t3.micro"
}

variable "project_name" {
  type        = string
  description = "Project name used for resource tagging."
  default     = "betkz"
}

variable "ssh_key_name" {
  type        = string
  description = "Optional SSH key name to associate with the EC2 instance."
  default     = ""
}
