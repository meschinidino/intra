# Mission

## Mission statement

Build a local-first containment layer that lets small teams replace
standing broad network reach with scoped, expiring, audited access for
people, devices, services, and robots.

## Who we serve first

The first proven user is the founder's own Mac lab. The expected external
user is a tech lead or infrastructure owner at a small team who owns both
production incidents and access control.

That external buyer is still a hypothesis. Practitioner interviews gate
product claims beyond the lab.

## What we build

- Default-deny service reach through checked gateways.
- Short-lived grants tied to actor, device, service, action, reason, and
  expiry.
- Database-native read-only enforcement and fast session kill.
- One reviewable policy source for every enforcement layer.
- Metadata-only audit records with external integrity anchoring.
- Safe drills and replay tools that keep containment measurable.
- Setup that reaches a first safe service in minutes, not weeks.

## Operating principles

1. **Containment over promises.** Claim a smaller blast radius, never
   attack prevention.
2. **Identity before location.** A network address never proves who acts.
3. **Expiry by default.** Standing access needs a written reason.
4. **Audit or deny.** A sensitive grant without a durable record does not
   exist.
5. **Deterministic allow paths.** Models may help later outside the allow
   path. Rules decide access today.
6. **Fail closed, explain clearly.** Denials name the missing identity,
   grant, approval, or service state.
7. **Prove locally, label simulations.** Fake tenants and fake services
   never become real-world claims.
8. **Demand before breadth.** Interviews and observed work outrank more
   platform surface.

## Measures

- Time from request to approved incident read.
- Count of standing paths to sensitive services.
- Percentage of decisions with complete actor, device, rule, and result.
- Time from revoke to live-session death.
- False-block rate on replayed normal decisions.
- Time for a new operator to protect the first service.
