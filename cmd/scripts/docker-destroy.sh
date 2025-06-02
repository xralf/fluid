#!/usr/bin/env bash

docker stop $(docker ps -q) || true
docker rm $(docker ps -a -q) || true
docker rmi $(docker images -a -q) || true
docker system prune -a || true
