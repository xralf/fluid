##
## This Makefile expects 2 input arguments, which are provided as shell variables:
##
##   - EXAMPLE_PATH:  Absolute path to the input example directory
##   - JOB_DIR:       Absolute path to the output job directory
##
## Example command for this Makefile:
##
##    EXAMPLE_PATH=~/git/xralf/fluid/examples/synthetic-slice-time-live \
##    JOB_DIR=/tmp/myjobs \
##    make build
##
##    EXAMPLE_PATH=~/git/xralf/fluid/examples/synthetic-slice-time-live JOB_DIR=/tmp/myjobs make all
##
##    EXAMPLE_PATH= JOB_DIR= make all
##

##
## Some default example source for missing argument
##
ifeq ($(EXAMPLE_PATH),)
EXAMPLE_PATH          := examples/synthetic-slice-time-live
endif
EXAMPLE_BASE	      := $(shell basename $(EXAMPLE_PATH))
EXAMPLE_DIR	          := $(shell dirname $(EXAMPLE_PATH))

##
## Some default destination example for missing argument
##
ifeq ($(JOB_DIR),)
JOB_DIR               := /tmp/myjobs2
endif
JOB_PATH	          := $(JOB_DIR)/$(EXAMPLE_BASE)

##
## Call this Makefile as shown in run-example.sh
##
## JOB_DIR is a shell variable passed to the Makefile.
##

THROTTLE_MILLISECONDS := 20
EXIT_AFTER_SECONDS    := 3600

REPO                  := github.com/xralf/fluid

JOB_DATA              := $(JOB_PATH)/sample.csv
JOB_ENGINE            := $(JOB_DIR)/fluid
JOB_PLANB             := $(JOB_DIR)/plan.bin
JOB_PLANJ             := $(JOB_DIR)/plan.json
JOB_LOG               := $(JOB_DIR)/fluid.log

CAPNP_PATH            := $(HOME)/git/capnproto
CATALOG_PATH          := cmd/catalog
COMPILER_PATH         := cmd/fluidc
ENGINE_PATH           := cmd/fluid
THROTTLE_PATH         := cmd/throttle
PKG_OUT_PATH          := pkg/_out
QUERY_PATH            := $(PKG_OUT_PATH)/query
FUNCTIONS_PATH        := $(PKG_OUT_PATH)/functions
OUT_PATH              := _out
PLAN_PATH             := $(OUT_PATH)
CATALOG_OUT_PATH      := $(OUT_PATH)/catalog
CSV_DATA_PATH         := $(OUT_PATH)/csv_data
CSV_TEMPLATE_PATH     := $(OUT_PATH)/csv_templates
EXAMPLE_QUERY_PATH    := $(JOB_PATH)/query.fql
TEMPLATE_PATH         := templates

LOG                   := fluid.log
CATALOGJ_MASTER       := $(JOB_PATH)/catalog.json
CATALOGB              := $(OUT_PATH)/catalog.bin
CATALOGJ              := $(OUT_PATH)/catalog.json
PLANB                 := $(PLAN_PATH)/plan.bin
PLANJ                 := $(PLAN_PATH)/plan.json
#TABLE_NAME            := $(shell head -1 $(EXAMPLE_QUERY_PATH) | awk '{print $$2}')

ANTLRGEN              := BaseListener Lexer Listener Parser
GRAMMAR_QUERY         := FQL
OS_NAME               := $(shell uname -s)

CATALOG               := $(CATALOG_PATH)/catalog
COMPILER              := $(COMPILER_PATH)/fluidc
ENGINE                := $(ENGINE_PATH)/fluid
THROTTLE			  := $(THROTTLE_PATH)/throttle

ifeq ($(OS_NAME), Darwin)
ANTLR4                := antlr
else ifeq ($(OS_NAME), Linux)
ANTLR4                := java -Xmx500m -cp "jars/antlr-4.13.2-complete.jar:CLASSPATH" org.antlr.v4.Tool
else
ANTLR4                := "Unknown operating system name: $(OS_NAME)"
endif

all: build
	@echo "JOB_DIR" $(JOB_DIR)
	@echo "JOB_PATH" $(JOB_PATH)
	@echo "EXAMPLE_PATH" $(EXAMPLE_PATH)
	@echo "EXAMPLE_BASE" $(EXAMPLE_BASE)
	@echo "EXAMPLE_DIR" $(EXAMPLE_DIR)

example: build generate run

generate:
	cp -f cmd/datagen/generator $(JOB_PATH)
	cd $(JOB_PATH); ./prep.sh

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
	mkdir -p $(JOB_DIR)
	cp -r $(EXAMPLE_PATH) $(JOB_DIR)
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
	cp $(PLANB) $(JOB_DIR)
	gofmt -w $(FUNCTIONS_PATH)/functions.go
	@cat $(PLANB) | $(COMPILER) show > $(PLANJ)
	cp $(PLANJ) $(JOB_DIR)
#	@cat $(PLANJ) | jq '.' --indent 4
	cd capnp/data; go generate
	go build $(FUNCTIONS_PATH)/functions.go
	go build -o $(ENGINE) $(ENGINE_PATH)/main.go
	cp $(ENGINE) $(JOB_DIR)
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
#	rm -rf $(JOB_DIR)
	rm -rf $(PKG_OUT_PATH)
	rm -rf $(CAPNP_PATH)/go-capnproto2
	rm -rf $(OUT_PATH)
	rm -f ./capnp/books/*.capnp.go
	rm -f ./capnp/data/data.capnp
	rm -f ./capnp/data/*.capnp.go
	rm -f ./capnp/foo/*.capnp.go
	rm -f ./capnp/fluid/*.capnp.go
	rm -f ./capnp/person/*.capnp.go
