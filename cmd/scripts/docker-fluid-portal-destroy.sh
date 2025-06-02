#!/usr/bin/env bash

CONTAINER=$(docker ps -a -q --filter="name=fluid-fluid")
docker stop $CONTAINER || true
docker rm $CONTAINER || true
docker rmi $(docker images -a -q "fluid-fluid") || true
