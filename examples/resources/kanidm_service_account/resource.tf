# Example: Service account for automation
resource "kanidm_service_account" "iac_bot" {
  id               = "iac-bot"
  displayname      = "IaC Automation Account"
  entry_managed_by = ["idm_admins"]
}

# Store the API token securely
output "terraform_api_token" {
  description = "API token for Terraform service account"
  value       = kanidm_service_account.iac_bot.api_token
  sensitive   = true
}

# Example: Service account for CI/CD
resource "kanidm_service_account" "argocd" {
  id               = "argocd"
  displayname      = "ArgoCD Service Account"
  entry_managed_by = ["idm_admins"]
}

# Example: Service account for monitoring
resource "kanidm_service_account" "prometheus" {
  id               = "prometheus"
  displayname      = "Prometheus Monitoring"
  entry_managed_by = ["idm_admins"]
}

# Example: Imported existing service account
# Import command: tofu import kanidm_service_account.existing existing_account_id
resource "kanidm_service_account" "existing" {
  id               = "existing-service"
  displayname      = "Existing Service Account"
  entry_managed_by = ["idm_admins"]
}
