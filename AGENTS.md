# AGENTS.md

Instructions for coding agents working in this repository.

## Read first

Read these files in order before you change code or plans:

1. [README](README.md)
2. [Governing design](docs/designs/intra-design.md)
3. [Implementation status](docs/IMPLEMENTATION-STATUS.md)
4. [Architecture](ARCHITECTURE.md)
5. [Stage 0 picks](docs/STAGE0.md)
6. [Test plan](docs/test-plan.md)

The governing design wins where documents differ. The implementation
status file records what local code proves today. Do not promote a local
or simulated result into a production claim.

## Project shape

- Language and runtime: Go 1.26 or later.
- Operator CLI: `cmd/intra`.
- Local server: `cmd/intrad`.
- Fake lab services: `cmd/labservice`.
- macOS Touch ID helper: `cmd/touchid-helper`.
- Core packages: `internal/policy`, `internal/broker`,
  `internal/guard`, `internal/pggateway`, and `internal/audit`.
- Supporting packages: `internal/certs`, `internal/drill`,
  `internal/onboard`, `internal/providers`, and `internal/simulator`.
- Lab database setup: `lab/postgres/init.sql`.
- End-to-end scripts: `scripts/verify.sh`, `scripts/lab-e2e.sh`, and
  `scripts/pg-integration.sh`.

## Setup

```sh
go version
docker --version
docker compose version
./scripts/verify.sh
```

`verify.sh` is the full local gate. It runs unit tests, compiles the
Touch ID helper, runs race-enabled Postgres integration, runs the Docker
lab, issues one robot identity, and applies one onboarding import.

Start the persistent lab with the commands in
[getting started](docs/getting-started.md).

## Security invariants

Never weaken these rules to make a test pass:

1. Default deny wins. Unknown actor, service, action, identity, or grant
   state must deny.
2. Deterministic rules decide access. No model may enter an allow path.
3. Audit failure blocks new grants. Never create a silent grant.
4. Revoke flips a grant to a non-usable state before external kill calls.
5. Database grant secrets stay in broker memory only.
6. Duplicate request keys never return an existing token.
7. One audit writer owns hash order.
8. Logs hold metadata only. Never log bodies, query literals, results,
   tokens, passwords, or canary values.
9. The broker admin role must not gain direct reads on protected data.
10. Robot egress dials only validated global IP addresses.
11. Generated policy output comes from one policy source. Never hand-edit
    generated layer config.
12. TLS identity beats request-body identity when a client certificate is
    present.

If a change needs an exception, stop and update the governing design in
the same change. Name the new risk in plain language.

## Secrets and local state

- Never commit `.intra/`, `bin/`, keys, certificates, tokens, grant files,
  audit logs, or local databases.
- `.intra/` is generated local state. `intra init-demo --force` replaces
  demo keys, policy, grants, and audit state.
- The audit anchor key stays outside the repo at
  `~/.config/intra/anchor.key` unless `INTRA_ANCHOR_KEY` names another
  path.
- Lab passwords in Compose and SQL are development-only examples. Do not
  reuse them outside the fake lab.
- Before every commit, scan staged content for real credentials.

## How robot actors use Intra

- A robot run receives one short-lived client certificate and manifest
  from `intra issue-identity`.
- The certificate common name uses `actor/run-id`.
- Internal service access needs a matching policy rule or read template.
- Sensitive and write access needs human approval.
- A pre-approved template may issue only bounded read access inside its
  named services, actions, and TTL.
- Robots never create templates, edit templates, or approve grants.
- Internet access goes through the egress proxy with a live egress grant.
  Direct routes and direct DNS must fail.
- The robot must present its grant token to the broker gateway. It never
  receives the ephemeral Postgres password.

## Change rules

- Keep each change small and named by behavior.
- Add or update a focused test for every behavior change.
- Run the smallest proving test first, then the relevant package test.
- Run `./scripts/verify.sh` before claiming a cross-cutting change works.
- Use `-race` for Postgres gateway, broker concurrency, and session-kill
  changes.
- Update [CHANGELOG](CHANGELOG.md) for user-visible changes.
- Update [implementation status](docs/IMPLEMENTATION-STATUS.md) when proof
  level changes.
- Keep `VERSION`, `internal/version/version.go`, and the changelog in sync
  for releases.

## Documentation rules

- Write one idea per sentence.
- Keep instructions at 20 words or fewer where possible.
- Keep descriptions at 25 words or fewer where possible.
- Use active voice and one term per concept.
- Put warnings before explanations.
- Link every new document from [docs index](docs/README.md) and, when it
  is a top-level entry point, from the root [README](README.md).
- Mark simulated behavior as simulated in the same paragraph as the claim.

## Git rules

- Branch is `main` until a maintainer chooses another flow.
- Stage files by name. Never use `git add -A` for a mixed tree.
- Do not commit, tag, or push unless the user or maintainer asks.
- Do not rewrite history without an explicit request.
- Check `git status --short` before and after your work.
- Leave unrelated user files untouched.

## Definition of done

A change is done only when all true rows apply:

- The behavior matches the governing design or updates it.
- Focused tests pass.
- `go vet ./...` passes.
- Relevant integration or lab checks pass.
- Docs, status, and changelog match the real proof level.
- No secret or generated runtime state is staged.
