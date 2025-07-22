#!/bin/bash
set -e

if [ -z "$1" ]; then
  echo "Error: BOT_TOKEN not provided"
  exit 1
fi

BACKEND_URL="http://telegram-splitter-app:8080"

docker network inspect my-network >/dev/null 2>&1 || \
  docker network create my-network

docker pull maxloginoff/tg_splitter_adapter-app:latest
docker stop tg_splitter_adapter-app || true
docker rm tg_splitter_adapter-app || true

docker run -d \
  --name tg_splitter_adapter-app \
  --network my-network \
  -p 8081:8081 \
  -e BOT_TOKEN="$1" \
  -e BACKEND_URL="$BACKEND_URL" \
  maxloginoff/tg_splitter_adapter-app:latest
