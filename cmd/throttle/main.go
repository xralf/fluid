package main

import (
	"errors"
	"os"
	"strconv"

	"github.com/xralf/fluid/pkg/throttle"
)

func main() {
	if len(os.Args) != 5 {
		panic(errors.New("unknown or missing argument\nusage: throttle --milliseconds <integer> --append-timestamp <true|false>"))
	}

	var err error
	var sleepMilliseconds int
	if sleepMilliseconds, err = strconv.Atoi(os.Args[2]); err != nil {
		panic(err)
	}

	var appendTimestamp bool
	if appendTimestamp, err = strconv.ParseBool(os.Args[4]); err != nil {
		panic(err)
	}

	delimiter := "|"
	throttle.Throttle(sleepMilliseconds, appendTimestamp, delimiter)
}
