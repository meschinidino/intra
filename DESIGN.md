# Design

Intra has no committed web dashboard in v0.1.0. Its product design is the
CLI, its errors, its status output, and the words used around access.

## Product principles

1. **Calm by default.** The normal answer is no route and no grant. The
   interface should make that state feel controlled, not broken.
2. **Every allow has a reason.** A grant shows actor, device, service,
   action, reason, approver, and expiry in one view.
3. **Time is visible.** Expiry and revoke state always sit next to access.
   Hidden permanent access is a design failure.
4. **Failure says the next move.** Each denial names the missing identity,
   grant, approval, service, or healthy component.
5. **Proof beats theater.** Status shows measured health, late kills,
   anchor age, false blocks, and drill score. It does not show a generic
   “secure” badge.
6. **One vocabulary.** Use actor, device, service, action, grant, revoke,
   audit, and drill consistently across CLI, docs, and logs.

## CLI experience

- Commands use verbs: `request`, `approve`, `revoke`, `grants`,
  `status`, `anchor`, `verify-anchor`, `drills`, and `simulate`.
- JSON output stays stable for scripts and agents.
- Human approval shows the exact grant scope before Touch ID.
- Duplicate requests return state without returning the token again.
- Development-only approval must say so in the flag and output path.
- No command prints database passwords, grant secrets, or private keys.
- Long operations state the gate they wait on: approval, audit, database,
  guard, or identity.

## States

| State | User meaning | Required display |
|---|---|---|
| `requested` | Waiting for approval | Actor, scope, expiry request |
| `issuing` | Broker is creating access | Grant ID and service |
| `active` | Access is live | Time left and session count |
| `revoking` | Access is already unusable, cleanup runs | Kill progress |
| `partial_revoke` | Some service cleanup failed | Failed service and next action |
| `revoke_escalated` | Three kill attempts failed | Operator action required |
| `expired` | Time ended and sessions closed | Expiry time |
| `revoked` | Human or system closed access | Actor and time |
| `denied` | Policy or approval refused access | Exact reason |

## Visual direction for future UI

- Use a quiet operations surface, not a marketing dashboard.
- Lead with live reach, active grants, and failed safety checks.
- Keep red for failed containment only. Denied routine access is neutral.
- Show time left as text and bar together.
- Make every sensitive action reachable by keyboard.
- Keep tables readable at terminal width before adding charts.
- Never hide simulated data behind production labels.

## Brand

- Name: Intra.
- Voice: direct, technical, calm, and short.
- Promise: operable containment, not attack prevention.
- Signature idea: services stay dark until a named actor opens a timed,
  audited path.

## Design records

- [Governing design](docs/designs/intra-design.md)
- [CEO plan](docs/designs/intra-ceo-plan.md)
- [Product design notes](docs/design/product-design.md)
- [Design decisions](docs/explanation/design-decisions.md)
