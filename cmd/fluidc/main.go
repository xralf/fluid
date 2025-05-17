// This program translates a FQL query string and generates
//
//   1. A text Cap'n Proto file (plan.capnp) with the schema (mostly fields and types) of
//      each node in the query plan.
//
//   2. A binary Cap'n Proto query plan file according to the fluid schema (fluid.capnp)
//
// There are 2 different parameters:
//
//   1. compile:  Given a FQL query, generate the binary query plan
//
//   2. show:     Given a binary query plan, generate a JSON representation of the query plan
//
// Compilation:
//
// stdin (FQL query)  --->  ./compiler compile  --->  stdout (binary Cap'n Proto stream)
//                                               |
//                                               +->  file with Cap'n Proto schemas (schemas.capnp)
//                                                    (this file is a side effect)
// Example:
//
//   echo "from table1 where x >= 5 project a, b" | ./compiler compile > ./plan.bin
//   cat ./plan.bin | ./compiler show | jq . | tee ./plan_pretty.json
//

package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/xralf/fluid/pkg/plan"
	"github.com/xralf/fluid/pkg/utility"

	"github.com/xralf/fluid/pkg/compiler"
)

var (
	logger *slog.Logger
)

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}))
	logger.Info("Compiler says welcome!")
}

func main() {
	err := errors.New("unknown or missing argument\nusage: fluidc [compile|show]")

	if len(os.Args) != 2 {
		panic(err)
	}

	cmdArgs := os.Args[1]

	logger.Info(fmt.Sprintf("command used: %s", cmdArgs))

	compiler.Init()
	switch cmdArgs {
	case "compile":
		compiler.Compile()
	case "show":
		utility.ShowPlan()
	default:
		err := errors.New("unknown or missing argument/nusage: z [compile|show]")
		fmt.Print(err)
		logger.Error(err.Error())
		os.Exit(0)
	}

	logger.Info("Compiler says good-bye!")
}
