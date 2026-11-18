#!/bin/sh
set -eu
NAME=intra-test-pg
cleanup() { docker rm -f "$NAME" >/dev/null 2>&1 || true; }
trap cleanup EXIT
cleanup
docker run -d --name "$NAME" \
  -e POSTGRES_DB=labdb -e POSTGRES_USER=postgres -e POSTGRES_HOST_AUTH_METHOD=trust \
  -p 127.0.0.1:35432:5432 \
  -v "$PWD/lab/postgres/init.sql:/docker-entrypoint-initdb.d/001-intra.sql:ro" \
  postgres:16-alpine >/dev/null
# Wait for the init marker, not only the socket. The entrypoint temp server
# accepts connections before init.sql finishes, then restarts. A pg_isready
# pass during that window breaks the loop before the real server is ready.
ready() {
  docker exec "$NAME" psql -U postgres -d labdb -tAc \
    "SELECT 1 FROM pg_proc WHERE proname='intra_create_grant_role'" 2>/dev/null | grep -q 1
}
for i in $(seq 1 60); do
  if ready; then break; fi
  sleep 1
done
ready
INTRA_TEST_PG_ADMIN_URL='postgres://intra_broker_admin@127.0.0.1:35432/labdb?sslmode=disable' \
INTRA_TEST_PG_BASE_URL='postgres://unused@127.0.0.1:35432/labdb?sslmode=disable' \
  go test -race -tags integration ./internal/pggateway -run TestGatewayGrantLifecycle -count=1 -v
