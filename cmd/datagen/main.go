package main

import (
	"errors"
	"os"
	"strconv"

	"github.com/xralf/fluid/pkg/datagen"
)

func main() {
	err := errors.New("unknown or missing argument\nusage: generator <catalog-path> <table-name> <number-of-rows>")

	if len(os.Args) != 4 {
		panic(err)
	}

	catalogName := os.Args[1]
	tableName := os.Args[2]
	numRows, err := strconv.Atoi(os.Args[3])
	if err != nil {
		panic(err)
	}

	datagen.Generate(catalogName, tableName, numRows)
}
