#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
export INTRA_UID="$(id -u)" INTRA_GID="$(id -g)"
mkdir -p bin
go build -o bin/intra ./cmd/intra
go build -o bin/intrad ./cmd/intrad
go build -o bin/labservice ./cmd/labservice

./bin/intra init-demo --data .intra --force \
  --normal http://lab-normal:18090 \
  --logs http://lab-logs:18090 \
  --metrics http://lab-metrics:18090
export LAB_VIEWER_TOKEN="$(cat .intra/viewer.token)"

COMPOSE="docker compose -p intra-lab -f docker-compose.lab.yml"
cleanup() { $COMPOSE down -v >/dev/null 2>&1 || true; }
trap cleanup EXIT
cleanup
$COMPOSE up -d --build

for i in $(seq 1 60); do
  if curl -fsS --cacert .intra/certs/ca.crt --cert .intra/certs/owner.crt --key .intra/certs/owner.key https://127.0.0.1:18081/health >/dev/null 2>&1; then break; fi
  sleep 1
done
curl -fsS --cacert .intra/certs/ca.crt --cert .intra/certs/owner.crt --key .intra/certs/owner.key https://127.0.0.1:18081/health >/dev/null
echo "PASS broker health"

code=$(curl -s -o /dev/null -w '%{http_code}' --cacert .intra/certs/ca.crt --cert .intra/certs/owner.crt --key .intra/certs/owner.key https://127.0.0.1:18443/svc/normal/)
[ "$code" = 200 ] || { echo "FAIL owner normal $code"; exit 1; }
echo "PASS owner static normal read"

code=$(curl -s -o /dev/null -w '%{http_code}' --cacert .intra/certs/ca.crt --cert .intra/certs/robot.crt --key .intra/certs/robot.key https://127.0.0.1:18443/svc/logs/)
[ "$code" = 403 ] || { echo "FAIL robot logs $code"; exit 1; }
echo "PASS robot sensitive HTTP denied"

REQ=$(./bin/intra request --request-key e2e-db --actor owner --device mac --services sensitive-db,logs,metrics --actions read --reason INC-E2E --incident INC-E2E --ttl 600)
GID=$(printf '%s' "$REQ" | python3 -c 'import json,sys; print(json.load(sys.stdin)["grant"]["id"])')
TOKEN=$(printf '%s' "$REQ" | python3 -c 'import json,sys; print(json.load(sys.stdin)["token"])')
./bin/intra approve --data .intra --id "$GID" --approver owner --dev-approve >/dev/null
echo "PASS incident grant approved (dev approver; Touch ID manual lane not claimed)"

for svc in logs metrics; do
  code=$(curl -s -o /dev/null -w '%{http_code}' -H "Authorization: Bearer $TOKEN" --cacert .intra/certs/ca.crt --cert .intra/certs/owner.crt --key .intra/certs/owner.key "https://127.0.0.1:18443/svc/$svc/")
  [ "$code" = 200 ] || { echo "FAIL incident $svc $code"; exit 1; }
done
echo "PASS incident logs and metrics read"

OUT=$(docker run --rm --network intra-lab_edge -v "$PWD/.intra/certs:/certs:ro" postgres:16-alpine psql "host=intra port=15432 user=$GID password=$TOKEN dbname=labdb sslmode=verify-ca sslrootcert=/certs/ca.crt sslcert=/certs/owner.crt sslkey=/certs/owner.key" -Atc "SELECT value FROM secrets WHERE id=1")
[ "$OUT" = "CANARY-SECRET-123" ] || { echo "FAIL pg read: $OUT"; exit 1; }
echo "PASS Postgres read through broker gateway"

if docker run --rm --network intra-lab_edge -v "$PWD/.intra/certs:/certs:ro" postgres:16-alpine psql "host=intra port=15432 user=$GID password=$TOKEN dbname=labdb sslmode=verify-ca sslrootcert=/certs/ca.crt sslcert=/certs/owner.crt sslkey=/certs/owner.key" -Atc "INSERT INTO secrets VALUES (2,'x')" >/dev/null 2>&1; then
  echo "FAIL write succeeded"; exit 1
fi
echo "PASS Postgres write denied"

./bin/intra revoke --id "$GID" --by owner >/dev/null
if docker run --rm --network intra-lab_edge -v "$PWD/.intra/certs:/certs:ro" postgres:16-alpine psql "host=intra port=15432 user=$GID password=$TOKEN dbname=labdb sslmode=verify-ca sslrootcert=/certs/ca.crt sslcert=/certs/owner.crt sslkey=/certs/owner.key" -Atc "SELECT 1" >/dev/null 2>&1; then
  echo "FAIL revoked grant reconnected"; exit 1
fi
echo "PASS revoked grant denied"

./bin/intra anchor --data .intra
./bin/intra verify-anchor --data .intra
echo "PASS audit anchor"

./bin/intra status --data .intra >/dev/null
echo "PASS CLI status"

HITS=$(curl -fsS http://127.0.0.1:18093/hits)
case "$HITS" in *CANARY-SECRET-123*) echo "FAIL canary at attacker"; exit 1;; esac
echo "PASS canary absent from attacker sink"

./bin/intra latency --direct http://127.0.0.1:18093/hits --guarded https://127.0.0.1:18443/svc/normal/ --cacert .intra/certs/ca.crt --cert .intra/certs/owner.crt --key .intra/certs/owner.key --n 50
./bin/intra simulate --policy .intra/policy.json >/dev/null
./bin/intra draft --data .intra >/dev/null
./bin/intra providers >/dev/null
echo "PASS simulate, draft, provider simulators"

echo "LAB E2E PASS"
