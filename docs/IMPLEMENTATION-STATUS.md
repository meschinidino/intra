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
- Anchor failure visibility: `intra status` reports anchor age, `anchor_stale`, and `anchor_reason`. `verify-anchor` fails on a tampered or truncated chain. Scheduled anchoring runs in `intrad` when `INTRA_ANCHOR_INTERVAL_SEC` is set, or once through `intra anchor-once`. Alerts go to `.intra/alerts.jsonl` and to `INTRA_ALERT_WEBHOOK` when set. This is local delivery only. No real paging service is proven.
- Partial revoke: state, retry count, escalation after three failed tries, and an audit alert event exist. `intra status` now lists partial revokes with failed child services and late kills. A per-service partial success matrix for real Grafana and SaaS classes remains simulator-level.
- Backup and restore: `intra backup` and `intra restore` round-trip grants, audit, anchor, and policy with checksums and chain verification. The backup holds no keys or tokens. `./scripts/pilot-drills.sh` proves restore, approver rebinding, policy rollback, restart, and release-artifact checks on fake state.
- Store upgrade: grant store format v1 envelope loads the v0.1.0 bare-list format and rejects newer formats. The upgrade test covers both directions.
- Pilot packages: the demand-gate and Stage 5 packages in `docs/pilot/` are ready to use. They prepare external gates. They are not evidence.

## Demand gate

Status: unpassed, deferred by the founder amendment of 2026-11-18 in the
governing design. The amendment accepts demand risk for the founder-operated
pilot on his own service. It proves no external buyer. Demand claims still
need the five practitioner interviews. Interview work continues through
`docs/pilot/outreach-packet.md`. No person has been contacted yet.
