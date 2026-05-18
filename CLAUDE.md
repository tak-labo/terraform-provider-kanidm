# CLAUDE.md

OpenTofu provider for Kanidm. Built with Terraform Plugin Framework v1.17.0 and Go 1.24.

## Commands

```bash
make build    # Build provider binary
make install  # Install to ~/.local/share/opentofu/plugins/ (dev_override)
make test     # Unit tests
make testacc  # Acceptance tests (requires KANIDM_URL and KANIDM_TOKEN)
make fmt      # Format code
make lint     # golangci-lint
make docs     # Regenerate docs/resources/*.md via tfplugindocs
make clean    # Remove build artifacts
```

## Provider Configuration

Via HCL or environment variables `KANIDM_URL` / `KANIDM_TOKEN`.

## Conventional Commits

Scopes: `resource`, `client`, `provider`, `docs`, `deps`, `build`

## References

@.claude/architecture.md
@.claude/workflow.md
