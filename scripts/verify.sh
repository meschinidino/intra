#!/bin/sh
set -eu
cd "$(dirname "$0")/.."

echo "== unit =="
go vet ./...
go test ./...

echo "== touchid helper compile (manual biometric run not claimed) =="
swiftc -O -o bin/touchid-helper cmd/touchid-helper/main.swift

echo "== postgres gateway integration =="
./scripts/pg-integration.sh

echo "== docker lab e2e =="
./scripts/lab-e2e.sh

echo "== identity and onboarding =="
go build -o bin/intra ./cmd/intra
./bin/intra issue-identity --data .intra --actor robot --run verify1 >/dev/null
cat > /tmp/intra-tailscale.json <<'JSON'
{"acls":[{"src":["tag:dev"],"dst":["tag:db:5432"]}]}
JSON
rm -rf /tmp/intra-onboard-test
cp -R .intra /tmp/intra-onboard-test
START=$(date +%s)
./bin/intra onboard --data /tmp/intra-onboard-test --kind tailscale --file /tmp/intra-tailscale.json --service imported-check --upstream http://lab-normal:18090 >/dev/null
END=$(date +%s)
echo "onboard apply seconds: $((END-START))"

echo "VERIFY PASS"
