# CLI reference

The operator binary is `intra`. Most commands print JSON. Errors print
to stderr and exit non-zero.

## Global behavior

- Default data directory: `.intra`, or `INTRA_DATA`.
- Default broker URL: `https://127.0.0.1:18081` when local certs exist,
  otherwise `http://127.0.0.1:18081`.
- `INTRA_BROKER_URL` overrides the broker URL.
- Broker calls send `X-Intra-Admin` when an admin token is available.
- Secrets print once only where the command contract says so.

## `intra version`

Print the release version.

```sh
intra version
```

## `intra init-demo`

Create demo policy, keys, certificates, compiled config, and empty
runtime stores.

| Flag | Default | Meaning |
|---|---:|---|
| `--data` | `.intra` | Data directory |
| `--normal` | `http://127.0.0.1:18090` | Normal upstream |
| `--logs` | `http://127.0.0.1:18091` | Logs upstream |
| `--metrics` | `http://127.0.0.1:18092` | Metrics upstream |
| `--force` | `false` | Replace demo policy, keys, grants, and audit |

The anchor key goes to `INTRA_ANCHOR_KEY` when set. Otherwise it goes to
`~/.config/intra/anchor.key`.

## `intra compile`

Validate policy and write the deterministic compiled view.

| Flag | Default |
|---|---:|
| `--policy` | `.intra/policy.json` |
| `--out` | `.intra/compiled.json` |

## `intra request`

Create one grant request. Required fields are reason, at least one
service, and at least one action.

| Flag | Default |
|---|---:|
| `--request-key` | generated timestamp key |
| `--actor` | `owner` |
| `--device` | `mac` |
| `--services` | `sensitive-db` |
| `--actions` | `read` |
| `--reason` | empty, rejected |
| `--incident` | empty |
| `--template` | empty |
| `--ttl` | `1800` seconds |

The first response contains the one-time grant token. A duplicate with
the same key and same request shape returns state without the token.
The same key with a different shape fails.

## `intra approve`

Approve one requested grant.

| Flag | Default |
|---|---:|
| `--data` | `.intra` |
| `--id` | required |
| `--approver` | `owner` |
| `--touchid-helper` | `INTRA_TOUCHID_HELPER` |
| `--dev-approve` | `false` |

Sensitive grants need a successful Touch ID helper run or the explicit
development flag. The development flag proves no biometric event.

## `intra revoke`

Revoke one grant by `--id`. `--by` defaults to `owner`.

## `intra grants`

List public grant records. Database secrets and grant tokens never print.

## `intra status`

Print broker, audit, guard, audit head, grant counts, and anchor age.
`--data` defaults to `.intra`.

Output also holds `grants_by_state`, `partial_revokes` (grant id, failed
child services, revoke tries, escalation), `late_kills` (expired active
grants and grants stuck in a revoke state), `anchor_stale`, and
`anchor_reason`. `INTRA_ANCHOR_MAX_AGE_SEC` sets the stale limit.

## `intra anchor` and `intra verify-anchor`

Write or verify the HMAC-signed audit head at `.intra/anchor.json`.
`verify-anchor` prints `anchor OK` on success.

## `intra anchor-once`

Run one scheduled-anchor pass. It anchors when the anchor is missing, stale,
or behind the chain head. `--max-age-sec` defaults to `86400`.
Alerts go to `.intra/alerts.jsonl` and to `INTRA_ALERT_WEBHOOK` when set.
A broken chain refuses anchoring and writes one `anchor_failed` alert.

The server runs the same pass on a timer when `INTRA_ANCHOR_INTERVAL_SEC`
is set on `intrad`.

## `intra backup` and `intra restore`

`intra backup --out backup.zip` writes one zip with a checksum manifest.
It holds grants, audit, anchor, and policy state only. It never holds keys,
certificates, or tokens.

`intra restore --from backup.zip --data DIR` verifies every checksum, the
audit chain, the anchor, and the grant-store format before install.
Stop `intrad` first. Restore over existing state needs `--force`.

## `intra rebind-approver`

Replace the approval key after the old device is revoked or declared lost.
`--confirm-old-revoked` is required. The command writes fingerprint records
only, never key copies. Restart `intrad` after rebind.

## `intra release-checksums` and `intra verify-release`

`intra release-checksums --dir DIR --files intra --out DIR/SHA256SUMS`
writes SHA-256 checksums. `intra verify-release --dir DIR --checksums
DIR/SHA256SUMS` fails on any missing or mismatched artifact.

## `intra replay`

Score audit allows against a candidate policy.

| Flag | Default |
|---|---:|
| `--data` | `.intra` |
| `--policy` | `.intra/policy.json` |

Output fields: `total`, `false_blocks`, `rate_percent`, and `clean`.
Clean needs at least 200 decisions and a rate under 1 percent.

## `intra draft`

Group recorded allows into deterministic candidate rules. It uses no
model. `--data` defaults to `.intra`.

## `intra simulate`

List actor-service-action paths from policy rules. `--policy` defaults to
`.intra/policy.json`.

## `intra drills`

Run direct-path drills from a JSON target file passed with `--config`.
See [file formats](file-formats.md) for the target schema.

## `intra import`

```sh
intra import <wireguard|tailscale> <file>
```

Print imported reach without changing policy.

## `intra onboard`

Import reach and add one service behind the guard.

| Flag | Default |
|---|---:|
| `--data` | `.intra` |
| `--kind` | `tailscale` |
| `--file` | required |
| `--service` | `imported` |
| `--upstream` | required |
| `--owner` | `founder` |

## `intra issue-identity`

Mint a per-run client certificate and signed manifest.

| Flag | Default |
|---|---:|
| `--data` | `.intra` |
| `--actor` | `robot` |
| `--run` | required |
| `--ttl-hours` | `24`, allowed range `1..24` |

## `intra providers`

Exercise local Entra and Google provider simulators. Output marks every
result `simulated: true`.

## `intra latency`

Measure direct and guarded p95 latency. The command exits non-zero when
added p95 latency reaches 50 ms.

| Flag | Meaning |
|---|---|
| `--direct` | Direct comparison URL |
| `--guarded` | Guarded URL |
| `--n` | Request count per side, default `50` |
| `--actor` | Header actor for guarded test mode |
| `--cacert`, `--cert`, `--key` | Guard mTLS files |
