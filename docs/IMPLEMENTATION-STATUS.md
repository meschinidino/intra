# Implementation status

Updated from the 2026-11-18 local verification run.

| Stage | Local state | Evidence | Still outside local proof |
|---|---|---|---|
| 0 Picks | Done | [STAGE0](STAGE0.md), Go build, Docker images present | Real Tailscale install |
| 1 Lab skeleton | Done locally | Docker lab with Postgres, normal, logs, metrics, attacker sink | None for fake lab |
| 2 First slice | Done locally | mTLS owner read, robot denial, no direct service ports in compose | Real mesh route proof |
| 3 JIT grants | Done locally | Request keys, TTL caps, signed approval, read-only role, live kill | Real Touch ID prompt in automation |
| 4 Learn/replay | Done locally | Metadata audit, draft grouping, replay metric, reach paths | Real traffic corpus of 200 decisions |
| 5 Real pilot | Not run | Entry gates and rollback design only | Named real service, owner consent, security review |
| 5.5 Onboarding | Done locally | Tailscale and WireGuard import, first service apply, timing in verify script | Two non-founder timed runs |
| 6 Incident room | Done on fake services | Bundle grant reads Postgres, logs, metrics; revoke denies return | Real Grafana and named backup approver |
| 7 Robot parity | Done locally in part | Per-run certificates and manifests, templates, deterministic egress proxy code | Deployed robot workload identity in a real runtime |
| 8 SaaS identity | Simulated only | Entra and Google simulator issue, inventory, revoke, kill latency | Real tenants, paid licenses, provider session behavior |
| 9 Drills | Done locally | Distinct-path score, canary and direct-path drills | Scheduled production drills |
| 10 Full simulation | Done on local inventory | Reach paths include policy and simulator outputs | Complete real SaaS inventory |

## Critical gaps from reviews: local disposition

- Late session kill: integration test proves a live query dies after revoke and reconnect fails. The 60-second deadline is met in the local test by immediate broker close plus role drop.
- Anchor failure visibility: `intra status` reports anchor age. `verify-anchor` fails on a tampered or truncated chain. A scheduled alert sender is not implemented.
- Partial revoke: state, retry count, escalation after three failed tries, and an audit alert event exist. A per-service partial success matrix for real Grafana and SaaS classes remains simulator-level.

## Demand gate

The first lab user is Dino Meschini, founder and Mac owner. That closes the
lab-user question only. It does not prove an external buyer. Stage 3 product
claims beyond the local lab still need the five practitioner interviews in the
governing plan.
