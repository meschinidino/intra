# HTTP API reference

All JSON APIs use UTF-8 JSON. Control-plane routes use mTLS when local
certificates exist plus the `X-Intra-Admin` header.

## Broker

Base URL: `https://127.0.0.1:18081` in the lab.

### `GET /health`

Returns `200` with no body contract. It does not prove grant readiness.

### `GET /grants`

Returns public grant records. Secrets and grant tokens never appear.

### `POST /grants`

Creates one grant request. See `broker.Request` field names in the
[policy workflow](../how-to/request-access.md).

Responses:

- `201`: first creation. The response includes the one-time `token`.
- `200`: exact duplicate. The response never returns the token again.
- `400`: malformed JSON.
- `403`: policy, audit, TTL, template, or request-key shape failure.

### `GET /grants/{id}`

Returns one public grant record or `404`.

### `POST /grants/{id}/approve`

Body fields: `approver`, `signature`, and `touch_id`. With a peer
certificate, the certificate actor overrides the body approver. The
signature covers the full grant digest. Sensitive grants require Touch
ID proof or a server-side local Touch ID helper success.

### `POST /grants/{id}/revoke`

Body field: `by`. With a peer certificate, the certificate actor
overrides the body value. `200` returns the revoked record. `409`
means kill work remains in partial or escalated state.

### `GET /validate`

Internal query surface for service checks. Parameters are `actor`,
`device`, `service`, and `action`. The grant token travels in the
`Authorization: Bearer` header. Success is `204`; denial is `403`.

## Audit

Base URL: `https://127.0.0.1:18082` in the lab.

- `GET /health`: liveness only.
- `POST /events`: append one metadata event. Admin token required when
  configured.
- `GET /head`: current sequence and hash head. Admin token required when
  configured.

Audit downtime returns `503` on append. Grant issuance must stop.

## Guard

Guard paths use `/svc/<service>/<rest>`.

- `GET /health`: liveness only.
- Any `/svc/...` request: certificate identity, policy, manifest when
  enabled, grant validation where required, audit, then proxy.
- Upstreams never receive client `Authorization`, `X-Intra-Actor`,
  `X-Intra-Device`, or `X-Intra-Admin` headers.

## Egress proxy

The egress surface is an HTTP forward proxy on port `18083` in the lab.

- `CONNECT host:port`: policy, grant, DNS, and global-IP checks run
  before dialing the validated IP.
- Absolute-form HTTP requests use the same checks.
- HTTPS must use `CONNECT`. TLS bodies stay opaque.

## Provider simulators

Entra and Google simulators expose `GET /inventory`, `POST /grants`, and
`DELETE /grants/{id}`. They exist only for local exercises and mark no
real tenant proof.
