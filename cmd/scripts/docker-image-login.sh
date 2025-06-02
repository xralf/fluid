#!/usr/bin/env bash

##
## Example: docker-image-login.sh fluid-base:0.1
##

docker exec -it $(docker container ls --all | grep -w fluid-base:0.1 | awk '{print $1}') bash
