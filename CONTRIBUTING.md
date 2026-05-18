# Contributing

## Development Setup

### Requirements

- Go 1.24+
- Docker (for acceptance tests)
- `golangci-lint` (`brew install golangci-lint`)
- `tfplugindocs` (`go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest`)

### Build

```bash
make build    # Build provider binary
make install  # Install to OpenTofu dev_override path
```

### Testing

#### Unit tests

```bash
make test
```

#### Acceptance tests (against real Kanidm)

Start a local Kanidm instance in Docker and run acceptance tests:

```bash
make kanidm-up      # Start Kanidm + initialize service account
source .env.test    # Load KANIDM_URL and KANIDM_TOKEN
make testacc        # Run acceptance tests

make kanidm-down    # Clean up
```

Alternatively, run tests inside Docker (closer to production):

```bash
make testacc-docker
```

### Code Quality

```bash
make fmt   # Format code
make lint  # Run golangci-lint
make docs  # Regenerate docs/resources/*.md via tfplugindocs
```

## Architecture

Two-layer structure:

- **`internal/client/`** — Kanidm REST API client. HTTP-level, no Terraform types.
- **`internal/provider/`** — OpenTofu resources and data sources using Terraform Plugin Framework.

All resources embed `resourceWithClient` (from `helpers.go`) to share `Configure()`.

See `.claude/architecture.md` for more detail.

## Commit Format

This project uses [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(resource): add OAuth2 public client resource
fix(client): handle 404 in group delete
docs(guide): add Unix extension example
chore(deps): upgrade terraform-plugin-framework
```

Scopes: `resource`, `client`, `provider`, `docs`, `deps`, `build`

## Pull Request Process

1. Fork and create a feature branch
2. Write tests for new behavior
3. Run `make test` and `make lint`
4. Open a PR with a clear description of the change
