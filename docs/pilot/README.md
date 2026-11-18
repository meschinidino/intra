# Pilot packages

These packages prepare external gates. They create no evidence by themselves.
Dino supplies people, accounts, decisions, and manual actions.

## Demand gate

- [Interview script](interview-script.md)
- [Recruiting message](recruiting-message.md)
- [Evidence tracker](evidence-tracker.md)
- [Ranking sheet](ranking-sheet.md)
- [Observation template](observation-template.md)

## Stage 5 pilot

- [Service selection checklist](service-selection.md)
- [Owner consent template](owner-consent.md)
- [Security review checklist](security-review.md)
- [Watch-mode data plan](watch-mode-plan.md)
- [Replay and rollback gates](replay-rollback-gates.md)

## Decision packets (need Dino)

- [Outreach packet](outreach-packet.md): ranked slots, pool messages, booking text. Not sent.
- [Pilot service candidates](pilot-service-candidates.md): three ranked local candidates.
- [Security review packet](security-review-packet.md): reviewer brief and routes.
- [Approval and retention defaults](approval-retention-defaults.md): draft decisions.
- [Manual run sheets](manual-run-sheets.md): Tailscale proof and Touch ID lane.

## Local drills before any pilot

Run `./scripts/pilot-drills.sh`. It proves backup, restore, approver
rebinding, policy rollback, restart, and release-artifact checks on fake
state. It proves nothing about a real service.
