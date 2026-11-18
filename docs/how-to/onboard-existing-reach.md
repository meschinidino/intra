# How to onboard existing reach

Import one WireGuard or Tailscale reach file, inspect standing paths, and
move one low-risk service behind the guard.

## Prerequisites

- A running lab
- A WireGuard config or Tailscale policy export
- One named low-risk service and its upstream URL

## Steps

1. Inspect the imported reach without changing policy.

   ```sh
   ./bin/intra import tailscale ./tailscale-policy.json
   ```

   For WireGuard, replace `tailscale` with `wireguard` and pass the
   config file.

2. Pick one low-risk path. Do not convert every path at once.

3. Add that one service.

   ```sh
   ./bin/intra onboard \
     --data .intra \
     --kind tailscale \
     --file ./tailscale-policy.json \
     --service imported \
     --upstream http://lab-normal:18090
   ```

4. Restart the server.

   ```sh
   docker compose -p intra-lab -f docker-compose.lab.yml restart intra
   ```

5. Request, approve, and read through the guard. Then try the old direct
   path. The direct path must fail before you call the cut safe.

## Verification

- `compiled.json` changed with one new service.
- Guarded read succeeds only with the intended identity and grant.
- Direct read fails.
- Rollback still works by restoring the prior `policy.json`, recompiling,
  and restarting.

## Troubleshooting

- `service exists`: pick a new service name or edit the existing service
  by hand.
- Import shows zero paths: check that the export uses `acls` or `grants`
  with `src` and `dst`, or WireGuard `AllowedIPs` lines.
- Guarded read returns `503`: check audit health with `intra status`.
