# Security review checklist

The reviewer must not have built the guard. No open Critical or High finding may remain.

## Scope

- [ ] Threat model, including broker compromise.
- [ ] Enforcement topology and direct-route denial.
- [ ] Identity lifecycle: enrollment, approver binding, device loss, rebind.
- [ ] Grant lifecycle: request keys, TTL caps, approval signature, Touch ID gate.
- [ ] Revoke order: state flips non-usable before external kill.
- [ ] Partial revoke and escalation after three failed tries.
- [ ] Audit chain, external anchor, scheduled anchoring, and alert delivery.
- [ ] Audit failure blocks new grants. No silent grant path exists.
- [ ] Backup and restore: checksums, chain verify, no keys inside the backup.
- [ ] Store format upgrade and downgrade refusal.
- [ ] Egress proxy: validated global IPs only, no direct route or DNS.
- [ ] Logs hold metadata only. No bodies, literals, tokens, or canary values.
- [ ] Failure behavior: guard down means closed, never open.

## Method

- [ ] Reviewer read the governing design and this checklist.
- [ ] Reviewer ran `./scripts/verify.sh` and `./scripts/pilot-drills.sh`.
- [ ] Reviewer tried at least one bypass per topology class.

## Findings

| # | Severity | Finding | Owner | State |
|---|---|---|---|---|
| 1 |  |  |  |  |

Reviewer: ____________________  Date: ____________
