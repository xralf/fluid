##
## This Makefile expects 2 input arguments, which are provided as shell variables:
##
##   - SRC_EXAMPLE_PATH:  Absolute path to the input example directory
##   - DST_JOB_PATH:      Absolute path to the output job directory
##
## Example command for this Makefile:
##
##    SRC_EXAMPLE_PATH=~/git/xralf/fluid/examples/synthetic-slice-time-live \
##    DST_JOB_PATH=/tmp/myjobs \
##    make build
##

# SRC_EXAMPLE_PATH=~/git/xralf/fluid/examples/synthetic-slice-time-live DST_JOB_PATH=/tmp/myjobs make all
# SRC_EXAMPLE_PATH= DST_JOB_PATH= make all

## Some default destination example for missing argument
ifeq ($(DST_JOB_PATH),)
DST_JOB_PATH           := /tmp/myjobs2
endif

## Some default example source for missing argument
ifeq ($(SRC_EXAMPLE_PATH),)
SRC_EXAMPLE_PATH       := examples/synthetic-slice-time-live
endif

SRC_EXAMPLE_BASE	   := $(shell basename $(SRC_EXAMPLE_PATH))
SRC_EXAMPLE_DIR	       := $(shell dirname $(SRC_EXAMPLE_PATH))

DST_EXAMPLE_BASE	   := $(SRC_EXAMPLE_BASE)
DST_EXAMPLE_PATH	   := $(DST_JOB_PATH)/$(SRC_EXAMPLE_BASE)

##
## Call this Makefile as shown in run-example.sh
##
## JOB_PATH is a shell variable passed to the Makefile.
##

THROTTLE_MILLISECONDS  := 20
EXIT_AFTER_SECONDS     := 3600

REPO                   := github.com/xralf/fluid

# EXAMPLE                := synthetic-slice-time-live
# EXAMPLES_PATH          := examples
# EXAMPLE_TEMPLATE       := ${EXAMPLES_PATH}/$(EXAMPLE)
EXAMPLE                := $(DST_EXAMPLE_BASE)
EXAMPLES_PATH          := $(DST_EXAMPLE_PATH)
EXAMPLE_TEMPLATE       := $(DST_EXAMPLE_PATH)

#JOBS_PATH              := /tmp/jobs
JOBS_PATH              := $(DST_JOB_PATH)
#JOB_PATH               := $(JOBS_PATH)/$(EXAMPLE)
JOB_PATH               := $(DST_JOB_PATH)

#JOB_DATA               := $(JOB_PATH)/$(TABLE_NAME).csv
JOB_DATA               := $(DST_EXAMPLE_PATH)/sample.csv
JOB_ENGINE             := $(JOB_PATH)/fluid
JOB_PLANB              := $(JOB_PATH)/plan.bin
JOB_PLANJ              := $(JOB_PATH)/plan.json
JOB_LOG                := $(JOB_PATH)/fluid.log

CAPNP_PATH             := $(HOME)/git/capnproto
CATALOG_PATH           := cmd/catalog
COMPILER_PATH          := cmd/fluidc
ENGINE_PATH            := cmd/fluid
THROTTLE_PATH          := cmd/throttle
PKG_OUT_PATH           := pkg/_out
QUERY_PATH             := $(PKG_OUT_PATH)/query
FUNCTIONS_PATH         := $(PKG_OUT_PATH)/functions
OUT_PATH               := _out
PLAN_PATH              := $(OUT_PATH)
CATALOG_OUT_PATH       := $(OUT_PATH)/catalog
CSV_DATA_PATH          := $(OUT_PATH)/csv_data
CSV_TEMPLATE_PATH      := $(OUT_PATH)/csv_templates
EXAMPLE_QUERY_PATH     := $(DST_EXAMPLE_PATH)/query.fql
TEMPLATE_PATH          := templates

LOG                    := fluid.log
#CATALOGJ_MASTER        := $(JOB_PATH)/catalog.json
CATALOGJ_MASTER        := $(DST_EXAMPLE_PATH)/catalog.json
CATALOGB               := $(OUT_PATH)/catalog.bin
CATALOGJ               := $(OUT_PATH)/catalog.json
PLANB                  := $(PLAN_PATH)/plan.bin
PLANJ                  := $(PLAN_PATH)/plan.json
#TABLE_NAME             := $(shell head -1 $(EXAMPLE_QUERY_PATH) | awk '{print $$2}')

ANTLRGEN               := BaseListener Lexer Listener Parser
GRAMMAR_QUERY          := FQL
OS_NAME                := $(shell uname -s)

CATALOG                := $(CATALOG_PATH)/catalog
COMPILER               := $(COMPILER_PATH)/fluidc
ENGINE                 := $(ENGINE_PATH)/fluid
THROTTLE			   := $(THROTTLE_PATH)/throttle

ifeq ($(OS_NAME), Darwin)
ANTLR4                 := antlr
else ifeq ($(OS_NAME), Linux)
ANTLR4                 := java -Xmx500m -cp "jars/antlr-4.13.2-complete.jar:CLASSPATH" org.antlr.v4.Tool
else
ANTLR4                 := "Unknown operating system name: $(OS_NAME)"
endif

all: build
	@echo "DST_JOB_PATH" $(DST_JOB_PATH)
	@echo "SRC_EXAMPLE_PATH" $(SRC_EXAMPLE_PATH)
	@echo "SRC_EXAMPLE_BASE" $(SRC_EXAMPLE_BASE)
	@echo "SRC_EXAMPLE_DIR" $(SRC_EXAMPLE_DIR)
	@echo "DST_EXAMPLE_BASE" $(DST_EXAMPLE_BASE)
	@echo "DST_EXAMPLE_PATH" $(DST_EXAMPLE_PATH)

example: build generate run

generate:
	cp -f cmd/datagen/generator $(DST_EXAMPLE_PATH)
	cd $(DST_EXAMPLE_PATH); ./prep.sh

run:
	@cat $(JOB_DATA) | $(THROTTLE) --milliseconds 100 --append-timestamp false | $(ENGINE) -p $(PLANB) -x $(EXIT_AFTER_SECONDS) 2>> $(LOG)

build: clean prepare build_compiler build_engine build_datagen build_throttle build_reverse

#again: clean_log mini_build run

# setup_example:
# 	mkdir -p /tmp/jobs
# 	cp -r examples/$(EXAMPLE_NAME) /tmp/jobs

doc_run:
	go install golang.org/x/tools/cmd/godoc@latest
	godoc -http=:6060

doc_view:
	open http://localhost:6060/pkg/github.com/xralf/fluid/pkg/compiler/

mini_build:
	$(ANTLR4) -Dlanguage=Go -o $(QUERY_PATH) $(GRAMMAR_QUERY).g4
	go build -o $(CATALOG) $(CATALOG_PATH)/main.go
	go build -o $(COMPILER) $(COMPILER_PATH)/main.go

prepare:
	mkdir -p $(JOBS_PATH)
	cp -r $(SRC_EXAMPLE_PATH) $(JOBS_PATH)
	mkdir -p $(JOB_PATH)
	mkdir -p $(CAPNP_PATH)
	mkdir -p $(OUT_PATH)
	mkdir -p $(FUNCTIONS_PATH)
	mkdir -p $(QUERY_PATH)
	mkdir -p $(PLAN_PATH)
	mkdir -p $(CSV_DATA_PATH)
	mkdir -p $(CSV_TEMPLATE_PATH)
	go mod init $(REPO)
#	go mod tidy
	go get github.com/antlr4-go/antlr/v4
	go get zombiezen.com/go/capnproto2
	go get capnproto.org/go/capnp/v3
	go install capnproto.org/go/capnp/v3/capnpc-go@latest
	go get github.com/DataDog/hyperloglog
	cd $(CAPNP_PATH); git clone https://github.com/capnproto/go-capnproto2.git
	cd capnp/fluid; go generate
	go mod edit -require=$(REPO)/capnp/data@v0.0.0-unpublished
	go mod edit -replace=$(REPO)/capnp/data@v0.0.0-unpublished=./capnp/data

build_compiler:
	$(ANTLR4) -Dlanguage=Go -o $(QUERY_PATH) $(GRAMMAR_QUERY).g4
	cd $(QUERY_PATH); go mod init $(REPO)/$(QUERY_PATH); go mod tidy
	cd $(FUNCTIONS_PATH); go mod init $(REPO)/$(FUNCTIONS_PATH); go mod tidy
	go mod edit -require=$(REPO)/pkg/_out/query/parser@v0.0.0-unpublished
	go mod edit -replace=$(REPO)/pkg/_out/query/parser@v0.0.0-unpublished=./pkg/_out/query
	go mod edit -require=$(REPO)/pkg/_out/functions@v0.0.0-unpublished
	go mod edit -replace=$(REPO)/pkg/_out/functions@v0.0.0-unpublished=./pkg/_out/functions
	go build -o $(CATALOG) $(CATALOG_PATH)/main.go
	go build -o $(COMPILER) $(COMPILER_PATH)/main.go

build_engine:
	@cat $(CATALOGJ_MASTER) | $(CATALOG) -i json -o capnp -t $(CSV_TEMPLATE_PATH) 2>> $(LOG) > $(CATALOGB)
#	@cat $(CATALOGB) | $(CATALOG) -i capnp -o jmson -t $(CSV_TEMPLATE_PATH) 2>> $(LOG) | tee $(CATALOGJ) | jq '.' --tab
	@cat $(CATALOGB) | $(CATALOG) -i capnp -o json -t $(CSV_TEMPLATE_PATH) 2>> $(LOG) > $(CATALOGJ)
	mkdir -p $(PLAN_PATH)
	@cat $(EXAMPLE_QUERY_PATH) | $(COMPILER) compile > $(PLANB) 2>> $(LOG)
	cp $(PLANB) $(JOB_PATH)
	gofmt -w $(FUNCTIONS_PATH)/functions.go
	@cat $(PLANB) | $(COMPILER) show > $(PLANJ)
	cp $(PLANJ) $(JOB_PATH)
#	@cat $(PLANJ) | jq '.' --indent 4
	cd capnp/data; go generate
	go build $(FUNCTIONS_PATH)/functions.go
	go build -o $(ENGINE) $(ENGINE_PATH)/main.go
	cp $(ENGINE) $(JOB_PATH)
	go mod tidy

build_datagen:
	go build -o cmd/datagen/generator cmd/datagen/main.go

build_throttle:
	go build -o cmd/throttle/throttle cmd/throttle/main.go

build_reverse:
	go build -o cmd/tools/reverse cmd/tools/reverse/reverse.go

run2:
#	@cat $(TABLE_NAME_CSV) | $(ENGINE) -p $(PLANB) 2>> $(LOG)
#	@cat $(JOB_DATA) | $(THROTTLE) --milliseconds 100 --append-timestamp false | $(ENGINE) -p $(PLANB) -x $(EXIT_AFTER_SECONDS) 2>> $(LOG)
	@cat $(JOB_DATA) | $(THROTTLE) --milliseconds 100 --append-timestamp false | $(JOB_ENGINE) -p $(JOB_PLANB) -x $(EXIT_AFTER_SECONDS) 2>> $(JOB_LOG)

rerun: # Don't remove the go.mod or install software again
	rm -f $(LOG)
	go build -o $(CATALOG) $(CATALOG_PATH)/main.go
	go build -o $(COMPILER) $(COMPILER_PATH)/main.go
	go build -o $(ENGINE) $(ENGINE_PATH)/main.go

run_syslog:
	$(SYSLOG) | $(ENGINE) -p $(PLANB) -x $(EXIT_AFTER_SECONDS) 2>> $(LOG)

# run_condition:
# 	$(CODEGEN_CONDITION) $(CONDITION)

run_fast:
	@cat $(JOB_DATA) | $(ENGINE) -p $(PLANB) -x $(EXIT_AFTER_SECONDS) 2>> $(LOG)

clean_log:
	rm -f $(LOG)

clean:
	rm -rf .antlr
	rm -f *.log
	rm -f $(LOG)
	rm -f $(CATALOG)
	rm -f $(COMPILER)
	rm -f $(ENGINE)
	rm -f cmd/datagen/generator
	rm -f cmd/throttle/throttle
	rm -f cmd/tools/reverse/reverse
	rm -f go.mod
	rm -f go.sum
	rm -f go.work.sum
	rm -rf $(DST_JOB_PATH)
	rm -rf $(JOBS_PATH)
	rm -rf $(PKG_OUT_PATH)
	rm -rf $(CAPNP_PATH)/go-capnproto2
	rm -rf $(OUT_PATH)
	rm -f ./capnp/books/*.capnp.go
	rm -f ./capnp/data/data.capnp
	rm -f ./capnp/data/*.capnp.go
	rm -f ./capnp/foo/*.capnp.go
	rm -f ./capnp/fluid/*.capnp.go
	rm -f ./capnp/person/*.capnp.go

showc: # Show catalog
	@cat $(CATALOGB) | $(CATALOG) -i capnp -o json | jq '.' --tab

showp: # Show plan
	@cat $(PLANB) | $(COMPILER) show | jq . --tab

tiny:
	./$(ENGINE) tiny > $(PLAN_PATH)/tiny.bin
	cat $(PLAN_PATH)/tiny.bin | ./$(ENGINE) show

example:
	mkdir -p $(PLAN_PATH)
	$(ENGINE) example > $(PLAN_PATH)/example.bin
	cat $(PLAN_PATH)/example.bin | $(ENGINE) show | jq .

justrun:
	echo $(QUERY) | ./$(COMPILER) compile | $(ENGINE_PATH)/$(ENGINE) example

test:
	go test -v
