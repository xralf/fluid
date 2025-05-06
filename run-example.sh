#!/bin/bash

EXAMPLE=synthetic-slice-time-live

TEMPLATES_DIR=./examples
TEMPLATE=${TEMPLATES_DIR}/${EXAMPLE}

JOBS_DIR=/tmp/jobs
JOB=${JOBS_DIR}/${EXAMPLE}

# rm -rf ${JOBS_DIR}
# mkdir -p ${JOBS_DIR}
# cp -r ${TEMPLATE} ${JOBS_DIR}
# JOB_DIR=${JOB} make build

cd ${JOB}; ./prep.sh


# #ls -l ${TEMPLATE}
# #@cat $(JOB_DATA) | $(THROTTLE) --milliseconds 100 --append-timestamp false | $(JOB_ENGINE) -p $(JOB_PLANB) -x $(EXIT_AFTER_SECONDS) 2>> $(JOB_LOG)
