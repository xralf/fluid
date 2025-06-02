#!/usr/bin/env bash

EXAMPLE=$1
JOB_PATH=/tmp/jobs/${EXAMPLE}

cd ${JOB_PATH}; cat sample.csv | ./throttle --milliseconds 300 --append-timestamp false | ./fluid -p ./plan.bin -x 3600 2>> ./fluid.log
