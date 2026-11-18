# Configuration

Intra keeps runtime state in one data directory. Source configuration is
reviewable. Generated state is local and ignored by git.

## Data directory

Default: `.intra/`

| Path | Owner | Contents |
|---|---|---|
| `policy.json` | Operator | One policy source |
| `compiled.json` | `intra compile` | Deterministic generated view |
| `grants.json` | Broker | Public grant records and token hashes |
| `audit.jsonl` | Audit writer | Hash-chained decision metadata |
| `audit.jsonl.lock` | Audit writer | Single-writer lock |
| `anchor.json` | `intra anchor` | Signed audit head |
| `approval.key` | Init | Approval signing key, mode `0600` |
| `admin.token` | Init | Control-plane token, mode `0600` |
| `viewer.token` | Init | Lab log and metric viewer token |
| `certs/` | Init and identity commands | CA, server, client, run certs |

Never commit this directory.

## Server command

```text
intrad [all|audit|broker|guard|providers] --data <dir> --policy <file>
```

- `all` starts audit, broker, guard, egress, and provider simulators.
- `audit` starts only the audit writer service.
- `broker` starts broker and, when Postgres env is set, the gateway.
- `guard` requires broker state in the same process in this build. Use
  `all` for the supported local path.
- `providers` starts only the two local provider simulators.

## Ports

| Port | Surface | Notes |
|---:|---|---|
| `15432` | Postgres gateway | Client certificate in lab mode |
| `18080` | Guard plain HTTP | Only when no guard certificate exists |
| `18081` | Broker | mTLS plus admin token when certs exist |
| `18082` | Audit | mTLS plus admin token when certs exist |
| `18083` | Egress proxy | mTLS when certs exist |
| `18084` | Entra simulator | Lab only |
| `18085` | Google simulator | Lab only |
| `18443` | Guard HTTPS | Primary operator path |

## Startup rules

- `intrad` fails before serving control traffic when `admin.token` is
  missing outside provider-only mode.
- Control services require guard TLS material.
- The Postgres gateway requires TLS material when a Postgres admin URL is
  configured.
- Plain HTTP guard mode exists only for unit and development wiring where
  no certificate was made.

## Updates

Updates must download first, pass attack and smoke suites, then prompt
the owner. No unattended update may change guard behavior.

Grant-store and audit readers must read the prior release format during
an upgrade window.
