#!/bin/bash
set -e
apk add postgresql-client postgresql-libs
while ! pg_isready --user orchestrator --host postgresql &> /dev/null; do
  sleep 2
  echo "Waiting for the orchestrator database to become active."
done
migrate -verbose -path /home/databases/orchestrator -database postgres://orchestrator:orchestrator@postgresql:5432/orchestrator_development?sslmode=disable up
