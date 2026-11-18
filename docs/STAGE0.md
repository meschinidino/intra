# Stage 0 picks

Status: settled for the local build on 2026-11-18.

## Picks

| Need | Pick | Why | Rejected for now |
|---|---|---|---|
| Mesh target | Tailscale | Device identity, tags, ACL export, fast Mac enrollment | Plain WireGuard (more key work), NetBird (second choice) |
| Lab identity substitute | Local CA mTLS in Docker | Tailscale is not installed on this Mac. Docker networks prove route denial. Certificates prove device identity shape. | Header-only identity, never accepted outside unit tests |
| HTTP guard | Custom Go guard | Per-action policy, generated config, CLI status, and test control in one small surface | Pomerium and oauth2-proxy need an identity provider this lab does not have |
| Postgres path | Broker-owned Go gateway | CEO D12. Broker holds ephemeral credentials and kills sessions | Direct mesh database access |
| Language | Go | Eng R3. Static binaries, stdlib networking, pgx | Node and Python runtimes in the local path |
| Database | PostgreSQL 16 in Docker | Native roles, session termination, local image present | SQLite for the sensitive-service proof |
| Policy source | One JSON policy file | D16. Compiles mesh notes, guard rules, and broker templates | Hand-edited per-layer rules |
| Audit | One Go audit queue writer | Eng R2. One hash order and one external anchor | Two chains |
| Guard host | This Mac, lab only | D11. Sleep risk accepted for lab and revisited at Stage 5 entry | Linux host for the first proof |
| Model use | None | D10. No local or hosted model in allow, draft, or approval paths | Hosted drafting helper |
| SaaS spike budget | USD 0 | Local provider simulators only in this build. Real tenant proof needs a new written budget | Paid test tenants |

## Trust notes

- Tailscale coordination is vendor-hosted. That trust is accepted for the target deployment and recorded here.
- The Docker lab does not prove Tailscale behavior. It proves Intra topology and identity logic with mTLS.
- No real secret, customer data, or production credential enters this build.
- Touch ID runs through a signed macOS helper when built. Automated tests use a fake approver behind the same interface and never claim biometric proof.
