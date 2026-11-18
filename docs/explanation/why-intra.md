# Why Intra exists

## The problem

Small teams often give a VPN login broad reach. One copied config, one
sick laptop, or one tricked robot can then reach services it never needs.

After a scare, teams often swing to the other extreme. They cut
production access for everyone. Incidents then take longer because even
read-only database checks become impossible.

Attack automation makes that swing worse. The FBI warned in November
2026 that ARTEX-assisted intrusions can reach privileged access in under
three hours. Security teams cannot rely on slow manual containment after
an attacker or robot is already inside.

## The approach

Intra keeps services dark by default. It opens one named path only after
identity, device, service, action, reason, and time checks pass.

- Standing reach becomes a short grant.
- Database access becomes an expiring role, not a shared password.
- Revoke closes live sessions, not only future logins.
- Audit records hold decision metadata, never service payloads.
- Safe drills keep asking whether closed paths stayed closed.

## Why not only a VPN replacement?

A VPN replacement can improve entry control. It does not by itself solve
timed database roles, incident bundles, robot task scopes, SaaS grants,
or post-entry reach between services.

Intra composes those tools instead of replacing each one blindly. A mesh
can carry identity. A guard can check HTTP paths. A broker can own grants
and database sessions.

## Why not only an agent egress proxy?

CrabTrap-style egress control limits what robots send to external APIs.
It does not decide which internal services a human, service, or robot may
reach. Intra borrows observe, draft, and replay patterns from that model,
but keeps deterministic rules in every allow path.

## Trade-offs

- More moving parts can fail than a plain VPN.
- Approval adds seconds before sensitive work.
- Fail-closed behavior can block incident work when the guard is down.
- Local certificates, keys, and policy need careful operations.
- A compromised approved Mac during a live grant remains dangerous.

## Alternatives considered

- **Break-glass database demo only.** Faster proof, narrower product.
  The founder rejected it as the main path.
- **Standing-reach audit service.** Strong evidence hunt, no product
  control at the end. Its interviews remain in the roadmap.
- **Hosted control plane first.** Easier updates, weaker local trust
  story, and more account scope before demand proof.
- **LLM policy judge.** Rejected for allow decisions. It adds probability
  and prompt risk to a boundary that needs repeatable answers.
