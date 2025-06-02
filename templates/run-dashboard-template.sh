#!/usr/bin/env bash

PIPE_1_INGRESS_PORT=@@@PIPE_1_INGRESS_PORT@@@
PIPE_1_EGRESS_PORT=@@@PIPE_1_EGRESS_PORT@@@
PIPE_2_INGRESS_PORT=@@@PIPE_2_INGRESS_PORT@@@
PIPE_2_EGRESS_PORT=@@@PIPE_2_EGRESS_PORT@@@
DASHBOARD_PORT=@@@DASHBOARD_PORT@@@
JOB_LOG=@@@JOB_LOG@@@
EXIT_AFTER_SECONDS=@@@EXIT_AFTER_SECONDS@@@

cp dashboard-template.html dashboard.html

if [[ $OSTYPE == 'darwin'* ]]; then
  sed -i '' "s/@@@PORT1@@@/$PIPE_1_EGRESS_PORT/g" dashboard.html
  sed -i '' "s/@@@PORT2@@@/$PIPE_2_EGRESS_PORT/g" dashboard.html
else
  sed -i "s/@@@PORT1@@@/$PIPE_1_EGRESS_PORT/g" dashboard.html
  sed -i "s/@@@PORT2@@@/$PIPE_2_EGRESS_PORT/g" dashboard.html
fi

./demo-pipe-server $PIPE_1_INGRESS_PORT $PIPE_1_EGRESS_PORT &
./demo-pipe-server $PIPE_2_INGRESS_PORT $PIPE_2_EGRESS_PORT &
./demo-web-server $DASHBOARD_PORT &

@@@DATA_SPOUT@@@ \
| tee \
>(./demo-pipe-client localhost:$PIPE_1_INGRESS_PORT in 1>/dev/null) \
>(./fluid -p ./plan.bin -x $EXIT_AFTER_SECONDS 2>> $JOB_LOG | ./demo-pipe-client localhost:$PIPE_2_INGRESS_PORT in) 1>/dev/null &

# Now:
#open http://localhost:$DASHBOARD_PORT
