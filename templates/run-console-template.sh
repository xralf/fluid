#!/usr/bin/env bash

CSV_FILE=@@@CSV_FILE@@@
THROTTLE_MILLISECONDS=@@@THROTTLE_MILLISECONDS@@@
EXIT_AFTER_SECONDS=@@@EXIT_AFTER_SECONDS@@@

cat $CSV_FILE | ./throttle --milliseconds $THROTTLE_MILLISECONDS --append-timestamp false | ./fluid -p ./plan.bin -x $EXIT_AFTER_SECONDS 2>> ./fluid.log
