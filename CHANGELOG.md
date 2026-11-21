# Changelog

All notable changes to Intra follow this file. The format follows
Keep a Changelog, and versions follow Semantic Versioning.

## [Unreleased]

### Added

- Project vision, mission, roadmap, TODO, architecture, design, security,
  operations, reference, tutorial, how-to, and agent documentation.
- Apache-2.0 license, notice, contribution rules, security policy, and
  code of conduct.
- Scheduled audit anchoring in `intrad` and `intra anchor-once`, with file
  and webhook alert delivery.
- `intra backup`, `intra restore`, and `intra rebind-approver` for grant
  store, audit log, and approver recovery on local state.
- `intra release-checksums` and `intra verify-release` for artifact checks.
- Operator status fields for partial revokes, stale anchors, and late kills.
- Grant store format v1 envelope. It loads the v0.1.0 format and rejects
  newer formats.
- `./scripts/pilot-drills.sh` for backup, restore, rebind, rollback,
  restart, and release-artifact drills. `verify.sh` runs it.
- Demand-gate and Stage 5 pilot packages in `docs/pilot/`. The packages
  prepare external gates. They are not evidence.

### Known limitations

- Version stays `0.1.0`. The v0.2.0 pilot gate has not passed.
- No real service, tenant, reviewer, or interview evidence exists yet.
- Founder amendment of 2026-11-18 defers the demand gate for the
  founder-operated pilot. The gate stays unpassed. No demand claim follows
  from the pilot.

## [0.1.0] - 2026-11-18

### Added

- Go policy compiler with actors, services, routes, rules, templates, and
  egress rules.
- Hash-chained audit log with one queue writer and external HMAC anchor.
- Broker grant lifecycle with request keys, TTL caps, signed approvals,
  Touch ID interface, expiry, revoke, and partial-revoke states.
- Broker-owned Postgres gateway with ephemeral read-only roles.
- mTLS HTTP guard with per-service action checks and token stripping.
- Deterministic robot egress proxy with validated-IP dialing.
- Docker lab with Postgres, normal, logs, metrics, and attacker services.
- Incident bundle reads against fake database, log, and metric services.
- Per-run robot certificates and signed manifests.
- Local Microsoft Entra ID and Google Workspace provider simulators.
- Replay, rule drafting, reach simulation, drills, and latency commands.
- WireGuard and Tailscale reach import plus first-service onboarding.
- Unit tests, race-enabled Postgres integration tests, and Docker lab
  end-to-end verification scripts.

### Security

- Audit failure blocks new grants.
- Revoke flips grants to a non-usable state before session kill.
- Duplicate requests never return an existing grant token.
- Postgres secrets stay in broker memory and die with the grant.
- Control-plane HTTP APIs require local mTLS plus an admin token in the
  lab server build.

### Known limitations

- Local lab only. No production claim.
- Real Tailscale, real SaaS tenants, real service pilot, and outside-user
  demand remain unproven.
- Automated lab approval uses an explicit development flag. The compiled
  macOS Touch ID helper still needs a manual prompt run.

[Unreleased]: https://github.com/meschinidino/intra/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/meschinidino/intra/tree/main
