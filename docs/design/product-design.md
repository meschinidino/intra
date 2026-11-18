# Product design

This document extends the root [DESIGN](../../DESIGN.md) with concrete
interface rules for the CLI now and any local operations surface later.

## Primary job

Help one operator answer four questions fast:

1. What can reach this service now?
2. Which grants are live, and when do they end?
3. What failed closed in the last hour?
4. What is the next safe cut?

## First screen

`intra status` is the home surface in v0.1.0. It must show:

- Guard, broker, and audit health
- Active and total grants
- Audit sequence head and anchor age
- Late session kills when later metrics exist

No landing page, hero copy, or decorative score belongs here.

## Request flow

1. Actor and device come from identity, not free text where possible.
2. Service and action use stable policy names.
3. Reason is required and plain text only.
4. TTL shows both requested time and hard cap.
5. The one-time token appears once. Later reads never reveal it.
6. Approval shows the exact scope before Touch ID.

## Error design

Every denial uses this order:

1. What was denied
2. Why it was denied
3. What safe action comes next

Examples:

```text
denied: no active grant
denied: Touch ID confirmation required
denied: template does not cover action write
```

Never print `403` alone in operator commands when a reason exists.

## Terminal rules

- Tables must stay readable at 100 columns.
- JSON remains available where commands already emit JSON.
- Color may support state, never carry state alone.
- Long IDs may truncate only when a full-ID command exists.
- Times use UTC in records and local time only when labeled.

## Future dashboard rules

- Lead with live reach and failed safety checks.
- Keep simulated and real evidence in separate labeled groups.
- Show expiry as text and progress together.
- Make revoke reachable in one action from every live grant.
- Require keyboard operation for every action.
- Do not add a generic security score without its formula beside it.
