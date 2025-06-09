#!/bin/bash

if [ -z "$1" ]; then
  echo "Error: BOT_TOKEN not provided"
  exit 1
fi

docker pull brawleryura1/tg_splitter_adapter-app:latest
docker stop tg_splitter_adapter-app || true
docker rm tg_splitter_adapter-app || true
docker run -d --name tg_splitter_adapter-app -p 8081:8081 -e BOT_TOKEN="$1" brawleryura1/tg_splitter_adapter-app:latest