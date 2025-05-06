package main

import (
	"errors"
	"os"
	"strconv"
	"strings"

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

	fieldNames, fieldTypes, fieldUsages, csvTypes, maxValues := datagen.ExtractFieldInfo(catalogName, tableName)
	preamble := "# " + strings.Join(fieldNames, ", ") +
		"\n# " + strings.Join(fieldTypes, ", ") +
		"\n# " + strings.Join(csvTypes, ", ") +
		"\n# " + strings.Join(datagen.AsText(maxValues), ", ") + "\n"

	data := [][]string{}
	datagen.CreateData(fieldNames, fieldUsages, csvTypes, maxValues, numRows, &data)
	datagen.WriteData(preamble, data)
}
