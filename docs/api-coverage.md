## API Coverage: 29/85 endpoints (34.1%)
> Kanidm v1.9.2 · OpenTofu Provider

### Coverage by category

| Category | Implemented | Total | % |
|---|---|---|---|
| account | 0 | 8 | ⬜⬜⬜⬜⬜ 0% |
| group | 6 | 12 | 🟩🟩⬜⬜⬜ 50% |
| oauth2 | 8 | 16 | 🟩🟩⬜⬜⬜ 50% |
| other | 0 | 1 | ⬜⬜⬜⬜⬜ 0% |
| person | 8 | 28 | 🟩⬜⬜⬜⬜ 29% |
| service_account | 7 | 20 | 🟩⬜⬜⬜⬜ 35% |

### Not implemented (56 endpoints)

| Method | Endpoint |
|---|---|
| `GET` | `/status` |
| `GET` | `/v1/account/{id}/_radius/_token` |
| `POST` | `/v1/account/{id}/_radius/_token` |
| `GET` | `/v1/account/{id}/_ssh_pubkeys` |
| `GET` | `/v1/account/{id}/_ssh_pubkeys/{tag}` |
| `POST` | `/v1/account/{id}/_unix/_auth` |
| `POST` | `/v1/account/{id}/_unix/_token` |
| `GET` | `/v1/account/{id}/_user_auth_token` |
| `GET` | `/v1/account/{id}/_user_auth_token/{token_id}` |
| `GET` | `/v1/group/_search/{id}` |
| `GET` | `/v1/group/{id}/_attr/{attr}` |
| `PUT` | `/v1/group/{id}/_attr/{attr}` |
| `POST` | `/v1/group/{id}/_attr/{attr}` |
| `DELETE` | `/v1/group/{id}/_attr/{attr}` |
| `GET` | `/v1/group/{id}/_unix/_token` |
| `GET` | `/v1/oauth2` |
| `POST` | `/v1/oauth2/{rs_name}/_claimmap/{claim_name}` |
| `POST` | `/v1/oauth2/{rs_name}/_claimmap/{claim_name}/{group}` |
| `DELETE` | `/v1/oauth2/{rs_name}/_claimmap/{claim_name}/{group}` |
| `POST` | `/v1/oauth2/{rs_name}/_image` |
| `DELETE` | `/v1/oauth2/{rs_name}/_image` |
| `POST` | `/v1/oauth2/{rs_name}/_sup_scopemap/{group}` |
| `DELETE` | `/v1/oauth2/{rs_name}/_sup_scopemap/{group}` |
| `GET` | `/v1/person/_search/{id}` |
| `GET` | `/v1/person/{id}/_attr/{attr}` |
| `PUT` | `/v1/person/{id}/_attr/{attr}` |
| `POST` | `/v1/person/{id}/_attr/{attr}` |
| `DELETE` | `/v1/person/{id}/_attr/{attr}` |
| `GET` | `/v1/person/{id}/_certificate` |
| `POST` | `/v1/person/{id}/_certificate` |
| `GET` | `/v1/person/{id}/_credential/_status` |
| `GET` | `/v1/person/{id}/_credential/_update` |
| `POST` | `/v1/person/{id}/_identify/_user` |
| `GET` | `/v1/person/{id}/_radius` |
| `POST` | `/v1/person/{id}/_radius` |
| `DELETE` | `/v1/person/{id}/_radius` |
| `GET` | `/v1/person/{id}/_radius/_token` |
| `GET` | `/v1/person/{id}/_ssh_pubkeys` |
| `POST` | `/v1/person/{id}/_ssh_pubkeys` |
| `GET` | `/v1/person/{id}/_ssh_pubkeys/{tag}` |
| `DELETE` | `/v1/person/{id}/_ssh_pubkeys/{tag}` |
| `PUT` | `/v1/person/{id}/_unix/_credential` |
| `DELETE` | `/v1/person/{id}/_unix/_credential` |
| `DELETE` | `/v1/service_account/{id}/_api_token/{token_id}` |
| `GET` | `/v1/service_account/{id}/_attr/{attr}` |
| `PUT` | `/v1/service_account/{id}/_attr/{attr}` |
| `POST` | `/v1/service_account/{id}/_attr/{attr}` |
| `DELETE` | `/v1/service_account/{id}/_attr/{attr}` |
| `GET` | `/v1/service_account/{id}/_credential/_generate` |
| `GET` | `/v1/service_account/{id}/_credential/_status` |
| `POST` | `/v1/service_account/{id}/_into_person` |
| `GET` | `/v1/service_account/{id}/_ssh_pubkeys` |
| `POST` | `/v1/service_account/{id}/_ssh_pubkeys` |
| `GET` | `/v1/service_account/{id}/_ssh_pubkeys/{tag}` |
| `DELETE` | `/v1/service_account/{id}/_ssh_pubkeys/{tag}` |
| `POST` | `/v1/service_account/{id}/_unix` |
