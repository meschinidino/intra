# Agent quick start

This guide is for coding agents and robot operators. Human maintainers
should start with the root [README](../../README.md).

## First five minutes

1. Read [AGENTS](../../AGENTS.md).
2. Read the [governing design](../designs/intra-design.md).
3. Read [implementation status](../IMPLEMENTATION-STATUS.md).
4. Run `git status --short`.
5. Run the smallest relevant test before you edit behavior.

## Safe local proof

```sh
go vet ./...
go test ./...
./scripts/verify.sh
```

`verify.sh` rebuilds demo state, starts and stops Docker lab services,
and exercises fake data only. It must never receive real credentials.

## Robot actor setup

1. Ask an operator to run `intra init-demo` or use an approved data dir.
2. Mint one identity per run:

   ```sh
   ./bin/intra issue-identity \
     --data .intra \
     --actor robot \
     --run task-123 \
     --ttl-hours 2
   ```

3. Use the issued client certificate for guard or gateway calls.
4. Request only the service and action named in your task.
5. Use a stable request key per task attempt.
6. Revoke the grant when the task ends. If revoke fails, report the grant
   ID and exact error.

## Robot hard limits

- Never approve your own grant.
- Never create or edit a template.
- Never ask for a database password. The broker holds it.
- Never use another run's certificate, token, or request key.
- Never bypass the guard, broker gateway, or egress proxy.
- Stop when policy, audit, identity, or grant state is unclear.

## Where to work

| Task | Start here |
|---|---|
| Policy behavior | `internal/policy/` |
| Grant lifecycle | `internal/broker/` |
| HTTP access | `internal/guard/` |
| Postgres sessions | `internal/pggateway/` |
| Audit | `internal/audit/` |
| CLI | `cmd/intra/` |
| Server wiring | `cmd/intrad/` |
| Lab behavior | `scripts/` and `docker-compose.lab.yml` |

Read [development workflow](development-workflow.md) before a multi-file
change. Read [safety boundaries](safety-boundaries.md) before any auth,
grant, audit, egress, or database change.
