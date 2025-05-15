#!/bin/#!/usr/bin/env bash

EXAMPLE=synthetic-slice-time-live

TEMPLATES_DIR=./examples
TEMPLATE=${TEMPLATES_DIR}/${EXAMPLE}

JOBS_DIR=/tmp/jobs
JOB=${JOBS_DIR}/${EXAMPLE}
JOB_DATA=${JOB}/sample.csv
THROTTLE=${JOB}/throttle
JOB_ENGINE=${JOB}/fluid
JOB_PLANB=${JOB}/plan.bin
EXIT_AFTER_SECONDS=3600
JOB_LOG=${JOB}/fluid.log

rm -rf ${JOBS_DIR}
mkdir -p ${JOBS_DIR}
cp -r ${TEMPLATE} ${JOBS_DIR}

JOB_PATH=${JOB} make build

cp -f cmd/throttle/throttle ${JOB}
cp -f cmd/datagen/generator ${JOB}

cd ${JOB}; ./prep.sh
cat ${JOB_DATA} | ${THROTTLE} --milliseconds 100 --append-timestamp false | ${JOB_ENGINE} -p ${JOB_PLANB} -x ${EXIT_AFTER_SECONDS} 2>> ${JOB_LOG}
