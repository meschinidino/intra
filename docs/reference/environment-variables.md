# Environment variables

## CLI

| Variable | Default | Purpose |
|---|---|---|
| `INTRA_DATA` | `.intra` | Default data directory |
| `INTRA_BROKER_URL` | Local broker URL | Override broker base URL |
| `INTRA_ADMIN_TOKEN` | Data-dir token | Override control-plane token |
| `INTRA_ANCHOR_KEY` | `~/.config/intra/anchor.key` | External anchor HMAC key path |
| `INTRA_TOUCHID_HELPER` | empty | Default Touch ID helper binary |
| `INTRA_AUDIT_ADDR` | `127.0.0.1:18082` | Audit address used by `status` |
| `INTRA_GUARD_HTTP_ADDR` | `127.0.0.1:18080` | Plain guard address used by `status` |
| `INTRA_ANCHOR_MAX_AGE_SEC` | `86400` | Anchor age that marks `status` output stale |
| `INTRA_ALERT_WEBHOOK` | empty | Alert webhook URL for `anchor-once` and the scheduler |

## Server

| Variable | Default | Purpose |
|---|---|---|
| `INTRA_DATA` | `.intra` | Data directory default |
| `INTRA_AUDIT_ADDR` | `127.0.0.1:18082` | Audit listen or client address |
| `INTRA_BROKER_ADDR` | `127.0.0.1:18081` | Broker listen address |
| `INTRA_GUARD_ADDR` | `0.0.0.0:18443` | Guard HTTPS listen address |
| `INTRA_GUARD_HTTP_ADDR` | `127.0.0.1:18080` | Plain guard fallback listen address |
| `INTRA_EGRESS_ADDR` | `127.0.0.1:18083` | Egress proxy listen address |
| `INTRA_PG_ADDR` | `127.0.0.1:15432` | Postgres gateway listen address |
| `INTRA_PG_ADMIN_URL` | empty | Broker-admin Postgres connection string |
| `INTRA_PG_BASE_URL` | Admin URL | Backend host and database template |
| `INTRA_TOUCHID_HELPER` | empty | Server-side Touch ID helper path |
| `INTRA_ALLOW_HEADER_ID` | unset | Set `1` only for tests or controlled lab wiring |
| `INTRA_VERIFY_MANIFESTS` | unset | Set `1` to verify robot run manifests |
| `INTRA_ANCHOR_INTERVAL_SEC` | empty | Audit anchor scheduler interval; empty disables it |
| `INTRA_ANCHOR_MAX_AGE_SEC` | `86400` | Anchor age limit for the scheduler |
| `INTRA_ALERT_WEBHOOK` | empty | Alert webhook URL for scheduled anchoring |
| `INTRA_ENTRA_SIM_ADDR` | `127.0.0.1:18084` | Entra simulator listen address |
| `INTRA_GOOGLE_SIM_ADDR` | `127.0.0.1:18085` | Google simulator listen address |

## Compose and tests

| Variable | Purpose |
|---|---|
| `INTRA_UID`, `INTRA_GID` | Container user for bind-mounted lab state |
| `LAB_VIEWER_TOKEN` | Shared fake log and metric viewer token |
| `LAB_MODE` | `normal`, `sensitive-http`, `logs`, `metrics`, or `attacker` |
| `LAB_ADDR` | Fake service listen address |
| `LAB_CANARY` | Fake sensitive value for leak checks |
| `INTRA_TEST_PG_ADMIN_URL` | Integration-test admin connection |
| `INTRA_TEST_PG_BASE_URL` | Integration-test backend template |
