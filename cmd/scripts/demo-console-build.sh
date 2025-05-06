#!/bin/bash

EXAMPLE=$1
JOB_DIR=/tmp/jobs/${EXAMPLE}

mkdir -p ${JOB_DIR}
cp -r /tmp/demo/* ${JOB_DIR}
cp -r examples/${EXAMPLE} /tmp/jobs
cd repos/fluid; JOB_DIR=${JOB_DIR} make build
cd ${JOB_DIR}; ./prep.sh
