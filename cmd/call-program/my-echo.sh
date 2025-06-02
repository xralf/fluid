#!/usr/bin/env bash

for (( ; ; ))
do
  echo "Enter some text:"
  read -r input
  sleep 1
  echo "You typed: $input"
done
