## API Coverage: 29/86 endpoints (33.7%)
> Kanidm v1.10.2 · OpenTofu Provider

### Coverage by category

| Category | Implemented | Total | % |
|---|---|---|---|
| account | 0 | 8 | ⬜⬜⬜⬜⬜ 0% |
| group | 6 | 12 | 🟩🟩⬜⬜⬜ 50% |
| oauth2 | 8 | 16 | 🟩🟩⬜⬜⬜ 50% |
| other | 0 | 1 | ⬜⬜⬜⬜⬜ 0% |
| person | 8 | 29 | 🟩⬜⬜⬜⬜ 28% |
| service_account | 7 | 20 | 🟩⬜⬜⬜⬜ 35% |

---

### account (0/8 · 0%)

| | Method | Endpoint |
|---|---|---|
| ❌ | `GET` | `/v1/account/{id}/_radius/_token` |
| ❌ | `POST` | `/v1/account/{id}/_radius/_token` |
| ❌ | `GET` | `/v1/account/{id}/_ssh_pubkeys` |
| ❌ | `GET` | `/v1/account/{id}/_ssh_pubkeys/{tag}` |
| ❌ | `POST` | `/v1/account/{id}/_unix/_auth` |
| ❌ | `POST` | `/v1/account/{id}/_unix/_token` |
| ❌ | `GET` | `/v1/account/{id}/_user_auth_token` |
| ❌ | `GET` | `/v1/account/{id}/_user_auth_token/{token_id}` |

### group (6/12 · 50%)

| | Method | Endpoint |
|---|---|---|
| ✅ | `GET` | `/v1/group` |
| ✅ | `POST` | `/v1/group` |
| ❌ | `GET` | `/v1/group/_search/{id}` |
| ✅ | `GET` | `/v1/group/{id}` |
| ✅ | `DELETE` | `/v1/group/{id}` |
| ✅ | `PATCH` | `/v1/group/{id}` |
| ❌ | `GET` | `/v1/group/{id}/_attr/{attr}` |
| ❌ | `PUT` | `/v1/group/{id}/_attr/{attr}` |
| ❌ | `POST` | `/v1/group/{id}/_attr/{attr}` |
| ❌ | `DELETE` | `/v1/group/{id}/_attr/{attr}` |
| ✅ | `POST` | `/v1/group/{id}/_unix` |
| ❌ | `GET` | `/v1/group/{id}/_unix/_token` |

### oauth2 (8/16 · 50%)

| | Method | Endpoint |
|---|---|---|
| ❌ | `GET` | `/v1/oauth2` |
| ✅ | `POST` | `/v1/oauth2/_basic` |
| ✅ | `POST` | `/v1/oauth2/_public` |
| ✅ | `GET` | `/v1/oauth2/{rs_name}` |
| ✅ | `DELETE` | `/v1/oauth2/{rs_name}` |
| ✅ | `PATCH` | `/v1/oauth2/{rs_name}` |
| ✅ | `GET` | `/v1/oauth2/{rs_name}/_basic_secret` |
| ❌ | `POST` | `/v1/oauth2/{rs_name}/_claimmap/{claim_name}` |
| ❌ | `POST` | `/v1/oauth2/{rs_name}/_claimmap/{claim_name}/{group}` |
| ❌ | `DELETE` | `/v1/oauth2/{rs_name}/_claimmap/{claim_name}/{group}` |
| ❌ | `POST` | `/v1/oauth2/{rs_name}/_image` |
| ❌ | `DELETE` | `/v1/oauth2/{rs_name}/_image` |
| ✅ | `POST` | `/v1/oauth2/{rs_name}/_scopemap/{group}` |
| ✅ | `DELETE` | `/v1/oauth2/{rs_name}/_scopemap/{group}` |
| ❌ | `POST` | `/v1/oauth2/{rs_name}/_sup_scopemap/{group}` |
| ❌ | `DELETE` | `/v1/oauth2/{rs_name}/_sup_scopemap/{group}` |

### other (0/1 · 0%)

| | Method | Endpoint |
|---|---|---|
| ❌ | `GET` | `/status` |

### person (8/29 · 28%)

| | Method | Endpoint |
|---|---|---|
| ✅ | `GET` | `/v1/person` |
| ✅ | `POST` | `/v1/person` |
| ❌ | `GET` | `/v1/person/_search/{id}` |
| ✅ | `GET` | `/v1/person/{id}` |
| ✅ | `DELETE` | `/v1/person/{id}` |
| ✅ | `PATCH` | `/v1/person/{id}` |
| ❌ | `GET` | `/v1/person/{id}/_attr/{attr}` |
| ❌ | `PUT` | `/v1/person/{id}/_attr/{attr}` |
| ❌ | `POST` | `/v1/person/{id}/_attr/{attr}` |
| ❌ | `DELETE` | `/v1/person/{id}/_attr/{attr}` |
| ❌ | `GET` | `/v1/person/{id}/_certificate` |
| ❌ | `POST` | `/v1/person/{id}/_certificate` |
| ❌ | `GET` | `/v1/person/{id}/_credential/_status` |
| ❌ | `GET` | `/v1/person/{id}/_credential/_update` |
| ✅ | `GET` | `/v1/person/{id}/_credential/_update_intent` |
| ✅ | `GET` | `/v1/person/{id}/_credential/_update_intent/{ttl}` |
| ❌ | `POST` | `/v1/person/{id}/_credential/_update_intent_send` |
| ❌ | `POST` | `/v1/person/{id}/_identify/_user` |
| ❌ | `GET` | `/v1/person/{id}/_radius` |
| ❌ | `POST` | `/v1/person/{id}/_radius` |
| ❌ | `DELETE` | `/v1/person/{id}/_radius` |
| ❌ | `GET` | `/v1/person/{id}/_radius/_token` |
| ❌ | `GET` | `/v1/person/{id}/_ssh_pubkeys` |
| ❌ | `POST` | `/v1/person/{id}/_ssh_pubkeys` |
| ❌ | `GET` | `/v1/person/{id}/_ssh_pubkeys/{tag}` |
| ❌ | `DELETE` | `/v1/person/{id}/_ssh_pubkeys/{tag}` |
| ✅ | `POST` | `/v1/person/{id}/_unix` |
| ❌ | `PUT` | `/v1/person/{id}/_unix/_credential` |
| ❌ | `DELETE` | `/v1/person/{id}/_unix/_credential` |

### service_account (7/20 · 35%)

| | Method | Endpoint |
|---|---|---|
| ✅ | `GET` | `/v1/service_account` |
| ✅ | `POST` | `/v1/service_account` |
| ✅ | `GET` | `/v1/service_account/{id}` |
| ✅ | `DELETE` | `/v1/service_account/{id}` |
| ✅ | `PATCH` | `/v1/service_account/{id}` |
| ✅ | `GET` | `/v1/service_account/{id}/_api_token` |
| ✅ | `POST` | `/v1/service_account/{id}/_api_token` |
| ❌ | `DELETE` | `/v1/service_account/{id}/_api_token/{token_id}` |
| ❌ | `GET` | `/v1/service_account/{id}/_attr/{attr}` |
| ❌ | `PUT` | `/v1/service_account/{id}/_attr/{attr}` |
| ❌ | `POST` | `/v1/service_account/{id}/_attr/{attr}` |
| ❌ | `DELETE` | `/v1/service_account/{id}/_attr/{attr}` |
| ❌ | `GET` | `/v1/service_account/{id}/_credential/_generate` |
| ❌ | `GET` | `/v1/service_account/{id}/_credential/_status` |
| ❌ | `POST` | `/v1/service_account/{id}/_into_person` |
| ❌ | `GET` | `/v1/service_account/{id}/_ssh_pubkeys` |
| ❌ | `POST` | `/v1/service_account/{id}/_ssh_pubkeys` |
| ❌ | `GET` | `/v1/service_account/{id}/_ssh_pubkeys/{tag}` |
| ❌ | `DELETE` | `/v1/service_account/{id}/_ssh_pubkeys/{tag}` |
| ❌ | `POST` | `/v1/service_account/{id}/_unix` |
