#!/bin/bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
if [ -f "$ROOT_DIR/.env" ]; then
    set -a
    . "$ROOT_DIR/.env"
    set +a
fi

MONGO_PORT=27018
REDIS_PORT=6380
MAX_RETRIES=5
MONGO_DB="${MONGO_DB:-rent-house}"
MONGO_AUTH_SOURCE="${MONGO_AUTH_SOURCE:-rent-house}"
MONGO_USERNAME="${MONGO_USERNAME:-${MONGODB_USERNAME:-}}"
MONGO_PASSWORD="${MONGO_PASSWORD:-${MONGODB_PASSWORD:-}}"

check_port() {
    nc -z -w 2 127.0.0.1 $1 > /dev/null 2>&1
}

mongo_uri() {
    if [ -n "${MONGO_USERNAME}" ]; then
        printf 'mongodb://%s:%s@127.0.0.1:%s/%s?authSource=%s' "${MONGO_USERNAME}" "${MONGO_PASSWORD}" "${MONGO_PORT}" "${MONGO_DB}" "${MONGO_AUTH_SOURCE}"
    else
        printf 'mongodb://127.0.0.1:%s/%s' "${MONGO_PORT}" "${MONGO_DB}"
    fi
}

check_mongo_ready() {
    mongosh "$(mongo_uri)" --quiet --eval 'db.runCommand({ ping: 1 }).ok' > /dev/null 2>&1
}

check_redis_ready() {
    if [ -n "${REDIS_PASSWORD:-}" ]; then
        redis-cli -h 127.0.0.1 -p "${REDIS_PORT}" -a "${REDIS_PASSWORD}" PING 2>/dev/null | grep -q '^PONG$'
    else
        redis-cli -h 127.0.0.1 -p "${REDIS_PORT}" PING 2>/dev/null | grep -q '^PONG$'
    fi
}

restart_listener() {
    local port="$1"
    local pids
    pids="$(lsof -tiTCP:${port} -sTCP:LISTEN 2>/dev/null || true)"
    if [ -n "${pids}" ]; then
        kill ${pids} 2>/dev/null || true
        sleep 1
    fi
}

if check_port $MONGO_PORT; then
    if check_mongo_ready; then
        echo "✅ Mongo 隧道已保持连接 ($MONGO_PORT)"
    else
        echo "⚠️  Mongo 端口已占用但不可用，准备重连 ($MONGO_PORT)..."
        restart_listener $MONGO_PORT
    fi
fi

if ! check_mongo_ready; then
    echo "🔌 正在打通 Mongo 隧道 ($MONGO_PORT)..."
    kubectl114 port-forward -n ai-house svc/mongo-svc $MONGO_PORT:27017 > /dev/null 2>&1 &
    
    COUNT=0
    while ! check_mongo_ready; do
        ((COUNT++))
        if [ $COUNT -ge $MAX_RETRIES ]; then
            echo "❌ Mongo 隧道启动超时，请检查 K8s 集群状态或 kubectl114 命令。"
            exit 1
        fi
        echo "⏳ 等待 Mongo 就绪 ($COUNT/$MAX_RETRIES)..."
        sleep 1
    done
    echo "✅ Mongo 隧道已成功建立"
fi

if check_port $REDIS_PORT; then
    if check_redis_ready; then
        echo "✅ Redis 隧道已保持连接 ($REDIS_PORT)"
    else
        echo "⚠️  Redis 端口已占用但不可用，准备重连 ($REDIS_PORT)..."
        restart_listener $REDIS_PORT
    fi
fi

if ! check_redis_ready; then
    echo "🔌 正在打通 Redis 隧道 ($REDIS_PORT)..."
    kubectl114 port-forward -n ai-house svc/redis-svc $REDIS_PORT:6379 > /dev/null 2>&1 &
    
    COUNT=0
    while ! check_redis_ready; do
        ((COUNT++))
        if [ $COUNT -ge $MAX_RETRIES ]; then
            echo "❌ Redis 隧道启动超时，请检查 svc/redis-svc 是否正常。"
            exit 1
        fi
        echo "⏳ 等待 Redis 就绪 ($COUNT/$MAX_RETRIES)..."
        sleep 1
    done
    echo "✅ Redis 隧道已成功建立"
fi

echo "----------------------------------------"
echo "🚀 所有基础设施已就绪！"
