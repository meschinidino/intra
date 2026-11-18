# Watch-mode data collection plan

Goal: at least 200 decisions on the named service across at least 7 days.

## Rules

1. Watch mode records decision metadata only.
2. No request bodies, query literals, results, tokens, or secrets enter the corpus.
3. The corpus names actor, device, service, action class, rule, result, and timestamp.
4. The founder reviews the corpus daily for canary or secret leakage.
5. `intra status` runs daily. Record anchor age, partial revokes, and late kills.

## Daily log

| Day | Date | Decisions | Allows | Denies | Anchor age (s) | Partial revokes | Late kills | Leak check clean? |
|---|---|---|---|---|---|---|---|---|
| 1 |  |  |  |  |  |  |  |  |
| 2 |  |  |  |  |  |  |  |  |
| 3 |  |  |  |  |  |  |  |  |
| 4 |  |  |  |  |  |  |  |  |
| 5 |  |  |  |  |  |  |  |  |
| 6 |  |  |  |  |  |  |  |  |
| 7 |  |  |  |  |  |  |  |  |

## Exit

- [ ] At least 200 decisions recorded.
- [ ] At least 7 days covered.
- [ ] Zero canary or secret values found in the corpus.
- [ ] Corpus handed to replay ([gates](replay-rollback-gates.md)).
