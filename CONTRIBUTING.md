# Contributing to Intra

Thank you for helping build Intra. Small, tested changes land fastest.

## Before you start

1. Read [README](README.md).
2. Read [AGENTS](AGENTS.md).
3. Read the [governing design](docs/designs/intra-design.md).
4. Check [TODOS](TODOS.md) and open issues for duplicate work.
5. For a security flaw, follow [SECURITY](SECURITY.md). Do not open a
   public issue with exploit details or real credentials.

## Development setup

Requirements:

- Go 1.26 or later
- Docker with Compose
- macOS for the optional Touch ID helper build

Run the full local gate:

```sh
./scripts/verify.sh
```

Run the fast gate during development:

```sh
go vet ./...
go test ./...
```

Run the Postgres gateway integration test after broker, gateway, audit,
or SQL changes:

```sh
./scripts/pg-integration.sh
```

## Contribution rules

- Keep one behavior change per pull request.
- Add or update a focused Go test for every behavior change.
- Preserve every security invariant in [AGENTS](AGENTS.md).
- Never commit `.intra/`, `bin/`, keys, tokens, certificates, grant files,
  audit logs, or local databases.
- Use fake services and fake secrets only in tests and examples.
- Update docs in the same pull request when commands, config, behavior,
  or proof level changes.
- Add a changelog entry for user-visible changes.
- Name simulations as simulations wherever the claim appears.

## Pull request checklist

- [ ] The change matches the governing design or updates it.
- [ ] Focused tests cover the new or fixed behavior.
- [ ] `go vet ./...` passes.
- [ ] `go test ./...` passes.
- [ ] Relevant integration or lab scripts pass.
- [ ] No secret or generated runtime state is staged.
- [ ] Docs and changelog match the behavior.
- [ ] Screenshots or terminal output appear when CLI output changes.

## Commit style

Use short imperative subjects:

```text
broker: block grants when audit is unavailable
docs: add first-service onboarding guide
test: cover duplicate grant token handling
```

## Ownership

By contributing, you agree that your contribution is licensed under the
Apache License, Version 2.0. You keep copyright in your contribution
unless you agree otherwise in writing.
