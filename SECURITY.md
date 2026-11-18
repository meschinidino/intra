# Security policy

## Project status

Intra v0.1.0 is a local laboratory implementation. It is not a production
security product. Do not place real secrets, customer data, payment data,
or production credentials in the lab.

## Supported versions

Security fixes target the latest `0.1.x` line until a later supported line
exists. Older snapshots receive no guaranteed backports.

## Report a vulnerability

Do not open a public issue for a live vulnerability, exploit chain, real
credential, or private user data.

Use GitHub private vulnerability reporting on the
`meschinidino/intra` repository when available. If that path is not
available, contact the repository owner through GitHub with the subject
`Intra security report` and ask for a private channel before sending
details.

Include only the minimum proof needed:

- Affected commit or version
- Component and entry point
- Preconditions
- Minimal reproduction with fake data
- Impact
- Suggested fix, if known

## Response goals

These are goals, not service-level agreements:

- First response: 7 days
- Triage decision: 14 days
- Fix or mitigation plan for accepted critical issues: 30 days

## Safe research rules

- Test only systems you own or have written permission to test.
- Use fake identities, fake secrets, and isolated lab networks.
- Do not test third-party SaaS tenants with project tooling.
- Stop if a test reaches real data or another person's system.
- Remove real tokens, hostnames, and personal data from reports.

## Security model

Read the [security model](docs/explanation/security-model.md) before
reporting an issue that may be an explicit non-goal.
