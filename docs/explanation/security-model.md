# Security model

## Claim

Intra aims to reduce blast radius while keeping approved work possible.
It does not claim attack prevention or production readiness in v0.1.0.

## Assets

- Sensitive service data and actions
- Grant tokens and ephemeral database credentials
- Approval, admin, anchor, and certificate keys
- Policy and generated configuration
- Audit history and reach inventory
- Live database and service sessions

## Actors

| Actor | Expected ability | Main limit |
|---|---|---|
| Owner | Approve, request, revoke, operate | Bound Mac plus Touch ID for sensitive grants |
| Robot | Request bounded task access | No self-approval, short identity, scoped egress |
| Service | Serve its own upstream traffic | Guard or broker identity only |
| Drill | Probe denied paths with fake data | No write grant |
| Unknown device | None | No route and no identity |
| Compromised robot | Valid run identity until revoke | Template scope and egress allowlist |
| Compromised service | Its own service identity | No sensitive or admin route |

## Trust boundaries

### Identity

Certificates carry `actor/device` in the common name for the local lab.
TLS verifies the certificate. The guard reads identity from the peer
certificate, not from a caller header, unless a test-only header mode is
explicitly enabled.

Per-run robot certificates expire with the signed manifest window. The
guard can verify manifest signature, expiry, and certificate fingerprint
when manifest verification is enabled.

### Grant authority

- Policy decides whether a request shape can exist.
- A signed approval covers grant ID, approver, actor, device, services,
  actions, expiry, and template.
- Sensitive approval needs fresh local Touch ID confirmation in the real
  approval path.
- The broker flips grants to a non-usable state before revoke cleanup.
- Duplicate request keys bind to one request shape and never return the
  token twice.

### Database authority

The broker admin role can execute only two grant functions. Those
functions create and drop short-lived read-only roles for named objects.
The broker admin holds no direct select grant on protected data.

Gateway sessions authenticate with a grant token plus client
certificate when TLS is configured. The ephemeral database password
stays in broker memory.

### Audit integrity

Guard, broker, gateway, and egress submit metadata to one writer. Query
text becomes a literal-free fingerprint. Bodies, literals, results,
tokens, and passwords never enter audit records.

The hash chain detects partial edits. The external HMAC anchor catches
clean truncation and full-chain replacement only if the anchor key and
anchor file stay outside attacker control.

## Failure behavior

| Failure | Required result |
|---|---|
| Unknown identity | Deny |
| Expired identity or grant | Deny |
| Audit writer down | No new grants, no guarded allows |
| Guard down | No network path |
| Broker down | No new grants or database sessions |
| Database down | Grant issue fails closed |
| Session kill timeout | Partial-revoke state and alert path |
| Policy compile error | No generated config ships |

## Explicit non-claims

- An authorized reader can copy data by hand.
- Malware fully controlling an approved Mac during a live grant can act
  inside that grant window.
- A trusted vendor or SaaS path outside the private network stays out of
  Stages 0-5.
- Simulated provider results prove no real tenant behavior.
- Local mTLS in Docker proves identity shape, not real Tailscale behavior.
- Development approval in scripts proves no biometric event.

## Security invariants for contributors

1. Default deny wins.
2. Unknown actions map to admin, never read.
3. No silent grant exists.
4. No database password reaches a client.
5. No hand-edited generated config.
6. No secret enters logs, docs, examples, or tests.
7. No new allow path without a bypass test.
