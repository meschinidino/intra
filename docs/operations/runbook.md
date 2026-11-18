# Operations runbook

This runbook covers the local lab only. It makes no production claim.

## Check health

```sh
./bin/intra status --data .intra
docker compose -p intra-lab -f docker-compose.lab.yml ps
docker compose -p intra-lab -f docker-compose.lab.yml logs --tail=80 intra
```

Healthy lab state: broker `200`, audit `200`, guard `200`, and an audit
head that moves after a new decision.

## Guard or broker down

1. Confirm the `intra` container state.
2. Read the last 80 log lines.
3. Check disk space and `.intra` ownership.
4. Restart only the `intra` service:

   ```sh
   docker compose -p intra-lab -f docker-compose.lab.yml restart intra
   ```

5. Run `intra status` again.
6. Verify active Postgres grants after restart. The broker should
   restore in-memory database secrets for still-active grants.

Network access must stay closed while the guard is down.

## Audit paused

Symptoms: new grants return `audit paused`, or status shows audit down.

1. Check disk space.
2. Run `intra verify-anchor --data .intra`.
3. Confirm only one `intrad` owns the audit writer lock.
4. Restart the `intra` service.
5. Do not delete `audit.jsonl` to force progress.

## Grant will not revoke

1. Read the grant state with `intra grants`.
2. `revoking`: wait for broker cleanup, then check again.
3. `partial_revoke`: inspect child service states and broker logs.
4. `revoke_escalated`: stop new grants. Kill the database role through a
   database owner session, then record the manual action outside Intra.
5. Re-run revoke only after the failed service is reachable.

## Anchor mismatch

1. Stop new grants.
2. Preserve `audit.jsonl`, `anchor.json`, and container logs.
3. Run `intra verify-anchor --data .intra`.
4. Compare the reported sequence with the last known good backup.
5. Restore the log and grant store together. Never restore one alone.

## Lost Mac or approver device

1. Remove the old device from the mesh or local certificate trust set.
2. Revoke active grants for that device.
3. Bind the new device from local console on the guard host.
4. Record the declared loss in the emergency or incident record.
5. Test one read grant, then revoke it.

## Rollback

1. Restore the prior `policy.json`.
2. Run `intra compile`.
3. Restart the `intra` service.
4. Run allow and direct-bypass smoke checks.
5. Keep the failed policy and audit records for review.

Rollback may restore watch-mode routing for one named service for at
most 24 hours with owner approval. It must never restore broad standing
reach automatically.

## Recovery drills

Before a real pilot, test:

- Grant-store and audit-log restore to a clean host
- Approver-device rebinding
- Guard restart during an active grant
- Audit disk-full behavior
- Postgres session kill after broker restart
