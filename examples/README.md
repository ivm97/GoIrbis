# Examples

Runnable demos for `github.com/ivm97/GoIrbis/irbis`.

| Directory | Needs IRBIS64 server | Notes |
|-----------|----------------------|-------|
| `search_and_read` | yes | search + read + `@brief` |
| `write_records` | yes | create records in `SANDBOX` |
| `safe_experiments` | yes | smoke walk across many API methods |
| `rqst_shrink` | yes | admin-oriented sample |
| `direct_access` | no | reads local MST/XRF under `data/` |
| `iso` | no | ISO2709 sample under `data/` |
| `mst2text` / `quintessence` | no | offline MST utilities |
| `icqbot` | yes | legacy ICQ/mail.ru bot; needs `github.com/mail-ru-im/bot-golang` |

```bash
# from repo root
go run ./examples/search_and_read
go run ./examples/direct_access
```

Server-backed examples use host `localhost`, user `librarian`, password `secret` — change before running.
