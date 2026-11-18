# How to run the lab

Run the full fake network locally and confirm every core service is
healthy.

## Prerequisites

- Go 1.26 or later
- Docker with Compose
- Ports `15432`, `18080` to `18085`, `18093`, and `18443` free on
  `127.0.0.1`

## Steps

1. Build the CLI.

   ```sh
   go build -o bin/intra ./cmd/intra
   ```

2. Create fresh demo state.

   ```sh
   ./bin/intra init-demo --data .intra --force \
     --normal http://lab-normal:18090 \
     --logs http://lab-logs:18090 \
     --metrics http://lab-metrics:18090
   ```

   `--force` replaces demo policy, keys, grants, and audit state.

3. Export the Compose user and lab viewer token.

   ```sh
   export INTRA_UID="$(id -u)" INTRA_GID="$(id -g)"
   export LAB_VIEWER_TOKEN="$(cat .intra/viewer.token)"
   ```

4. Start the lab.

   ```sh
   docker compose -p intra-lab -f docker-compose.lab.yml up -d --build
   ```

5. Check status.

   ```sh
   ./bin/intra status --data .intra
   ```

## Verification

Broker, audit, and guard each report `200`. Grants start at zero before
you make a request.

## Stop

Remove containers and lab volumes:

```sh
docker compose -p intra-lab -f docker-compose.lab.yml down -v
```

## Troubleshooting

- Port busy: stop the process using that port or change the matching
  Compose host port.
- Broker cannot read keys: rerun Step 2 and confirm `.intra/` belongs to
  your user.
- Postgres stays unhealthy: inspect
  `docker compose -p intra-lab -f docker-compose.lab.yml logs postgres`.
- Guard returns `403`: check the client certificate actor with
  `openssl x509 -in .intra/certs/owner.crt -noout -subject`.
