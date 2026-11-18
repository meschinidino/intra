# Pilot service candidates — local inventory

Method: read-only listing of `~/Documents/personal-projects` on 2026-11-18.
I read directory names and README heads only. I read no secret, env, or
credential file. Owner for all three: Dino Meschini.
None of these is a consented pilot yet. Dino chooses. The Stage 5 checklist
in [service selection](service-selection.md) still applies in full.

## Rank 1 — Vonder website

- Path: `~/Documents/personal-projects/websites/vonder` (`web/` Astro app, `site/` static copy).
- Owner: Dino Meschini.
- Data class: public marketing content. No customer, personal, or payment data found in the listing.
- Dependencies: Astro, npm, Cloudflare through `wrangler.jsonc`. No database.
- Rollback path: redeploy the prior `dist` output, or keep the service in watch mode on a local copy first.
- Why it qualifies: static HTTP service, no data store, instant rollback, real deployment surface.
- Gap: production URL and current traffic source are unconfirmed. Dino confirms before any consent step.

## Rank 2 — Personal portfolio site

- Path: `~/Documents/personal-site` (static `index.html` with `dist/` and `dist-pages/`).
- Owner: Dino Meschini.
- Data class: public portfolio content. No customer, personal, or payment data found in the listing.
- Dependencies: static hosting only. Heavier local clutter (`node_modules`) but no runtime service.
- Rollback path: republish the prior static output.
- Why it qualifies: same static profile as rank 1 with fewer moving parts.
- Gap: it proves less than Vonder because it has no app framework or deploy config to observe.

## Rank 3 — Unihero study app

- Path: `~/Documents/personal-projects/apps/unihero/app` (single local `index.html`).
- Owner: Dino Meschini, single user.
- Data class: Dino's own study records. Personal data, but only his own. No third-party data.
- Dependencies: browser only. No server, no network service.
- Rollback path: stop serving the file. Nothing external changes.
- Why it qualifies: watch mode can observe one real daily workflow with zero outside exposure.
- Gap: it is local-only, so it proves no remote route or real session kill. It can support the demand observation before it supports Stage 5.

## Not candidates

- `startup/agent-wallet`: design brief only, no code, and its topic is real money. It fails low-risk by subject.
- `apps/tinspire-loader`: USB tool, no network service to place behind a guard.
- `portfolio/model-lobotomy*`: model and data stores, heavier surface, unclear data classes. It needs its own classification first.
