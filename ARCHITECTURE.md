# Architecture

This page gives the short system map. The governing detail lives in
[the design record](docs/designs/intra-design.md).

## System map

```text
Owner CLI ──mTLS + admin token──► Broker ──► Grant store
    │                                │
    │                                ├──► Postgres gateway ──► Postgres
    │                                └──► Audit queue ──► Hash chain
    │
    └────mTLS + grant token────► Guard ──► HTTP services

Robot ──mTLS identity──► Guard or broker gateway
Robot ──grant token──► Egress proxy ──► approved public hosts only

Policy JSON ──compile──► Mesh notes + guard upstreams + broker rules
```

## Components

| Component | Owns | Must never own |
|---|---|---|
| Mesh | Device identity and encrypted routes | Service action decisions |
| Guard | HTTP identity, policy check, reverse proxy | Grant issuance or database roles |
| Broker | Grant lifecycle, approval checks, sessions | Policy authoring |
| Postgres gateway | Broker-carried database sessions and kill | Client-held database passwords |
| Audit writer | One ordered hash chain | Grant decisions |
| Policy compiler | Deterministic generated layer views | Hand-edited generated output |
| Egress proxy | Deterministic host and port exit rules | TLS body inspection |
| CLI | Operator and client workflows | Hidden approval or secret storage |

## Trust boundaries

- Client certificates identify actor and device or run.
- The control plane also requires local mTLS and an admin token.
- Services accept traffic only from the guard or broker gateway identity.
- Postgres accepts grant-role sessions only through the broker gateway in
  the lab topology.
- The broker database credential can call restricted grant functions. It
  cannot read protected tables directly.
- The audit anchor key stays outside the data directory by default.

## Main flows

### Static HTTP read

1. The client presents a certificate.
2. The guard maps certificate identity to actor and device.
3. The guard checks one policy rule and route action.
4. The guard writes an audit decision.
5. The guard strips client auth headers and proxies the request.

### Sensitive database read

1. The client requests a grant with a unique request key.
2. A human approves the full grant digest. Sensitive grants require
   Touch ID proof in the real approval path.
3. The broker creates one expiring Postgres role with selected reads.
4. The role secret stays in broker memory.
5. The client connects to the broker gateway with its certificate and
   grant token.
6. The gateway checks the statement class before execution.
7. Expiry or revoke terminates live sessions and drops the role.

### Revoke

1. The broker flips and persists a non-usable state.
2. The broker clears the in-memory database secret.
3. The gateway closes live sessions.
4. Postgres terminates remaining backends and drops the role.
5. Audit failure cannot return the grant to active use.

## Failure posture

- Guard down: network access fails closed.
- Audit down: no new grants. Guarded allows stop.
- Broker down: no new grants and no new database sessions.
- Postgres down: grants fail before a live session exists.
- Policy compile failure: no generated config ships.
- Rollback: return to the last known-good rule set. Never restore broad
  standing reach as an automatic rollback.

## Scale boundary

Version 0.1.0 targets one Mac and one local lab. It does not claim team,
multi-host, or production scale. File stores, one audit writer, one
broker, and one guard host are explicit lab limits.

## Related

- [Security model](docs/explanation/security-model.md)
- [Why Intra exists](docs/explanation/why-intra.md)
- [Design decisions](docs/explanation/design-decisions.md)
- [Configuration reference](docs/reference/configuration.md)
