# Vision

## The world Intra wants

Every small team can keep its services unreachable by default without
making daily work impossible.

A person, device, service, vendor tool, or robot receives only the access
its current task needs. That access has a clear owner, a short life, and a
record. One stolen key, sick laptop, hostile service, or tricked robot
cannot roam the whole network.

## The twelve-month ideal

- A new service joins in minutes and has no standing inbound reach.
- A live map shows who can reach what now, not what a stale document says.
- Incident access opens in minutes, covers the needed read paths, and
  closes itself.
- One revoke kills the live sessions it can truthfully kill and names the
  classes it cannot.
- Safe drills keep testing the doors after launch. Fresh proof replaces a
  launch-day security claim.
- Robots hold per-run identity and task-scoped grants, never shared
  long-lived production keys.
- Vendor and SaaS grants sit in the same inventory as network paths.

## What success looks like

A small team can answer five questions at any moment:

1. Who or what can reach this service?
2. Why does that actor have access?
3. When does the access end?
4. What did the actor do with it?
5. What happens if that actor is compromised right now?

If the team cannot answer one question, Intra has not finished its job.

## Non-goals

- Intra does not promise to prevent every attack.
- Intra does not replace patching, backups, endpoint protection, or
  incident response.
- Intra does not need a new public internet topology.
- Intra does not put a language model in an allow decision.
- Intra does not start as a hosted control plane.

## Design promise

Default deny must stay usable. If safe access takes longer than an unsafe
shortcut, people and robots will route around the guard. Speed, clear
errors, and recoverable failure are security features.
