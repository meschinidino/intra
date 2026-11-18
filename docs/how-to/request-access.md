# How to request, approve, and revoke access

Create one timed read grant, use it, then close it before expiry.

## Prerequisites

- A running lab. See [run the lab](run-the-lab.md).
- The `bin/intra` binary.
- Optional Touch ID helper in `bin/touchid-helper` on macOS.

## Steps

1. Request one scoped grant.

   ```sh
   ./bin/intra request \
     --request-key incident-42-db \
     --actor owner \
     --device mac \
     --services sensitive-db \
     --actions read \
     --reason "INC-42 database check" \
     --incident INC-42 \
     --ttl 600
   ```

   Save the returned grant `id` and one-time `token`. The same request
   key with the same shape returns state without the token.

2. Approve with Touch ID.

   ```sh
   ./bin/intra approve \
     --data .intra \
     --id <grant-id> \
     --approver owner \
     --touchid-helper ./bin/touchid-helper
   ```

   Development automation may use `--dev-approve`. That flag does not
   prove a biometric event.

3. Inspect the grant.

   ```sh
   ./bin/intra grants
   ./bin/intra status --data .intra
   ```

4. Revoke as soon as the work ends.

   ```sh
   ./bin/intra revoke --id <grant-id> --by owner
   ```

## Verification

The grant state becomes `revoked`. A later connection with the old token
fails. A late duplicate request with the same key returns the revoked
state and no token.

## Troubleshooting

- `Touch ID confirmation required`: pass a working helper. Check
  `INTRA_TOUCHID_HELPER` or `--touchid-helper`.
- `audit paused`: run `intra status`, fix the audit service, then submit
  a new request key after recovery.
- `request_key already used`: the key names a different request shape.
  Pick a new key.
- `ttl_seconds must be 1..N`: lower the request to the printed cap.
