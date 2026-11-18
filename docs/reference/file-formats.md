# File formats

## Policy and compiled policy

See the [policy schema](policy-schema.md). Policy is source.
`compiled.json` is generated and must match the source hash after
`intra compile`.

## Grant store

`grants.json` is a JSON array of public grant records.

Important fields:

| Field | Meaning |
|---|---|
| `id` | Stable grant ID |
| `request_key` | Client idempotency key |
| `request_hash` | Hash binding the key to one request shape |
| `token_hash` | SHA-256 of the one-time grant token |
| `state` | Request, issue, active, revoke, or terminal state |
| `expires_at` | Hard grant expiry |
| `children` | Per-service states for bundle grants |

Database passwords and raw grant tokens never persist.

## Audit log

`audit.jsonl` holds one JSON object per line.

| Field | Meaning |
|---|---|
| `ts` | UTC event time |
| `actor`, `device` | Decision subject |
| `service`, `action` | Requested service and action class |
| `rule`, `result` | Rule class and decision outcome |
| `grant_id` | Grant when one applies |
| `query_hash` | Literal-free SQL fingerprint |
| `seq`, `prev`, `hash` | Hash-chain position and links |

No request bodies, query literals, query results, tokens, or passwords
belong in this file.

## Audit anchor

`anchor.json` contains `seq`, `head`, `sig`, and `ts`. The signature is
HMAC-SHA256 over `seq:head`. Keep its key outside the repository and
outside the server data volume.

## Run manifest

Each per-run certificate has a sibling `<name>.manifest.json`:

| Field | Meaning |
|---|---|
| `actor`, `run_id`, `device` | Certified identity and run scope |
| `expires_at` | Run expiry |
| `cert_name` | Certificate file stem |
| `cert_sha256` | SHA-256 of the DER certificate |
| `signature` | HMAC over actor, run ID, and expiry |

Certificate expiry matches the manifest TTL. The guard also verifies the
fingerprint when manifest mode is enabled.

## Drill targets

`intra drills --config` takes a JSON array:

```json
[
  {
    "name": "sensitive-db",
    "url": "http://127.0.0.1:18093/",
    "sensitive": true,
    "owner": "founder"
  }
]
```

## Import files

- WireGuard: peer `PublicKey` and `AllowedIPs` lines.
- Tailscale: JSON `acls` or `grants` entries with `src` and `dst`
  string arrays.
