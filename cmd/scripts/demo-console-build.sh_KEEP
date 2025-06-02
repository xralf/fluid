#!/bin/#!/usr/bin/env bash

EXAMPLE=$1
JOB_PATH=/tmp/jobs/${EXAMPLE}

mkdir -p ${JOB_PATH}
cp -r examples/${EXAMPLE} /tmp/jobs
cd repos/fluid; JOB_PATH=${JOB_PATH} make build
cp -r /tmp/demo/* ${JOB_PATH}
cd ${JOB_PATH}; ./prep.sh
