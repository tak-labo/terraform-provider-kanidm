# Architecture

## References

- Kanidm source: https://github.com/kanidm/kanidm
- Kanidm API docs: https://kanidm.github.io/kanidm/stable/

## Two-Layer Structure

1. **API Client Layer** (`internal/client/`)
   - `client.go` - HTTP client, auth, error handling, common types
   - `person.go`, `service_account.go`, `group.go`, `oauth2.go` - Resource-specific API operations
   - All functions accept `context.Context` as first parameter
   - Returns typed errors: `ErrNotFound`, `ErrUnauthorized`, `ErrForbidden`

2. **OpenTofu Resource Layer** (`internal/provider/`)
   - `provider.go` - Provider configuration, creates API client from `KANIDM_URL`/`KANIDM_TOKEN`
   - `*_resource.go` - CRUD operations using the client layer
   - `helpers.go` - Shared `resourceWithClient`, `boolPtrFromTypes`, `scopeMapsToGroupMap`
   - Each resource has a model struct with `tfsdk` tags

## Kanidm API Patterns

- **Attribute Format**: All attributes are arrays. Use `Entry.GetString()` (single) / `Entry.GetStringSlice()` (multi)
- **FQN Members**: Group members are returned as `name@domain` — stripped to plain name on read
- **Shared Namespace**: Person, service account, and OAuth2 clients share the same name namespace
- **Sensitive Data**: OAuth2 secrets require `GET /v1/oauth2/{name}/_basic_secret` after creation; service account tokens only available at creation

## Known Gotchas

1. **OAuth2 Attribute Mapping**: `oauth2_rs_origin` (multi) = redirect URIs; `oauth2_rs_origin_landing` (single) = portal URL. Root URLs must include trailing `/`
2. **OAuth2 Secrets**: Retrieved via separate GET after creation
3. **OAuth2 Type Detection**: Check attribute key presence, not value (value is "hidden")
4. **Group Members**: Managed as a complete set — changes replace all members
5. **Empty Collections**: Return `[]string{}`, not `nil`, to avoid null vs empty drift
6. **Service Account DisplayName**: Not supported by Kanidm API (persons only)

## Resource Implementation Pattern

1. Define model struct with `tfsdk` tags
2. Implement `Metadata()`, `Schema()`, `Configure()` — embed `resourceWithClient` for Configure
3. Implement CRUD: `Create()`, `Read()`, `Update()`, `Delete()`
4. Implement `ImportState()` for `tofu import` support
5. Use `RequiresReplace()` plan modifier for immutable fields
