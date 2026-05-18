terraform {
  required_providers {
    kanidm = {
      source = "tak-labo/kanidm"
    }
  }
}

# Configure the Kanidm Provider with explicit credentials
provider "kanidm" {
  url   = "https://idm.example.com"
  token = var.kanidm_token
}

# Alternative: Configure using environment variables
# export KANIDM_URL="https://idm.example.com"
# export KANIDM_TOKEN="your-api-token"
provider "kanidm" {}
