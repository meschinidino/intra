# Outreach packet — decision ready, not sent

Status: no person has been contacted. Dino approves the exact list and text first.
This packet invents no names. Each slot names a profile and a channel.
Dino fills the name column from his own network.

## Reachable pools

| Pool | Why reachable | Public channel | Buyer fit |
|---|---|---|---|
| Former colleagues outside the example employer | Dino worked with them directly | Direct message, email | High if they own production |
| Local developer and infra communities | Dino can attend or post once | Meetup, community Slack or Discord, LinkedIn | Medium |
| Open-source maintainers who run production | Public repos and contact pages | GitHub profile contact, project community channel | Medium |
| Small-team or MSP contacts | Dino's own network | Direct message, email | High if they run client production |
| Agent builders | Demand-gate mix allows at most two | Community channel, GitHub | Low for buyer profile |

## Ranked candidate slots

Fill `Name` before approval. Rank follows buyer fit first, reach second.

| Rank | Slot | Profile that must match | Pool | Name (Dino fills) |
|---|---|---|---|---|
| 1 | Tech lead, small team, owns incidents | Owns production access decisions | Former colleagues |  |
| 2 | Infra owner, small team | Runs VPN or bastion today | Former colleagues |  |
| 3 | MSP operator | Runs production for clients | Small-team or MSP contacts |  |
| 4 | Tech lead, second team | Owns one blunt access cut story | Local communities |  |
| 5 | Infra owner, second team | Owns incident database reads | Local communities |  |
| 6 | Founder-engineer with production | Approves own emergency access | Former colleagues |  |
| 7 | OSS maintainer with production | Runs a public service alone or in a small team | Open-source maintainers |  |
| 8 | MSP owner | Decides client access policy | Small-team or MSP contacts |  |
| 9 | Agent builder A | Runs agents against internal services | Agent builders |  |
| 10 | Agent builder B | Runs agents against internal services | Agent builders |  |

Gate mix check: slots 1 to 8 give at least three tech leads or infra owners.
Slots 9 and 10 stay the only agent builders.

## Messages per pool

Dino sends. No agent sends. Each message names no product before the incident questions.

### Former colleagues — initial

Hi [name]. I am researching how small teams handle production access during incidents. I need 30 minutes, not a sales call. I will ask about your last incident, your current access path, and where access blocked you. I record notes only with your consent. Can we talk this week?

### Former colleagues — follow-up, five days later

Hi [name]. One follow-up only. The call stays 30 minutes. Your incident experience is the evidence I need, even if your current setup works well. Does [time A] or [time B] work?

### Local communities — initial (one post or direct message)

I am interviewing tech leads and infra owners about production access during incidents. The call takes 30 minutes. I ask about one real incident and one access delay. No pitch. If you own production access, I would value your account. Reply here and I will send times.

### Local communities — follow-up, five days later

Still looking for two more infra owners for 30-minute incident-access interviews this week. No pitch, notes only with consent. If you ran one incident this year, your account counts.

### Open-source maintainers — initial

Hi [name]. I maintain an access-control project and I am interviewing people who run production. Your project runs a real service, so your access path matters to the research. May I take 30 minutes on your last incident and where access helped or blocked you? Notes only with consent.

### Open-source maintainers — follow-up, five days later

Hi [name]. Closing my interview list this week. The call stays 30 minutes and covers one incident only. Does [time A] or [time B] work?

### Small-team or MSP contacts — initial

Hi [name]. I am researching production access during incidents for small teams and MSPs. I need 30 minutes with the person who owns access decisions. I ask about one incident, the current access path, and one delay. No pitch. Who on your side owns that?

### Small-team or MSP contacts — follow-up, five days later

Hi [name]. One follow-up. Thirty minutes, one incident, no pitch. If another person owns access decisions, an introduction counts as help.

### Agent builders — initial

Hi [name]. I am researching how teams grant agents access to internal services. I need 30 minutes on one real agent run: what it could reach, and what blocked it. No pitch. Notes only with consent.

### Agent builders — follow-up, five days later

Hi [name]. Closing the list this week. Thirty minutes on one agent run and its access path. Does [time A] or [time B] work?

## Booking questions (send after a yes)

1. Which two times this week suit you? I will confirm one.
2. May I record notes? Recording stays off unless you say yes.
3. May I quote you by role only, never by name or company?
4. After the call, may I ask for one introduction?

## Scheduling text

Subject: 30-minute interview — [date, time]

Thanks for agreeing. The call takes 30 minutes at [time, timezone].
I will ask about your last production incident and your current access path.
No preparation helps. No product pitch. Notes only with your consent.

## Tracker fields (add one row per slot)

| Field | Value |
|---|---|
| Slot rank and pool |  |
| Name, role, company size |  |
| Buyer profile? (tech lead or infra owner) |  |
| Channel and date sent |  |
| Follow-up date |  |
| Reply and booked time |  |
| Consent to notes and role-only quotes |  |
| Current production access path |  |
| One incident delay |  |
| Ranking sheet completed? |  |
| Walkthrough or access map agreed? |  |
| Introduction offered? |  |

Copy completed rows into the [evidence tracker](evidence-tracker.md).

## Approval gate

Dino approves three things before any send: the filled name list, the exact
message text, and the send order. Until then, send nothing.
