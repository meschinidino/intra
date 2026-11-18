#!/bin/sh
# Pilot drills: backup, restore, approver rebind, policy rollback, restart,
# and release-artifact verification on fake local state only.
set -eu
cd "$(dirname "$0")/.."
TMP=$(mktemp -d /tmp/intra-pilot-drills.XXXXXX)
trap 'rm -rf "$TMP"; kill ${INTRAD_PID:-} 2>/dev/null || true' EXIT

go build -o "$TMP/intra" ./cmd/intra
go build -o "$TMP/intrad" ./cmd/intrad
export INTRA_ANCHOR_KEY="$TMP/anchor.key"
export INTRA_DATA="$TMP/a"

echo "== init and anchor =="
"$TMP/intra" init-demo --data "$TMP/a" --force >/dev/null
"$TMP/intra" anchor --data "$TMP/a"
"$TMP/intra" verify-anchor --data "$TMP/a"

echo "== backup and clean restore =="
"$TMP/intra" backup --data "$TMP/a" --out "$TMP/backup.zip"
"$TMP/intra" restore --data "$TMP/b" --from "$TMP/backup.zip"
"$TMP/intra" verify-anchor --data "$TMP/b"
test ! -f "$TMP/b/approval.key" || { echo "FAIL backup carried approval key"; exit 1; }
echo "PASS backup/restore"

echo "== tampered restore refused =="
cp -R "$TMP/a" "$TMP/corrupt-src"
printf '{}\n' >> "$TMP/corrupt-src/audit.jsonl"
"$TMP/intra" backup --data "$TMP/corrupt-src" --out "$TMP/bad.zip"
if "$TMP/intra" restore --data "$TMP/corrupt" --from "$TMP/bad.zip" 2>/dev/null; then
  echo "FAIL restore accepted broken audit chain"; exit 1
fi
echo "PASS tampered restore refused"

echo "== approver rebind =="
cp "$TMP/a/approval.key" "$TMP/old.key"
if "$TMP/intra" rebind-approver --data "$TMP/b" 2>/dev/null; then
  echo "FAIL rebind without confirmation succeeded"; exit 1
fi
# Restore target has no approval key by design; copy the demo key in as the
# bound approver, then rebind it. This models a clean-host restore where the
# approver key arrives through the operator channel, not the backup.
cp "$TMP/a/approval.key" "$TMP/b/approval.key"
"$TMP/intra" rebind-approver --data "$TMP/b" --confirm-old-revoked
cmp -s "$TMP/old.key" "$TMP/b/approval.key" && { echo "FAIL key unchanged"; exit 1; }
test -f "$TMP/b/approver-rebind.json"
echo "PASS approver rebind"

echo "== policy rollback =="
cp "$TMP/a/policy.json" "$TMP/policy.good.json"
"$TMP/intra" restore --data "$TMP/c" --from "$TMP/backup.zip"
"$TMP/intra" compile --policy "$TMP/c/policy.json" --out "$TMP/c/compiled.json" >/dev/null
cmp "$TMP/policy.good.json" "$TMP/c/policy.json"
echo "PASS policy rollback restores prior policy and compiles"

echo "== restart drill =="
export INTRA_BROKER_ADDR="127.0.0.1:18181" INTRA_AUDIT_ADDR="127.0.0.1:18182"
# intra status locates the broker through INTRA_BROKER_URL (docs: environment
# variables) and the guard at the default TLS port 18443. Align both here.
export INTRA_BROKER_URL="https://127.0.0.1:18181"
export INTRA_GUARD_ADDR="127.0.0.1:18443" INTRA_EGRESS_ADDR="127.0.0.1:18183"
export INTRA_GUARD_HTTP_ADDR="127.0.0.1:18180" INTRA_ANCHOR_INTERVAL_SEC="1"
start_intrad() {
  "$TMP/intrad" all --data "$TMP/a" >"$TMP/intrad.log" 2>&1 &
  INTRAD_PID=$!
  i=0
  while [ $i -lt 50 ]; do
    if "$TMP/intra" status --data "$TMP/a" 2>/dev/null | grep -q '"broker": 200'; then return 0; fi
    i=$((i+1)); sleep 0.2
  done
  echo "FAIL intrad did not become healthy"; cat "$TMP/intrad.log"; exit 1
}
start_intrad
kill "$INTRAD_PID"; wait "$INTRAD_PID" 2>/dev/null || true
start_intrad
"$TMP/intra" status --data "$TMP/a" | grep -q '"anchor_stale"'
"$TMP/intra" status --data "$TMP/a" | grep -q '"guard": 200'
"$TMP/intra" status --data "$TMP/a" | grep -q '"audit": 200'
kill "$INTRAD_PID"; wait "$INTRAD_PID" 2>/dev/null || true
INTRAD_PID=""
echo "PASS restart drill (status healthy after restart, anchor fields present)"

echo "== release artifact verification =="
mkdir -p "$TMP/release"
cp "$TMP/intra" "$TMP/release/intra"
"$TMP/intra" release-checksums --dir "$TMP/release" --files intra --out "$TMP/release/SHA256SUMS"
"$TMP/intra" verify-release --dir "$TMP/release" --checksums "$TMP/release/SHA256SUMS"
printf 'x' >> "$TMP/release/intra"
if "$TMP/intra" verify-release --dir "$TMP/release" --checksums "$TMP/release/SHA256SUMS" 2>/dev/null; then
  echo "FAIL tampered artifact verified"; exit 1
fi
echo "PASS release artifact verification"

echo "PILOT DRILLS PASS"
