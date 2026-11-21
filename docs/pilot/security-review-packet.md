# Security review packet and reviewer brief

Status: Dino chose route C, community review, on 2026-11-18.
No reviewer is engaged yet. No finding exists yet.
Route C closes the gate only when one named external reviewer completes the
full checklist and signs the findings table. A comment thread alone does not close it.

## Reviewer brief (send as-is after choosing a route)

You will review Intra, a default-deny access layer for one founder-operated pilot.
The reviewer must not have built the guard. The review covers the threat model,
topology, identity lifecycle, broker compromise, stores, and failure behavior.
The gate: no open Critical or High finding before a real service enters.

You get: this repo at the named commit, `./scripts/verify.sh`,
`./scripts/pilot-drills.sh`, the governing design in `docs/designs/intra-design.md`,
and the checklist in [security review](security-review.md).

You try: one bypass per topology class, one audit-failure grant attempt, one
partial-revoke attempt, and one restore from a tampered backup. Each must fail closed.

You write: findings with severity, one file and line per finding, and a clear
pass or fail on the gate. Effort estimate: one to two days. Dino confirms fee
or reciprocity before you start.

## Packet contents

1. Commit hash and `intra version` output.
2. `./scripts/verify.sh` full output with `VERIFY PASS`.
3. `./scripts/pilot-drills.sh` full output with `PILOT DRILLS PASS`.
4. The [security review checklist](security-review.md), blank findings table.
5. The [service selection record](pilot-service-candidates.md) for the chosen service.
6. The approval and retention defaults, signed by Dino.
7. Known limitations list from [implementation status](../IMPLEMENTATION-STATUS.md).

## Review routes (Dino chooses one)

| Route | Who | Cost shape | Risk |
|---|---|---|---|
| A — independent peer engineer | A named engineer outside this project | Reciprocity or small fee | Availability varies |
| B — external security consultant | A hired reviewer | Cash fee, written report | Cost, slower start |
| C — community review | A public call for review on the repo | Free, slow, uneven depth | May never land |

Decision recorded: route C, community review (Dino, 2026-11-18).
The public call goes out only after Dino approves its text.
The gate condition stands: one named reviewer, full checklist, signed findings.

## Community review call — draft, not posted

Post as a GitHub issue on `meschinidino/intra` after Dino approves this text.

Title: Security review wanted — default-deny access layer, checklist provided

Intra is a founder-built, default-deny access layer (Go, local lab at v0.1.0).
It needs one independent security review before any real service pilot.
The reviewer must not have built the guard. A full checklist, test commands,
and a findings table ship in `docs/pilot/security-review.md`.
The gate: no open Critical or High finding. Reply here if you will take it
as the named reviewer. The review closes only with a signed findings table.

## After the review

Record findings in the checklist table. Fix or accept each in writing.
A Critical or High finding stays open until the reviewer confirms the fix.
