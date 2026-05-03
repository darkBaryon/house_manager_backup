#!/bin/bash

MONGO_PORT=27018
REDIS_PORT=6380
MAX_RETRIES=5

check_port() {
    nc -z -w 2 127.0.0.1 $1 > /dev/null 2>&1
}

if ! check_port $MONGO_PORT; then
    echo "🔌 正在打通 Mongo 隧道 ($MONGO_PORT)..."
    kubectl114 port-forward -n ai-house svc/mongo-svc $MONGO_PORT:27017 > /dev/null 2>&1 &
    
    COUNT=0
    while ! check_port $MONGO_PORT; do
        ((COUNT++))
        if [ $COUNT -ge $MAX_RETRIES ]; then
            echo "❌ Mongo 隧道启动超时，请检查 K8s 集群状态或 kubectl114 命令。"
            exit 1
        fi
        echo "⏳ 等待 Mongo 就绪 ($COUNT/$MAX_RETRIES)..."
        sleep 1
    done
    echo "✅ Mongo 隧道已成功建立"
else
    echo "✅ Mongo 隧道已保持连接 ($MONGO_PORT)"
fi

if ! check_port $REDIS_PORT; then
    echo "🔌 正在打通 Redis 隧道 ($REDIS_PORT)..."
    kubectl port-forward -n ai-house svc/redis-svc $REDIS_PORT:6379 > /dev/null 2>&1 &
    
    COUNT=0
    while ! check_port $REDIS_PORT; do
        ((COUNT++))
        if [ $COUNT -ge $MAX_RETRIES ]; then
            echo "❌ Redis 隧道启动超时，请检查 svc/redis-svc 是否正常。"
            exit 1
        fi
        echo "⏳ 等待 Redis 就绪 ($COUNT/$MAX_RETRIES)..."
        sleep 1
    done
    echo "✅ Redis 隧道已成功建立"
else
    echo "✅ Redis 隧道已保持连接 ($REDIS_PORT)"
fi

echo "----------------------------------------"
echo "🚀 所有基础设施已就绪！"