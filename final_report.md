# BetKZ Project Report: Incident Response Simulation and Infrastructure as Code Implementation

## Assignment 4: Incident Response Simulation

### Incident Summary
A production-style failure was simulated by introducing an incorrect backend database configuration in the Docker Compose deployment. The backend service failed to connect to PostgreSQL, causing API outages for betting, user login, and admin operations.

### Impact Assessment
- **Affected Services**: Backend API, user authentication, betting operations, admin functions.
- **User Impact**: Complete API unavailability, frontend errors on any data-dependent actions.
- **Business Impact**: Loss of revenue from betting operations, user frustration.

### Severity Classification
- **Severity**: High (Medium in controlled environment)
- **Rationale**: Full API outage affecting core functionality, but contained to development environment.

### Timeline of Events
1. **T-0**: System started normally with all services healthy.
2. **T+1 min**: Backend configuration changed to invalid `DB_HOST` value ("invalid-db-host").
3. **T+2 min**: Backend health check `/health/db` began failing with connection errors.
4. **T+3 min**: Frontend API calls returned 500 errors, WebSocket connections failed.
5. **T+5 min**: Prometheus metrics showed increased error rates and reduced request throughput.
6. **T+7 min**: Configuration corrected, backend restarted.
7. **T+8 min**: System functionality restored, all health checks passing.

### Root Cause Analysis
The root cause was an invalid `DB_HOST` environment variable in the backend service configuration. This caused the PostgreSQL connection to fail at startup and during runtime, preventing the backend from serving requests.

**Contributing Factors**:
- Lack of configuration validation at startup.
- No immediate alerting on database connection failures.
- Environment variable typos in deployment configuration.

### Mitigation Steps
1. **Detection**: Monitor `/health/db` endpoint and Prometheus metrics for database connection errors.
2. **Investigation**: Check backend logs with `docker logs betkz-backend` for connection errors.
3. **Fix**: Correct `DB_HOST` to "db" in `docker-compose.yml`.
4. **Recovery**: Restart backend service with `docker-compose restart backend`.
5. **Verification**: Confirm health endpoints return success.

### Resolution Confirmation
- Backend `/health` and `/health/db` endpoints return `{"status": "ok"}`.
- Frontend API calls (login, events) work successfully.
- Grafana dashboard shows restored request rates and zero error rates.
- WebSocket connections established successfully.

### Supporting Evidence

#### Screenshot 1: System Before Incident - Healthy Services
![Healthy Services](screenshots/healthy_containers.png)
*All containers running normally before incident.*

#### Screenshot 2: Incident Detection - Backend Health Failure
![Backend Health Failure](screenshots/backend_health_failure.png)
*Backend `/health/db` endpoint showing connection error.*

#### Screenshot 3: Prometheus Metrics During Incident
![Prometheus During Incident](screenshots/prometheus_during_incident.png)
*Prometheus showing increased error rates and failed requests.*

#### Screenshot 4: Grafana Dashboard During Incident
![Grafana During Incident](screenshots/grafana_during_incident.png)
*Grafana dashboard displaying service degradation.*

#### Screenshot 5: System After Resolution - Restored Services
![Restored Services](screenshots/restored_containers.png)
*All containers running normally after resolution.*

## Assignment 5: Infrastructure as Code Implementation

### Terraform Configuration Overview
The Terraform configuration provisions AWS infrastructure for the BetKZ application, including an EC2 instance with Docker and Docker Compose pre-installed, along with security groups for required ports.

### Files Structure
- `main.tf`: Core infrastructure resources (EC2 instance, security group, AMI lookup)
- `variables.tf`: Input variables for customization
- `outputs.tf`: Output values (public IP, security group ID)
- `terraform.tfvars`: Default variable values

### Provisioned Resources
1. **AWS Security Group**: Allows inbound traffic on ports 22 (SSH), 80 (HTTP), 3000 (Grafana), 9090 (Prometheus)
2. **AWS EC2 Instance**: Amazon Linux 2 instance with Docker and Docker Compose installed via user_data script
3. **AMI Lookup**: Automatically selects the latest Amazon Linux 2 AMI

### Key Configuration Details

#### main.tf
```hcl
terraform {
  required_version = ">= 1.5.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

resource "aws_security_group" "betkz" {
  name        = "${var.project_name}-sg"
  description = "Allow HTTP, Grafana, Prometheus and SSH access for BetKZ instance."

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 3000
    to_port     = 3000
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 9090
    to_port     = 9090
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project_name}-sg"
  }
}

resource "aws_instance" "betkz" {
  ami                         = aws_ami.amazon_linux.id
  instance_type               = var.instance_type
  associate_public_ip_address = true
  vpc_security_group_ids      = [aws_security_group.betkz.id]
  key_name                    = var.ssh_key_name != "" ? var.ssh_key_name : null

  user_data = <<-EOF
    #!/bin/bash
    yum update -y
    amazon-linux-extras install docker -y
    service docker start
    usermod -a -G docker ec2-user
    curl -L "https://github.com/docker/compose/releases/download/v2.20.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
  EOF

  tags = {
    Name = "${var.project_name}-instance"
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }
}
```

#### variables.tf
```hcl
variable "aws_region" {
  description = "AWS region for resources"
  type        = string
  default     = "us-east-1"
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t2.micro"
}

variable "project_name" {
  description = "Project name for resource naming"
  type        = string
  default     = "betkz"
}

variable "ssh_key_name" {
  description = "SSH key pair name for EC2 access"
  type        = string
  default     = ""
}
```

#### outputs.tf
```hcl
output "instance_public_ip" {
  description = "Public IP address of the EC2 instance"
  value       = aws_instance.betkz.public_ip
}

output "security_group_id" {
  description = "ID of the security group"
  value       = aws_security_group.betkz.id
}

output "instance_id" {
  description = "ID of the EC2 instance"
  value       = aws_instance.betkz.id
}
```

#### terraform.tfvars
```hcl
aws_region    = "us-east-1"
instance_type = "t2.micro"
project_name  = "betkz"
ssh_key_name  = ""
```

### Deployment Instructions
1. **Prerequisites**: AWS CLI configured with credentials, Terraform >= 1.5.0 installed
2. **Initialize**: `terraform init`
3. **Plan**: `terraform plan`
4. **Apply**: `terraform apply`
5. **Access**: Use output `instance_public_ip` to SSH and deploy application

### Benefits of This IaC Implementation
- **Reproducibility**: Exact infrastructure can be recreated anytime
- **Version Control**: Infrastructure changes tracked in Git
- **Automation**: No manual VM setup required
- **Consistency**: Same configuration across environments
- **Security**: Declarative security group rules

### Supporting Evidence

#### Screenshot 6: Terraform Plan Output
![Terraform Plan](screenshots/terraform_plan.png)
*Terraform plan showing resources to be created.*

#### Screenshot 7: Terraform Apply Output
![Terraform Apply](screenshots/terraform_apply.png)
*Terraform apply successfully provisioning resources.*

#### Screenshot 8: AWS Console - EC2 Instance
![EC2 Instance](screenshots/aws_ec2_instance.png)
*Provisioned EC2 instance in AWS console.*

#### Screenshot 9: AWS Console - Security Group
![Security Group](screenshots/aws_security_group.png)
*Security group with required ports open.*

#### Screenshot 10: SSH Access to Instance
![SSH Access](screenshots/ssh_instance.png)
*SSH connection to provisioned instance showing Docker ready.*

## Postmortem Analysis

### Incident Overview
The simulated incident involved a database connectivity failure in the backend service, causing complete API unavailability.

### Customer Impact
- Users unable to login, place bets, or access account information
- Frontend displayed error messages for all API-dependent features
- WebSocket connections for real-time updates failed

### Root Cause Analysis
Primary cause: Configuration error in environment variables
- Invalid `DB_HOST` setting prevented PostgreSQL connection
- No validation of database connectivity at startup
- Lack of immediate alerting mechanisms

### Detection and Response Evaluation
**Strengths**:
- Health endpoints provided quick status checks
- Prometheus metrics captured failure patterns
- Docker logs contained detailed error information

**Weaknesses**:
- No automated alerting on health endpoint failures
- Manual monitoring required for incident detection
- No circuit breakers or graceful degradation

### Resolution Summary
Configuration correction and service restart restored full functionality within minutes. All health checks passed and user operations resumed normally.

### Lessons Learned
1. **Configuration Management**: Implement configuration validation at startup
2. **Monitoring**: Add alerting rules for health endpoint failures
3. **Resilience**: Consider database connection pooling and retry logic
4. **Testing**: Include configuration testing in deployment pipeline

### Action Items
1. **High Priority**: Add startup configuration validation in backend
2. **High Priority**: Implement Prometheus alerting rules for service health
3. **Medium Priority**: Add database connection health checks with retries
4. **Medium Priority**: Implement circuit breaker pattern for database calls
5. **Low Priority**: Add automated chaos engineering tests for configuration failures

## Conclusion

This project successfully demonstrated the integration of containerized microservices with Infrastructure as Code and incident response practices. The Terraform configuration provides reproducible infrastructure provisioning, while the incident simulation validated monitoring and response capabilities. The comprehensive postmortem analysis identifies areas for improvement to enhance system reliability and operational excellence.

All requirements from the assignment specifications have been fulfilled, including containerized deployment, monitoring integration, incident simulation, and IaC implementation.