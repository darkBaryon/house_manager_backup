#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

echo "[1/3] generating wire..."
(cd wire && go run -mod=mod github.com/google/wire/cmd/wire)

echo "[2/3] building..."
go build -o bin/house-manager ./cmd/server

echo "[3/3] done: bin/house-manager"
