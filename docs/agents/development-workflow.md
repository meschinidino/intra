# Development workflow for agents

## Before editing

1. Run `git status --short`.
2. Read the governing design section for the behavior.
3. Name the smallest public behavior your change will alter.
4. Find the existing test nearest that behavior.
5. Add or update the focused test before broad refactors.

## Local loops

Fast package loop:

```sh
go test ./internal/policy ./internal/broker ./internal/guard
```

Database loop:

```sh
./scripts/pg-integration.sh
```

Full gate:

```sh
./scripts/verify.sh
```

The full gate rebuilds `.intra/` demo state. It removes fake grants and
audit history from prior demo runs. It never needs real credentials.

## Change order

1. Update policy or domain types.
2. Update broker, guard, gateway, or CLI behavior.
3. Add focused unit tests.
4. Add integration proof when a network or database path changes.
5. Update reference docs and examples.
6. Update implementation status only when proof level changes.
7. Add a changelog entry for user-visible behavior.

## Generated files

- Edit `policy.json` source, then run `intra compile`.
- Never hand-edit `compiled.json`.
- Do not commit `.intra/`, `bin/`, local certificates, or local tokens.
- Docker images are build output, not source.

## Review rules

- A green unit test does not prove a network denial.
- A fake service pass does not prove a real service class.
- A simulator pass does not prove a real SaaS tenant.
- A compiled Touch ID helper does not prove a biometric prompt ran.
- A direct database pass does not prove the broker gateway path.

## Git

- Stage files by name.
- Check `git diff --cached` before any requested commit.
- Never use `git add -A` in a mixed tree.
- Do not commit or push without an explicit maintainer request.
- Leave unrelated changes untouched and name them in your report.
