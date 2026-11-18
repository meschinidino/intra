# Intra — design brief

Status: SUPERSEDED on 2026-11-17 by the approved office-hours design in [docs/designs/intra-design.md](docs/designs/intra-design.md). That document governs where stage order or scope differs. This file stays as the pre-office-hours draft.

## Goal

Build a private network layer for one person, their devices, their
services, and their robots. A guard checks every request to a
sensitive service. The guard checks who asks, from which device,
and for which action. The default answer is no.

The system borrows three tricks from CrabTrap: watch traffic
before you block it, draft rules from observed traffic, and
replay old traffic against new rules. It adds what CrabTrap
skips: forced routing, device proof, and guards inside the
network.

## First user

One person in a personal lab. The main device is one Mac.
Robots run as local scripts or containers. No team features.
No support for unknown users.

## First slice

One sensitive fake service holds one fake secret. One normal
fake service holds public test data. One fake attacker server
records anything it receives.

The slice passes when all rows hold:

1. The Mac can read the sensitive service through the guard.
2. A robot can read the normal service through the guard.
3. A robot cannot reach the sensitive service.
4. A device without the Mac identity cannot reach it either.
5. No actor can bypass the guard with a direct connection.
6. The log records every decision and its rule.

## Non-goals for the first slice

1. No real secrets or real services.
2. No LLM judge. Rules stay deterministic.
3. No outbound internet guard. CrabTrap already covers that.
4. No write or delete actions. Read only.
5. No mobile devices, guests, or second users.

## Actors

| Actor | Identity | First-slice access |
|---|---|---|
| Owner on Mac | User login plus device key | Read both services |
| Robot | Own device key per robot | Read normal service only |
| Guard | Own service key | Reach both services |
| Anyone else | None | Nothing |

## Shape

Compose existing parts. Do not build crypto or mesh code.

1. A private mesh gives each device an identity and an
   encrypted route. Candidates: Tailscale, NetBird, plain
   WireGuard. Stage 0 picks one.
2. An identity-aware proxy sits in front of each fake service.
   It checks user, device, service, and action. Candidates:
   Pomerium, oauth2-proxy, Caddy with forward auth. Stage 0
   picks one.
3. Each fake service accepts traffic from the guard only.
   It has no public port and no direct mesh route.
4. Mesh access rules give a second layer. The robot has no
   route to the guard path for the sensitive service.
5. The guard writes one log record per decision. The record
   names actor, device, service, action, rule, and result.

```text
Mac ──► mesh ──► guard ──► sensitive service
Robot ─► mesh ──► guard ──► normal service
Robot ─► mesh ──► guard ──► sensitive service  DENIED
Anyone ─► direct route to a service           NO ROUTE
```

## Stages and test gates

### Stage 0 — pick the parts

1. Compare mesh candidates on device identity and rule support.
2. Compare proxy candidates on per-action rules and log quality.
3. Decide where the guard runs: Mac, home server, or small VPS.
4. Gate: one mesh, one proxy, one host, written here.

### Stage 1 — lab skeleton

1. Run both fake services and the attacker server in containers.
2. Put the fake secret `CANARY-SECRET-123` in the sensitive one.
3. Wire mesh and guard. Enforce nothing yet.
4. Gate: all test traffic appears in the guard log.

### Stage 2 — enforce the first slice

1. Turn on the rules from the Actors table.
2. Run normal-work tests for Mac and robot.
3. Run wrong-actor tests: robot, unknown device, expired key.
4. Run bypass tests: direct IP, other ports, no mesh, raw DNS.
5. Gate: zero canary leaks, zero open bypass paths, 100
   percent of decisions logged, under 1 percent false blocks
   on replayed normal traffic.

### Stage 3 — learn and replay

1. Record normal lab traffic in watch mode.
2. Draft rules from observed traffic, CrabTrap style.
3. Replay recorded traffic against each new rule set offline.
4. Gate: rule changes ship only after a clean replay.

### Stage 4 — robot task access

1. Give a robot read access tied to one task and one expiry.
2. Add a trick-the-robot test page on the attacker server.
3. Add a step-up check for dangerous owner actions.
4. Gate: expired task access fails, canary still never leaks.

### Stage 5 — first real service

1. Pick one low-risk real service.
2. Run it in watch mode first. Enforce after a clean replay.
3. Gate: owner approves the service and its rules in writing.

## What a lab pass proves

A pass proves guard logic, rule behavior, and log quality on
fake traffic. It does not prove a real network is safe. Real
safety needs the same tests after each real service joins.

## Open decisions

1. Mesh product and proxy product. Stage 0 decides.
2. Guard host. An always-on host beats the Mac, but the Mac
   is fine for the lab.
3. Login method for the owner: local account or an external
   identity provider.
4. Log store: plain files first, database only if files hurt.
