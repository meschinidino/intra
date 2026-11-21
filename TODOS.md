# TODOs

This file holds work that still needs evidence, people, or external
systems. Local code completion does not close these items.

## P1 — demand proof

Status: deferred by the founder amendment of 2026-11-18 for the
founder-operated pilot. Still owed before any demand claim.

- [ ] Book and run five practitioner interviews.
  - At least three tech leads or infrastructure owners.
  - At most two agent builders.
  - Record names, roles, quotes, current production access path, and one
    incident delay per person.
  - Rank blocked incident access and all six expansion pains.
  - Gate: current demand gate in `docs/designs/intra-design.md`.
- [ ] Observe one real access workflow without guiding the participant.
  - Record each delay, workaround, and broad standing path.
- [ ] Decide whether the buyer, approver, and daily user are the same role.
  - Name who runs the guard host in a small team.
  - Name who approves sensitive grants when the founder is away.

## P1 — first real pilot

- [ ] Name one low-risk service with no customer, personal, or payment
  data and no production secrets.
- [ ] Get written consent from its named owner.
- [ ] Run an independent security review with no open Critical or High
  findings.
- [ ] Test restore, approver-device rebinding, restart, and rollback on
  the real pilot service. Local fake-state drills pass in
  `./scripts/pilot-drills.sh`. Local proof does not close this item.
- [ ] Collect at least 200 watch-mode decisions across at least seven days.
- [ ] Run clean replay on that service before any enforcement.

## P1 — external proof

- [ ] Install Tailscale and repeat identity, route, revoke, and stale-device
  checks outside Docker substitute identity.
- [ ] Run the real macOS Touch ID helper in a manual approval lane.
- [ ] Repeat onboarding with two non-founder testers on fresh Mac user
  accounts.
- [ ] Test Mac sleep and wake during an active grant.

## P2 — provider proof

- [ ] Set a written Microsoft and Google test-tenant budget.
- [ ] Verify grant issue, expiry, revoke, and 15-minute session kill in a
  real Microsoft Entra ID test tenant.
- [ ] Verify the same lifecycle in a real Google Workspace test tenant.
- [ ] Mark any provider without timely session kill as inventory-only.

## P2 — product depth

- [ ] Prove one real log and metric surface, starting with Grafana Viewer.
- [ ] Name a backup approver or record accepted grant wait behavior.
- [ ] Run a robot workload in a real local runtime with a per-run cert.
- [x] Add scheduled anchor copy and an alert sender. Done locally:
  `intrad` scheduler, `intra anchor-once`, file and webhook alert sinks.
  A real paging destination stays unproven.
- [ ] Add SSH, MySQL, or Redis only after a named service needs it.

## P3 — release maturity

- [ ] Decide the support window for grant-store and audit formats.
- [ ] Add signed release artifacts. Checksum verification is done locally
  (`intra release-checksums`, `intra verify-release`). Signing stays open.
- [ ] Add container image publishing after the first real pilot.
- [x] Write an upgrade test from the prior release format. Done locally:
  the store test loads the v0.1.0 bare-list format and rejects newer formats.
