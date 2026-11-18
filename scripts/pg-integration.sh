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
for i in $(seq 1 60); do
  if docker exec "$NAME" pg_isready -U postgres -d labdb >/dev/null 2>&1; then break; fi
  sleep 1
done
docker exec "$NAME" pg_isready -U postgres -d labdb
INTRA_TEST_PG_ADMIN_URL='postgres://intra_broker_admin@127.0.0.1:35432/labdb?sslmode=disable' \
INTRA_TEST_PG_BASE_URL='postgres://unused@127.0.0.1:35432/labdb?sslmode=disable' \
  go test -race -tags integration ./internal/pggateway -run TestGatewayGrantLifecycle -count=1 -v
