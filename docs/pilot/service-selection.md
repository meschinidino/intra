# Low-risk service selection checklist

Name exactly one service. Every box must pass before Stage 5 entry.

- [ ] The service holds no customer data.
- [ ] The service holds no personal data.
- [ ] The service holds no payment data.
- [ ] The service holds no production secrets.
- [ ] A named owner exists and consents in writing ([template](owner-consent.md)).
- [ ] The owner classified the service sensitive or not, with a reason.
- [ ] A native read-only control exists, or the service stays watch-only.
- [ ] Session kill is testable on this service class.
- [ ] Rollback to watch mode takes under 15 minutes ([gates](replay-rollback-gates.md)).
- [ ] Guard-host availability is decided: Mac-first with accepted downtime, or an always-on host.
- [ ] Two-person approval decision is recorded, or solo approval for this pilot only.
- [ ] Log retention is recorded. Default is 90 days.

Fail any box: pick another service or stop. Do not waive on paper.
