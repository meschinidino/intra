# Pilot service candidates — local inventory

Method: read-only listing of local projects on 2026-11-18, revised after
Dino rejected static frontends. A frontend proves no sensitive read, no
database denial, and no session kill. Only backend services rank here.
I read directory names and README heads only. I read no secret, env, or
credential file. No candidate is consented yet. Dino chooses.

Excluded without ranking: `tes-um` and `hapi` under `~/Documents`.
They are employer or client systems. Dino cannot consent for them.
Static sites (Vonder, portfolio) also leave the list for the reason above.

## Rank 1 — Café Orders (demo_andes_tech)

- Path: `~/Documents/demo_andes_tech`. Next.js, Prisma, Postgres in Compose.
- Owner: Dino Meschini (local git history).
- Data class: café order data from a demo build and seed flows. No payment surface found in the README.
- Dependencies: Node, Postgres container, Prisma migrations.
- Rollback path: stop the service container, or return its route to watch mode. The database stays untouched by rollback.
- Why it qualifies: a real Postgres read surface matches the Intra wedge. Demo data keeps the risk low.
- Gap: confirm no live café deployment uses this build before consent. If one does, it drops off the list.

## Rank 2 — Antesala, seeded instance only

- Path: `~/Documents/antesala`. Go server, PostgreSQL, WhatsApp gateway.
- Owner: Dino Meschini (his GitHub repo, with collaborators).
- Data class: clinical intake. On real data this is patient personal data and it fails low-risk at once.
- Dependencies: Go, Postgres, WAHA container, model providers.
- Rollback path: stop the seeded instance. It shares nothing with any real clinic deployment.
- Why it qualifies: the strongest wedge match in the inventory, with Postgres at the center.
- Gap: it qualifies only as a fresh instance with fake seed patients. The real service never enters the pilot.

## Rank 3 — Oravela HR app, fake tenant only

- Path: `~/Documents/hrapp`. Single Go binary, Postgres, row-level security.
- Owner: Dino Meschini (his GitHub repo).
- Data class: employee HR data in a multi-tenant design. Real tenants would fail low-risk.
- Dependencies: Docker Compose, Postgres, migrations, Caddy.
- Rollback path: delete the fake tenant, or return the route to watch mode.
- Why it qualifies: a Postgres backend with a real authorization layer to observe.
- Gap: it qualifies only with one fake tenant and invented employees. No real company data.

## Decision needed

Pick rank 1 for lowest data risk. Pick rank 2 for the strongest wedge proof.
Both need the fake-data confirmation in writing before the Stage 5 checklist opens.
