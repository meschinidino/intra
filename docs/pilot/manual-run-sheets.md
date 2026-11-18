# Manual run sheets — Tailscale and Touch ID

An agent cannot run these lanes. Dino runs them. Each sheet names commands,
expected evidence, and Dino's time. Record results as observed, never as assumed.

## Sheet 1 — Real Tailscale proof

Time from Dino: about 90 minutes, plus install time.
Goal: repeat identity, route, revoke, and stale-device checks on real Tailscale.
Today the lab proves these only through local mTLS and Docker networks.

### Steps

1. Install Tailscale on the Mac and on the guard host. Enroll both by hand.
2. Record the tailnet name and both device names in the evidence block below.
3. Export the tailnet policy. Run `intra import tailscale <file>` and save the output.
4. Place one fake lab service behind the guard on the tailnet route only.
5. Read it from the Mac through the guard. Save the `intra status` output.
6. Try the direct tailnet route to the service, bypassing the guard. It must fail.
7. Revoke the Mac device in Tailscale. Repeat the read. It must fail.
8. Re-enroll, issue one grant, revoke it in Intra, and repeat the read. It must fail.
9. Let one device key go stale per Tailscale expiry settings, or remove and re-add the device. Repeat the read with the stale identity. It must fail.

### Expected evidence

- Import output with the tailnet paths.
- `intra status` before and after each revoke.
- Three failed reads: direct route, revoked device, revoked grant.
- Audit lines for each allow and deny, metadata only.

### Evidence record

| Field | Value |
|---|---|
| Date, tailnet, devices |  |
| Import output file |  |
| Direct route failed? |  |
| Device revoke failed the read? |  |
| Grant revoke failed the read? |  |
| Stale identity failed? |  |
| Minutes spent |  |

## Sheet 2 — Real Touch ID lane

Time from Dino: about 20 minutes.
Goal: one real biometric approval. The automated lab uses `--dev-approve`,
which proves no biometric event. Never claim Touch ID from the lab.

### Steps

1. Build both binaries.

   ```sh
   go build -o bin/intra ./cmd/intra
   swiftc -O -o bin/touchid-helper cmd/touchid-helper/main.swift
   ```

2. Start the lab without the development approver.

   ```sh
   export INTRA_TOUCHID_HELPER="$PWD/bin/touchid-helper"
   ```

3. Request one sensitive grant through the normal flow.
4. Approve it with `intra approve --id <grant> --touchid-helper bin/touchid-helper`. Do not pass `--dev-approve`.
5. Touch the sensor when the system prompt appears.
6. Read once through the grant, then revoke it.

### Expected evidence

- The macOS Touch ID prompt appeared and one touch approved it.
- The audit chain holds one `approve` event for the grant.
- One cancel or wrong-finger attempt denies the grant. Record that too.

### Evidence record

| Field | Value |
|---|---|
| Date, Mac model, macOS version |  |
| Prompt appeared? Touch approved? |  |
| Audit approve event seq |  |
| Cancel or wrong finger denied? |  |
| Minutes spent |  |

## Related manual lanes still owed

- Mac sleep and wake during an active grant: about 15 minutes.
- Fresh-user onboarding with two non-founder testers: about 30 minutes per tester.
- Guard restart during an active real grant: covered by the pilot rollback drill.
