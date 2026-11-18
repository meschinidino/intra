# How to add a service

Add an HTTP service behind the guard with one owner, one read rule, and
one denial test first.

## Prerequisites

- A running lab
- The upstream service reachable from the guard network
- A named service owner

## Steps

1. Add the service to `.intra/policy.json`, or edit
   [the source policy in version control](../reference/policy-schema.md#example).

   Set a stable name, kind, owner, sensitivity, upstream, and offered
   actions. Add route rules when method alone cannot map the action.

2. Add one narrow rule.

   Start with one actor, one action, and `grant` access for sensitive
   services. Do not start with static sensitive access.

3. Compile and review the generated diff.

   ```sh
   ./bin/intra compile --policy .intra/policy.json --out .intra/compiled.json
   ```

4. Restart the server so it loads the new policy.

   ```sh
   docker compose -p intra-lab -f docker-compose.lab.yml restart intra
   ```

5. Test three paths.

   ```text
   No identity -> 403
   Wrong actor -> 403
   Approved grant -> expected upstream response
   ```

6. Test direct access around the guard.

   The upstream must have no public host port, or it must reject callers
   without the guard identity.

## Verification

Run the full gate after the manual checks:

```sh
./scripts/verify.sh
```

## Troubleshooting

- `unknown service`: the guard did not load the edited policy. Restart
  the `intra` service.
- Policy will not load: run `intra compile`. It prints the first schema
  or reference error.
- Grant returns `403`: inspect actor, device, service, action, and grant
  expiry in `intra grants`.
- Direct upstream still answers: remove its host port or add its own
  caller check before calling the service protected.
