# Getting started with the Intra lab

You will run a fake network with one protected Postgres table. You will
request access, approve it, read the fake secret, fail a write, revoke the
grant, and verify the audit chain.

No real secret enters this tutorial.

## What you'll need

- macOS or Linux
- Go 1.26 or later
- Docker with Compose
- `curl`, `python3`, and a terminal

Check the tools:

```sh
go version
docker --version
docker compose version
```

## Step 1: Run the full verification

```sh
./scripts/verify.sh
```

This command runs unit tests first. It then proves the Postgres gateway
against a real local Postgres container. It then builds and exercises the
full Docker lab.

You should see `LAB E2E PASS`, then `VERIFY PASS`.

## Step 2: Create local demo state

```sh
go build -o bin/intra ./cmd/intra
./bin/intra init-demo --data .intra --force \
  --normal http://lab-normal:18090 \
  --logs http://lab-logs:18090 \
  --metrics http://lab-metrics:18090
```

This writes policy, local certificates, local keys, and empty runtime
stores under `.intra/`. The directory is ignored by git. Treat every key
inside it as a local secret.

## Step 3: Start the lab

```sh
export INTRA_UID="$(id -u)" INTRA_GID="$(id -g)"
export LAB_VIEWER_TOKEN="$(cat .intra/viewer.token)"
docker compose -p intra-lab -f docker-compose.lab.yml up -d --build
./bin/intra status --data .intra
```

Broker, audit, and guard should each report `200`.

## Step 4: Request and approve a grant

```sh
./bin/intra request \
  --request-key tutorial-1 \
  --actor owner \
  --device mac \
  --services sensitive-db \
  --actions read \
  --reason tutorial \
  --ttl 600
```

Save the returned grant `id` and one-time `token`. A duplicate request
with the same key never returns that token again.

For the local tutorial only, approve with the explicit development flag:

```sh
./bin/intra approve \
  --data .intra \
  --id <grant-id> \
  --approver owner \
  --dev-approve
```

`--dev-approve` does not prove Touch ID. It exists only for automated
local work.

## Step 5: Read through the gateway

```sh
docker run --rm --network intra-lab_edge \
  -v "$PWD/.intra/certs:/certs:ro" postgres:16-alpine \
  psql "host=intra port=15432 user=<grant-id> password=<grant-token> dbname=labdb sslmode=verify-ca sslrootcert=/certs/ca.crt sslcert=/certs/owner.crt sslkey=/certs/owner.key" \
  -Atc "SELECT value FROM secrets WHERE id=1"
```

Expected output:

```text
CANARY-SECRET-123
```

Now try a write. It must fail at the database role:

```sh
docker run --rm --network intra-lab_edge \
  -v "$PWD/.intra/certs:/certs:ro" postgres:16-alpine \
  psql "host=intra port=15432 user=<grant-id> password=<grant-token> dbname=labdb sslmode=verify-ca sslrootcert=/certs/ca.crt sslcert=/certs/owner.crt sslkey=/certs/owner.key" \
  -Atc "INSERT INTO secrets VALUES (2,'x')"
```

## Step 6: Revoke and verify

```sh
./bin/intra revoke --id <grant-id> --by owner
./bin/intra anchor --data .intra
./bin/intra verify-anchor --data .intra
./bin/intra status --data .intra
```

## What you built

You used a scoped grant instead of a permanent database password. The
broker held the real Postgres credential in memory. Revoke made the grant
unusable and dropped its database role.

Next, read [how grants work](how-to/request-access.md) and the
[security model](explanation/security-model.md).
