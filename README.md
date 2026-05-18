# terraform-provider-kanidm

OpenTofu provider for [Kanidm](https://github.com/kanidm/kanidm) — a modern, fast identity management platform.

[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)
[![Test](https://github.com/tak-labo/terraform-provider-kanidm/actions/workflows/test.yml/badge.svg)](https://github.com/tak-labo/terraform-provider-kanidm/actions/workflows/test.yml)
[![API Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/tak-55/13d21aea23677d03faba08bca92bf843/raw/badge.json)](https://github.com/tak-labo/terraform-provider-kanidm/actions/workflows/test.yml)

Manage Kanidm identity resources as code: person accounts, service accounts, groups, and OAuth2 clients.

## Quick Start

```hcl
terraform {
  required_providers {
    kanidm = {
      source  = "tak-labo/kanidm"
      version = "~> 0.1"
    }
  }
}

provider "kanidm" {
  url   = "https://idm.example.com"
  token = var.kanidm_token
}

resource "kanidm_person" "alice" {
  id                              = "alice"
  displayname                     = "Alice Smith"
  mail                            = ["alice@example.com"]
  generate_credential_reset_token = true
}

resource "kanidm_group" "developers" {
  id      = "developers"
  members = [kanidm_person.alice.id]
}
```

## Resources

| Resource | Description |
|---|---|
| `kanidm_person` | Person account with passkey/password auth, Unix extension, legalname |
| `kanidm_service_account` | Service account with API token |
| `kanidm_group` | Group with member management and Unix extension |
| `kanidm_oauth2_basic` | OAuth2 confidential client with scope mapping |

## Data Sources

| Data Source | Description |
|---|---|
| `data.kanidm_person` | Read existing person account |
| `data.kanidm_service_account` | Read existing service account |
| `data.kanidm_group` | Read existing group |
| `data.kanidm_oauth2_basic` | Read existing OAuth2 client |

## Documentation

- [Getting Started](docs/guides/getting-started.md)
- [Resource Reference](docs/resources/)
- [Data Source Reference](docs/data-sources/)

## Requirements

- [OpenTofu](https://opentofu.org/docs/intro/install/) >= 1.6
- Kanidm >= 1.9 (developed and tested against **1.9.2**)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and contribution guidelines.

## License

[Mozilla Public License 2.0](LICENSE)

## Acknowledgments

Inspired by [ssoriche/terraform-provider-kanidm](https://github.com/ssoriche/terraform-provider-kanidm).
