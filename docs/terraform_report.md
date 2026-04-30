# Terraform Implementation Report
This Terraform configuration provisions an AWS EC2 instance and network access rules for the BetKZ deployment.
## Files

- `main.tf` — AWS provider, security group, EC2 instance, and basic bootstrapping.
- `variables.tf` — region, instance type, project name, and optional SSH key.
- `outputs.tf` — public IP and security group ID.
- `terraform.tfvars` — sample values for local use.

## Provisioned resources
- EC2 instance with Amazon Linux 2.
- Security group allowing:
  - SSH: 22
  - HTTP: 80
  - Grafana: 3000
  - Prometheus: 9090
- Public IP output for the provisioned VM.

## Usage
1. Install Terraform.
2. Configure AWS credentials in environment variables:
   ```bash
   export AWS_ACCESS_KEY_ID=...
   export AWS_SECRET_ACCESS_KEY=...
   export AWS_DEFAULT_REGION=us-east-1
   ```
3. Run:
   ```bash
   cd terraform
   terraform init
   terraform plan
   terraform apply
   ```
4. After apply, use the `instance_public_ip` output to access the deployed VM.

## Notes
- The Terraform configuration includes a simple `user_data` script to install Docker and Docker Compose on Amazon Linux.
- This is intended as infrastructure provisioning for deployment automation and reproducibility.
- **Current Status**: Terraform configuration scaffold is complete and ready for deployment. Requires AWS credentials and Terraform CLI for execution.
- **Limitations**: Does not include automated deployment of the BetKZ application code to the provisioned instance. Manual deployment steps would be needed after infrastructure provisioning.
