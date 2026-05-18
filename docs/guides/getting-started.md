---
page_title: "Getting Started"
subcategory: ""
description: |-
  How to configure the Kanidm provider and manage identity resources with OpenTofu.
---

# Getting Started

## Provider Configuration

### Via provider block

```hcl
provider "kanidm" {
  url   = "https://idm.example.com"
  token = var.kanidm_token
}
```

### Via environment variables (recommended)

```bash
export KANIDM_URL="https://idm.example.com"
export KANIDM_TOKEN="your-service-account-token"
```

### Creating the service account token

The provider authenticates as a Kanidm service account. Create one with the necessary privileges:

```bash
# Create service account
kanidm service-account create tofu-provider "OpenTofu Provider" \
  -H https://idm.example.com -D idm_admin

# Add to required privilege groups
kanidm group add-members idm_people_admins tofu-provider -H https://idm.example.com -D idm_admin
kanidm group add-members idm_group_admins  tofu-provider -H https://idm.example.com -D idm_admin

# Generate API token
kanidm service-account api-token generate tofu-provider provider-token \
  -H https://idm.example.com -D idm_admin
```

## Managing Person Accounts

### Passkey / modern authentication (recommended)

Generate a one-time credential reset token. The user visits the Kanidm web UI to set up passkeys or a password.

```hcl
resource "kanidm_person" "alice" {
  id                              = "alice"
  displayname                     = "Alice Smith"
  mail                            = ["alice@example.com"]
  generate_credential_reset_token = true
  credential_reset_token_ttl      = 7200  # seconds, default 3600
}

output "alice_setup_token" {
  value     = kanidm_person.alice.credential_reset_token
  sensitive = true
}
```

### Password-based authentication

```hcl
resource "kanidm_person" "bob" {
  id          = "bob"
  displayname = "Bob Jones"
  password    = var.bob_password

  lifecycle {
    ignore_changes = [password]
  }
}
```

### Unix account integration

Enable Linux/PAM authentication by setting `unix_gid` and optionally `unix_shell`:

```hcl
resource "kanidm_person" "charlie" {
  id          = "charlie"
  displayname = "Charlie Brown"
  unix_gid    = 1001
  unix_shell  = "/bin/bash"
}
```

### Additional attributes

```hcl
resource "kanidm_person" "diana" {
  id          = "diana"
  displayname = "Diana Prince"
  legalname   = "Diana Prince"
  mail        = ["diana@example.com"]
}
```

## Managing Groups

```hcl
resource "kanidm_group" "developers" {
  id          = "developers"
  description = "Development team"

  members = [
    kanidm_person.alice.id,
    kanidm_person.bob.id,
  ]
}

# Unix group
resource "kanidm_group" "linux_users" {
  id      = "linux-users"
  unix_gid = 2000
  members = [kanidm_person.charlie.id]
}
```

## Managing Service Accounts

```hcl
resource "kanidm_service_account" "ci_runner" {
  id               = "ci-runner"
  displayname      = "CI Runner"
  entry_managed_by = ["idm_admins"]
}

output "ci_runner_token" {
  value     = kanidm_service_account.ci_runner.api_token
  sensitive = true
}
```

> **Note:** The API token is only available immediately after creation. Store it securely.

## OAuth2 / OIDC Integration

Configure an OAuth2 confidential client for an application:

```hcl
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
  value     = kanidm_oauth2_basic.grafana.client_secret
  sensitive = true
}
```

> **Note:** `client_secret` is only available at creation time. Store it immediately.

### Legacy client compatibility

For clients that don't support modern standards:

```hcl
resource "kanidm_oauth2_basic" "legacy_app" {
  name                            = "legacy-app"
  displayname                     = "Legacy Application"
  origin                          = "https://app.example.com"
  allow_insecure_client_disable_pkce = true   # for clients without PKCE support
  jwt_legacy_crypto_enable           = true   # RS256 instead of ES256
  prefer_short_username              = true   # "alice" instead of "alice@idm.example.com"

  redirect_uris = ["https://app.example.com/callback"]
}
```

## Importing Existing Resources

```bash
tofu import kanidm_person.alice alice
tofu import kanidm_group.developers developers
tofu import kanidm_service_account.ci_runner ci-runner
tofu import kanidm_oauth2_basic.grafana grafana
```

## Reading Existing Resources (Data Sources)

```hcl
data "kanidm_group" "admins" {
  id = "admins"
}

resource "kanidm_oauth2_basic" "myapp" {
  name   = "myapp"
  # ...
  scope_map {
    group  = data.kanidm_group.admins.id
    scopes = ["openid", "profile", "email"]
  }
}
```
