# Intra

**Operable containment for humans, services, and robots.**

Intra keeps services unreachable by default. Approved people, devices,
services, and robots receive scoped, expiring access through checked local
gateways. Every allow, deny, approval, and revoke enters a hash-chained
audit log.

[![License: Apache-2.0](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)
[![Version](https://img.shields.io/badge/version-0.1.0-informational.svg)](VERSION)

> Status: `v0.1.0`, local lab. Intra is not production-ready. Local proofs,
> simulations, and unproven external gates stay separate in
> [implementation status](docs/IMPLEMENTATION-STATUS.md).

## Why Intra exists

Small teams often swing between two bad states:

1. A VPN login grants broad standing reach across the internal network.
2. After a security scare, the team cuts production access for everyone and
   loses the read access needed to troubleshoot incidents.

Intra replaces that swing with default-deny routes and short-lived grants.
The goal is not to stop work or promise attack prevention. The goal is to
keep useful work moving while one breach, stolen key, or tricked robot
stays small.

Read the [vision](VISION.md), [mission](MISSION.md), and
[why Intra exists](docs/explanation/why-intra.md).

## What the local lab proves

- A Mac owner can read a fake sensitive Postgres service through the broker.
- A read-only database role denies writes at Postgres itself.
- Expiry and revoke terminate live sessions and drop the role.
- A robot identity cannot reach sensitive HTTP services without a grant.
- Direct, DNS, stale-key, and canary paths fail in the Docker lab.
- Audit failure blocks new grants. The audit chain can be anchored and
  verified outside the log file.
- Local guarded reads measured less than 50 ms of added p95 latency in the
  latest verification run.

See [verification honesty](#verification-honesty) before treating any lab
result as a production claim.

## Quick start

Requirements:

- Go 1.26 or later
- Docker with Compose
- macOS if you want to compile the Touch ID helper

Run the full local verification suite:

```sh
./scripts/verify.sh
```

Expected final line:

```text
VERIFY PASS
```

Start the lab and keep it running:

```sh
go build -o bin/intra ./cmd/intra
./bin/intra init-demo --data .intra --force \
  --normal http://lab-normal:18090 \
  --logs http://lab-logs:18090 \
  --metrics http://lab-metrics:18090
export INTRA_UID="$(id -u)" INTRA_GID="$(id -g)"
export LAB_VIEWER_TOKEN="$(cat .intra/viewer.token)"
docker compose -p intra-lab -f docker-compose.lab.yml up -d --build
./bin/intra status --data .intra
```

Stop the lab:

```sh
docker compose -p intra-lab -f docker-compose.lab.yml down -v
```

For the guided version, read the [getting-started tutorial](docs/getting-started.md).

## Core model

```text
Owner or robot
      |
      | mTLS identity
      v
    Guard ----------> HTTP services
      |
      | grant check
      v
    Broker ----------> Postgres with ephemeral read-only role
      |
      +--------------> Audit queue -> hash chain -> external anchor

Robot -> deterministic egress proxy -> approved hosts only
```

- **Mesh** carries identity and encrypted routes.
- **Guard** decides HTTP service and action access.
- **Broker** owns grants, approvals, Postgres sessions, and revocation.
- **Audit writer** is the only process that appends to the hash chain.
- **Policy** is one JSON source. Generated layer configs must never be
  hand-edited.

Read the [architecture](ARCHITECTURE.md) and
[security model](docs/explanation/security-model.md).

## Common commands

```text
intra version          print the current release version
intra init-demo        create demo policy, keys, certificates, and state
intra request          request a scoped grant
intra approve          approve a grant
intra revoke           revoke a grant and kill live sessions
intra grants           list grants without returning secrets
intra status           show guard, broker, audit, grants, and anchor age
intra anchor           write a signed audit-chain head
intra anchor-once      run one scheduled anchor pass with alerts
intra backup           write a checksummed store and audit backup
intra restore          verify and restore one backup
intra rebind-approver  rotate the approval key after device loss
intra verify-anchor    verify the audit chain against its anchor
intra replay           score recorded decisions against a candidate policy
intra simulate         list reachable actor-service-action paths
intra drills           run safe direct-path and canary drills
intra onboard          add one imported service to policy
intra issue-identity   mint a per-run client certificate and manifest
```

The full command surface lives in the [CLI reference](docs/reference/cli.md).

## Verification honesty

Proven locally:

- Grant lifecycle, duplicate suppression, role-scoped Postgres reads,
  write denial, live-session kill, audit pause, hash anchor checks,
  Docker network denial, incident reads, drills, replay, onboarding,
  and latency under the local 50 ms budget.

Simulated:

- Microsoft Entra ID and Google Workspace behavior.
- Mesh behavior through local mTLS certificates and Docker networks.
- Log and metric services, plus the dev approver inside automated lab runs.

Not proven:

- Real Tailscale behavior.
- A real automated Mac Touch ID prompt.
- Real SaaS tenant behavior.
- A real low-risk service pilot.
- Outside-user demand.
- Production attack resistance.

## Documentation

Start with the [documentation index](docs/README.md). It links the
tutorial, how-to guides, reference, explanations, agent instructions,
operations runbook, test plan, and governing design records.

Contributors and coding agents must read [AGENTS.md](AGENTS.md) first.

## Project records

- [Vision](VISION.md)
- [Mission](MISSION.md)
- [Roadmap](ROADMAP.md)
- [TODO list](TODOS.md)
- [Changelog](CHANGELOG.md)
- [Governing design](docs/designs/intra-design.md)
- [CEO plan](docs/designs/intra-ceo-plan.md)
- [Early superseded brief](DESIGN-BRIEF.md)

## Ownership and license

Copyright © 2026 Dino Meschini.

Intra is licensed under the Apache License, Version 2.0. See
[LICENSE](LICENSE) for the full text and [NOTICE](NOTICE) for project
attribution and trademark notes. Contributors keep copyright in their
contributions unless they agree otherwise in writing. See
[CONTRIBUTING](CONTRIBUTING.md).
