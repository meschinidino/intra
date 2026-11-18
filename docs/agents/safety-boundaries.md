# Safety boundaries for agents

These rules bind coding agents and robot actors.

## Never do

- Never read or print real passwords, private keys, grant tokens, or
  database secrets.
- Never copy real credentials into tests, fixtures, screenshots, issues,
  prompts, or commit messages.
- Never run the lab against a real database, SaaS tenant, or third-party
  system without written owner consent.
- Never approve your own sensitive grant as a robot.
- Never bypass the guard, broker gateway, egress proxy, or policy check
  to make a test pass.
- Never weaken a denial test to match broken behavior.
- Never claim a simulator result as a real provider result.
- Never run drills with real secrets.

## Stop and report

Stop before you act when:

- The target service, owner, or data class is unclear.
- Audit is unavailable and the task needs a new sensitive grant.
- A certificate actor and grant actor do not match.
- A policy change needs an action outside the approved task.
- The only way forward needs a real secret or external account.
- A test failure suggests a live path around the guard.

Report the exact command, error, grant ID when safe, and last known good
state. Do not include tokens or key material.

## Allowed local actions

- Build and test with fake data.
- Reset `.intra/` demo state with `init-demo --force`.
- Start and stop the `intra-lab` Compose project.
- Request robot access inside an approved test template.
- Run read-only status, simulate, draft, replay, and drill commands.

## Human-only actions

- Approve sensitive grants outside an explicit test fixture.
- Change license, ownership, repository visibility, or release tags.
- Add real tenants, providers, or production services.
- Rotate or revoke long-lived human credentials.
- Publish releases or security statements.
