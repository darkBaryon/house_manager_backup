#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
if [[ -f "$ROOT_DIR/.env" ]]; then
  set -a
  source "$ROOT_DIR/.env"
  set +a
fi

MONGO_HOST="${MONGO_HOST:-127.0.0.1}"
MONGO_PORT="${MONGO_PORT:-27018}"
MONGO_DB="${MONGO_DB:-rent-house}"
MONGO_AUTH_SOURCE="${MONGO_AUTH_SOURCE:-rent-house}"
MONGO_USERNAME="${MONGO_USERNAME:-${MONGODB_USERNAME:-}}"
MONGO_PASSWORD="${MONGO_PASSWORD:-${MONGODB_PASSWORD:-}}"

if [[ -n "$MONGO_USERNAME" ]]; then
  MONGO_URI="mongodb://${MONGO_USERNAME}:${MONGO_PASSWORD}@${MONGO_HOST}:${MONGO_PORT}/${MONGO_DB}?authSource=${MONGO_AUTH_SOURCE}"
else
  MONGO_URI="mongodb://${MONGO_HOST}:${MONGO_PORT}/${MONGO_DB}"
fi

echo "[seed-hmd] mongo db: ${MONGO_DB}"
mongosh "$MONGO_URI" --quiet --file "$ROOT_DIR/scripts/seed_hmd_collections.mongosh.js"
