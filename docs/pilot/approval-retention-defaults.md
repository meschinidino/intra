# Approval and retention defaults — draft for Dino's decision

Status: draft. Nothing here is a recorded decision until Dino accepts it.

## Two-person approval — recommended default

1. Sensitive grants need two people: the requester and one approver who is not the requester.
2. Dino names one backup approver before the pilot starts.
3. If no approver answers, the grant waits. No silent auto-approval exists.
4. Robot grants never self-approve. That rule already holds in code.
5. Solo approval stays allowed for the fake lab only.

Pilot exception, only if Dino accepts it in writing: for the one named
low-risk service, Dino may approve his own grants during watch mode.
Enforcement on that service still needs the two-person rule or a renewed
written exception with an end date.

Why: the governing design requires a recorded decision at Stage 5 entry.
The pilot service is low-risk, but its data is real. The two-person rule
costs one named person and removes the largest single-person risk.

## Log retention — recommended default

1. Decision metadata retention: 90 days. This matches the design default.
2. Audit chain and anchors: keep for the full pilot plus 90 days.
3. Watch-mode corpus: delete 30 days after the replay decision.
4. Alerts file: keep with the audit chain.
5. No log holds bodies, literals, results, tokens, or secrets. That rule already holds in code.

Why: 90 days covers one incident review cycle. Longer storage grows the
metadata exposure without a named need.

## Decision record (Dino completes)

- Two-person rule accepted / changed: ____________ Date: ________
- Backup approver named: ____________
- Pilot solo exception accepted? End date: ____________
- Retention accepted / changed: ____________ Date: ________
