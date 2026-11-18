# Design decisions

This page names the choices that shape the code. The governing
[design record](../designs/intra-design.md) carries full context,
evidence, review history, and stage gates.

## Product decisions

| Decision | Choice | Why | Cost |
|---|---|---|---|
| Scope posture | Dream platform in hard stages | Founder chose full scope | Demand can lag build breadth |
| First user | Founder Mac lab | Fastest honest proof loop | No external buyer proof |
| First surface | Mac-first local | Matches the sensitive-from-my-Mac example | Guard sleep risk |
| Model use | No model in committed paths | Deterministic and private decisions | Less flexible rule drafting |
| Hosted control | None | Local trust story stays simple | Updates and support stay manual |
| Interface | CLI first | Exact, scriptable, agent-friendly | No visual operations view |

## Security decisions

| Decision | Choice | Why | Cost |
|---|---|---|---|
| Database path | Broker-owned gateway | One owner for role, secret, session, kill | Broker becomes a data-path target |
| Audit failure | Block new grants | Silent access breaks the core promise | Disk failure stops new work |
| Sensitive approval | Touch ID per grant | Binds approval to a fresh human act | Friction and macOS-only helper |
| Duplicate requests | One grant per request key | Revoke keeps one target | Clients must retain request keys |
| Policy source | One source, generated views | Drift becomes a visible diff | Small compiler and schema to own |
| Update posture | Test, then owner prompt | Guard behavior must never change silently | Fixes wait for owner action |
| Egress check | Validated-IP dial | Closes DNS rebinding between check and dial | More proxy code |
| Audit order | One queue writer | One sequence and one anchor | Local choke point |

## Engineering decisions

| Decision | Choice | Why | Cost |
|---|---|---|---|
| Language | Go | Static binaries and network code | Touch ID needs a Swift helper |
| Arrangement | Separate guard, broker, compiler, CLI | Clear failure borders | More local processes |
| Runtime stores | Owner-only files | Easy local inspection and restore | No team concurrency story |
| Tests | Attack suite on every change | Security claims need current proof | Slower CI and Docker needs |
| Latency budget | Added p95 under 50 ms in lab | Incident speed stays visible | Role and audit work must stay lean |

## Deferred decisions

- Real mesh rollout after Tailscale installation and route proof.
- Guard host after Stage 5 uptime evidence.
- Two-person approval after a named backup approver exists.
- Real SaaS providers after tenant budget and session-kill spikes.
- Protocols after Postgres, ordered by named service need.
- Hosted drafting help after a separate privacy and cost decision.

## Rejected as main paths

- Break-glass database demo only. It remains an inner proof slice.
- Standing-reach audit service. Its interviews remain product research.
- Direct database access over the mesh. It weakens live control.
- One binary arrangement. The founder chose separate local units.
- Header-only identity outside tests and the controlled lab.
