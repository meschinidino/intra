# Architecture in depth

## Design center

Intra separates four questions:

1. Who or what asks?
2. Which device or run carries that identity?
3. Which exact service and action does it need now?
4. Can the system prove and later kill that access?

No component answers all four alone.

## Components

### Policy compiler

One JSON policy names actors, services, rules, templates, routes, and
egress hosts. The compiler validates references and TTLs, then writes one
deterministic compiled view with a content hash.

Generated config prevents three hand-edited rule copies from drifting.

### Guard

The guard terminates mTLS, reads actor and device from the client
certificate, maps the route to an action, checks policy, and asks the
broker to validate grant-bound access.

It strips client authorization and identity headers before proxying. A
weak upstream therefore cannot harvest a reusable grant token.

### Broker

The broker owns grant state. It checks request shape, template scope,
TTL caps, approval signatures, sensitive confirmation, expiry, and
revocation.

For Postgres, the broker also owns the data path. It creates one
ephemeral database role per grant, keeps that secret in memory, checks
each statement class, and terminates live sessions on expiry or revoke.

### Audit writer

Guard, broker, gateway, and egress submit decision metadata to one
queue. One writer appends hash-linked JSON records. A daily signed head
can be copied outside the data store and verified later.

If the writer cannot append, new grants and guarded allows stop. Revoke
still flips grant state first, so audit trouble cannot keep access alive.

### Egress proxy

Robot internet traffic uses a separate deterministic proxy. CONNECT sees
host and port only. TLS bodies stay opaque. The proxy resolves the host,
rejects every non-global answer, and dials the validated IP directly.

### Provider adapters

SaaS adapters expose one contract: issue, inventory, revoke, and measure
kill latency. Local simulators implement that contract for exercises.
Real tenant adapters must prove session kill before any enforcement
claim.

## Data stores

| Store | Writer | Contents | Secret level |
|---|---|---|---|
| Policy JSON | Operator | Actors, services, rules, templates | Internal config |
| Grants JSON | Broker account | Public grant state and token hashes | Sensitive metadata |
| Audit JSONL | Audit writer | Decision metadata only | Sensitive metadata |
| Anchor JSON | CLI | Signed audit head | Integrity record |
| Certificates | Local CA tooling | Device and service identity | Secret keys |

Database grant passwords never enter the grant store.

## Key boundaries

- Policy can allow a path. Only a live grant can use grant-bound paths.
- The broker admin database role can call grant functions. It cannot read
  protected tables directly.
- The approval key signs the full grant digest. It is not a grant token.
- The admin token protects local control APIs. It never replaces mTLS.
- The audit anchor key sits outside the repo data directory by default.

## Scaling limits in v0.1.0

- One Mac can host the full lab, but sleep stops the guard.
- File stores fit one local operator, not concurrent team control.
- Postgres comes first. Other database protocols need their own gateway
  proof and role model.
- One audit writer favors order and integrity over throughput.
- The reach simulator is exact only for inventoried paths.

## Diagrams

The governing [design record](../designs/intra-design.md) contains system,
data-flow, state-machine, deployment, and rollback diagrams. Update them
in the same change as architecture behavior.
