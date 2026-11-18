# Testing guide

## Test layers

| Layer | Command | Proves | Does not prove |
|---|---|---|---|
| Unit | `go test ./...` | Policy, broker, guard, audit, simulator logic | Docker routes or real database roles |
| Postgres integration | `./scripts/pg-integration.sh` | mTLS gateway, role denial, live kill with race detector | A real production database |
| Docker lab | `./scripts/lab-e2e.sh` | Guard, broker, audit, fake services, revoke, latency | Real mesh or real services |
| Full gate | `./scripts/verify.sh` | All local layers plus identity and onboarding | Demand, real tenants, or production safety |
| Manual lane | Mac Touch ID and sleep checks | Human prompt and host sleep behavior | Repeatability in CI |

## Which test to add

- Policy rule: table test in `internal/policy`.
- Grant state: broker test with the real store and audit server.
- HTTP denial: guard handler test with recorded upstream headers.
- Database behavior: integration test against real Postgres.
- Egress: validated-IP dial test plus a forged-identity denial.
- Audit: concurrent append, tamper, truncation, and anchor tests.
- CLI behavior: command test or a script assertion in `verify.sh`.

## Security assertions

Every new allow path needs all five checks:

1. Unknown identity denies.
2. Wrong actor or device denies.
3. Expired or missing grant denies where a grant is required.
4. Audit records the decision without secrets.
5. Direct upstream access fails or is explicitly outside the claim.

## Race and timing

Run race coverage for broker, gateway, and session-kill changes:

```sh
./scripts/pg-integration.sh
go test -race ./internal/broker ./internal/audit ./internal/guard
```

Use injected clocks where a package offers them. Real sleeps belong only
in narrow integration tests with a named deadline.

## Flake policy

- Fix the product race, ordering assumption, or test isolation first.
- Never raise a timeout as the only fix.
- Never delete a security assertion to make a suite green.
- Keep manual-only checks named as manual, never silent skips.
