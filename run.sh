#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"
./dev.sh
if [ -f .env ]; then
  set -a
  . ./.env
  set +a
fi

PORT=8080
LISTEN_PIDS="$(lsof -tiTCP:${PORT} -sTCP:LISTEN 2>/dev/null || true)"
if [ -n "${LISTEN_PIDS}" ]; then
  for pid in ${LISTEN_PIDS}; do
    cmd="$(ps -p "${pid}" -o command= 2>/dev/null || true)"
    if [[ "${cmd}" == *"house-manager"* ]] || [[ "${cmd}" == *"./cmd/server"* ]] || [[ "${cmd}" == *"config/config.local.yaml"* ]]; then
      kill "${pid}" 2>/dev/null || true
      wait_count=0
      while kill -0 "${pid}" 2>/dev/null; do
        wait_count=$((wait_count + 1))
        if [ "${wait_count}" -ge 20 ]; then
          echo "failed to stop existing house-manager process on port ${PORT}"
          exit 1
        fi
        sleep 0.2
      done
    else
      echo "port ${PORT} is already in use by another process:"
      echo "${cmd}"
      exit 1
    fi
  done
fi

go run ./cmd/server -c ./config/config.local.yaml
