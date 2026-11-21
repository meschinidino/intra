# Roadmap

Status key:

- **Done locally** means code and local verification exist.
- **Simulated** means local interfaces exercise the behavior without a real
  external system.
- **Gate** means work must stop until named evidence exists.
- **Planned** means accepted scope with no local proof yet.

Effort values marked founder-weeks are estimates, not measurements.

## v0.1.0 — local lab

- [x] Stage 0: Go, Mac-first local surface, Tailscale target, Docker lab
  identity substitute, custom guard, broker-owned Postgres gateway.
- [x] Stage 1: Postgres, normal service, logs, metrics, attacker sink.
- [x] Stage 2: owner read, robot denial, direct-route denial, audit records.
- [x] Stage 3 mechanics in lab: just-in-time grants, TTL caps, duplicate
  request keys, signed approval, read-only roles, live-session kill.
- [x] Stage 4 mechanics in lab: metadata watch data, deterministic draft
  rules, replay metric, network-only reach paths.
- [x] Stage 5.5 mechanics in lab: WireGuard and Tailscale import, first
  service apply, timed local onboarding path.
- [x] Incident bundle mechanics on fake services.
- [x] Per-run robot certificates and deterministic egress code.
- [x] Provider simulators, drills, standing-reach score, full local
  simulation.

## Demand gate before product claims

Status: unpassed. The founder amendment of 2026-11-18 defers this gate for
the founder-operated pilot. It stays required before any demand claim.

- [ ] Run five 30-minute practitioner calls: at least three tech leads or
  infrastructure owners and at most two agent builders.
- [ ] Record each current production access path and one incident delay.
- [ ] Rank blocked incident access against split human and robot policy.
- [ ] Rank all six expansion pains.
- [ ] Pass only if at least three of five, including two buyer-profile
  participants, describe flat reach, a blunt access cut, or blocked
  incident reads in their own words.
- [ ] Stop and reframe if incumbents already satisfy the wedge or split
  policy outranks blocked incident access.

## v0.2.0 — founder-operated pilot

- [ ] Pass the demand gate. Deferred for this pilot by the founder
  amendment of 2026-11-18. Still unpassed.
- [ ] Complete a security review by a reviewer who did not build the guard.
- [ ] Name one low-risk real service and its written owner consent.
- [ ] Record two-person approval and log-retention decisions.
- [x] Test grant-store restore, approver-device rebinding, restart, and
  policy rollback locally on fake state (`./scripts/pilot-drills.sh`).
  The 15-minute rollback on a real service stays open.
- [ ] Collect at least 200 watch-mode decisions over at least seven days.
- [ ] Enforce only after clean replay on that service's own traffic.
- [x] Prepare the demand-gate and Stage 5 pilot packages (`docs/pilot/`).
  The packages create no evidence.

## v0.3.0 — operational hardening

- [ ] Repeat ten-minute onboarding with at least two non-founder testers.
- [x] Add scheduled audit anchoring and alert delivery. Done locally ahead
  of schedule: `intrad` scheduler, `anchor-once`, file and webhook sinks.
- [x] Add store backup checks to release verification. Done locally ahead
  of schedule: backup checksums, restore verification, artifact checksums.
- [x] Add service-level partial-revoke reporting to operator status. Done
  locally ahead of schedule: `intra status` lists failed child services.
- [ ] Test Mac sleep, restart, disk-full, clock skew, and stale certificate
  behavior as named manual release lanes.
- [ ] Record the guard-host availability decision from pilot evidence.

## v0.4.0 — incident and robot workflows

- [ ] Prove a real log and metric surface, starting with Grafana Viewer.
- [ ] Name a backup approver or record accepted wait behavior.
- [ ] Run one robot workload in a real local runtime with per-run identity.
- [ ] Prove template escape, direct egress, and expired-run denial against
  that runtime.

## v0.5.0 — SaaS and vendor identity

- [ ] Set a written tenant and license budget before provider work starts.
- [ ] Spike Microsoft Entra ID grant issue, expiry, and session kill.
- [ ] Spike Google Workspace grant issue, expiry, and session kill.
- [ ] Keep any provider without timely session kill inventory-only.
- [ ] Join network and SaaS paths in one reach inventory.

## Later

- [ ] Schedule safe self-attack drills after real-service proof.
- [ ] Compare simulated reach against drill results with no missed
  sensitive path.
- [ ] Add protocols only when a named service and test owner exist.
- [ ] Consider SSH, MySQL, and Redis in that evidence order, not by guess.

## Explicit non-goals for the current line

- Hosted Intra control plane.
- Web dashboard.
- Language-model allow decisions.
- Production claims from simulated tenants or fake services.
- Real secrets in the demo state.
