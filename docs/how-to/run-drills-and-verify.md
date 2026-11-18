# How to run drills and verify

Prove the current local build from unit tests through the Docker lab.

## Prerequisites

- Go 1.26 or later
- Docker with Compose
- Swift compiler on macOS for the Touch ID helper compile check

## Steps

1. Run fast checks during development.

   ```sh
   go vet ./...
   go test ./...
   ```

2. Run the race-enabled Postgres gateway test after broker, gateway,
   audit, or SQL changes.

   ```sh
   ./scripts/pg-integration.sh
   ```

3. Run the full gate before you call a change done.

   ```sh
   ./scripts/verify.sh
   ```

4. Read every `FAIL` line before the final status. Fix the code or the
   wrong test expectation. Do not weaken a security assertion to pass.

## Run drills only

Create a target file such as:

```json
[
  {
    "name": "attacker-sink",
    "url": "http://127.0.0.1:18093/hits",
    "sensitive": false,
    "owner": "founder"
  }
]
```

Then run:

```sh
./bin/intra drills --config ./drill-targets.json
```

A pass means the target denies the drill identity or has no route. The
report blocks when any sensitive target passes traffic or the score goes
above 5 percent.

## Verification

`verify.sh` ends with `VERIFY PASS`. The Docker lab script ends with
`LAB E2E PASS`. The Postgres script ends with a Go test `PASS`.

## Troubleshooting

- Port already in use: run
  `docker compose -p intra-lab -f docker-compose.lab.yml down -v`, then
  check the named port.
- Audit writer already active: stop the other `intrad` using the same
  data directory.
- Postgres image pull fails: check Docker network access and rerun. The
  script uses `postgres:16-alpine`.
- Race failure: keep the full race stack. Fix ownership or cancellation,
  not the test timing.
