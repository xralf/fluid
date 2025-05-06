#!/bin/bash

./generator ./catalog.json instance1.database1.schema1.table1 50000 > ./sample.csv 2>> ./fluid.log

#$(GENERATOR) $(CATALOGJ) $(TABLE_NAME) $(NUM_ROWS) > $(TABLE_NAME_CSV) 2>> $(LOG)
#cp $(TABLE_NAME_CSV) $(JOB_DIR)
#cp $(TABLE_NAME_CSV) $(JOB_DIR)/sample.csv
