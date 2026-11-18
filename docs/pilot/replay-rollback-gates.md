# Replay and rollback gates

Enforcement on the named service needs every gate below. A lab replay alone never authorizes real enforcement.

## Clean replay gate

- [ ] Corpus: at least 200 watch-mode decisions from the named service.
- [ ] Corpus spans at least 7 days.
- [ ] `intra replay` false-block rate is under 1 percent.
- [ ] Zero false blocks on the sensitive-read class.
- [ ] The service owner approved enforcement in writing after seeing replay results.

## Rollback gate

- [ ] Rollback target is watch-mode routing for this one service only.
- [ ] Rollback completes in under 15 minutes. Time it; do not estimate it.
- [ ] The owner approved the rollback window of at most 24 hours.
- [ ] The rollback action is logged in the audit chain.
- [ ] Rollback never restores broad standing reach.
- [ ] `./scripts/pilot-drills.sh` passed on the current build.

## Drill record

| Date | Build commit | Rollback start | Rollback end | Minutes | Result |
|---|---|---|---|---|---|
|  |  |  |  |  |  |
