# Example: OAuth2 client for Grafana
# https://kanidm.github.io/kanidm/stable/integrations/oauth2/examples.html
resource "kanidm_oauth2_basic" "grafana" {
  name        = "grafana"
  displayname = "Grafana"
  origin      = "https://grafana.example.com"

  redirect_uris = [
    "https://grafana.example.com/login/generic_oauth",
  ]

  scope_map {
    group  = "admins"
    scopes = ["openid", "profile", "email", "groups"]
  }

  scope_map {
    group  = "developers"
    scopes = ["openid", "profile", "email"]
  }
}

output "grafana_client_secret" {
  description = "OAuth2 client secret for Grafana"
  value       = kanidm_oauth2_basic.grafana.client_secret
  sensitive   = true
}

# Example: OAuth2 client for Gitea
resource "kanidm_oauth2_basic" "gitea" {
  name        = "gitea"
  displayname = "Gitea"
  origin      = "https://gitea.example.com"

  redirect_uris = [
    "https://gitea.example.com/user/oauth2/kanidm/callback",
  ]

  scope_map {
    group  = "developers"
    scopes = ["email", "openid", "profile", "groups"]
  }
}

# Example: OAuth2 client for GitLab
resource "kanidm_oauth2_basic" "gitlab" {
  name        = "gitlab"
  displayname = "GitLab"
  origin      = "https://gitlab.example.com"

  redirect_uris = [
    "https://gitlab.example.com/users/auth/openid_connect/callback",
  ]

  scope_map {
    group  = "developers"
    scopes = ["openid", "profile", "email"]
  }
}

# Example: Simple OAuth2 client (no scope maps)
resource "kanidm_oauth2_basic" "simple_app" {
  name        = "my-app"
  displayname = "My Application"
  origin      = "https://app.example.com"

  redirect_uris = [
    "https://app.example.com/callback",
  ]
}
