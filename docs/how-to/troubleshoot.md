# How to troubleshoot

Start with one status view, then follow the first failing boundary.

## First checks

```sh
./bin/intra status --data .intra
docker compose -p intra-lab -f docker-compose.lab.yml ps
docker compose -p intra-lab -f docker-compose.lab.yml logs --tail=80 intra
```

## Broker does not answer

- Confirm the CLI uses `https://127.0.0.1:18081` when certificates exist.
- Confirm `.intra/admin.token` exists and matches the server data dir.
- Confirm `.intra/certs/owner.crt` and `owner.key` exist.
- Check that no older lab still owns port `18081`.

## Grant stays requested

- Check the approver name matches a human policy actor.
- Check the approval signature came from the current `approval.key`.
- For sensitive grants, confirm Touch ID passed or the explicit dev flag
  was used in a demo-only run.
- Check audit health. Approval cannot finish while audit is paused.

## Guard returns 403

- Read the exact reason in the response body.
- Check certificate common name format: `actor/device`.
- Check actor, device, service, and action against policy.
- For grant access, check grant state, expiry, and token.
- For robot runs, check manifest expiry when manifest mode is enabled.

## Postgres connection closes

- Confirm the grant is `active`, not `requested`, `issuing`, or revoked.
- Confirm the database user equals the grant ID.
- Confirm the client certificate actor and device match the grant.
- Confirm the client requests TLS when the gateway has a TLS config.
- Check `intra` logs for `pggateway backend connect`.

## Audit verification fails

- Stop every writer that uses the same data directory.
- Run `./bin/intra verify-anchor --data .intra`.
- Do not edit `audit.jsonl` by hand to force a pass.
- Restore the log and grant store from a known backup, then verify again.

## Docker lab will not start

- Run `docker compose -p intra-lab -f docker-compose.lab.yml down -v`.
- Re-run `init-demo --force` to recreate demo-only state.
- Confirm `LAB_VIEWER_TOKEN` came from the new `.intra/viewer.token`.
- Confirm host ports `15432`, `18080` to `18085`, `18093`, and `18443`
  are free.

## Still stuck

Collect the command, exact error, `intra status` output, and the last 80
`intra` log lines. Remove tokens, keys, and grant secrets before sharing.
