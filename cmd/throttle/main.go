package main

import (
	"errors"
	"os"
	"strconv"

	_ "github.com/xralf/fluid/pkg/throttle"
)

// func main() {
// 	if len(os.Args) != 5 {
// 		panic(errors.New("unknown or missing argument\nusage: throttle --milliseconds <integer> --append-timestamp <true|false>"))
// 	}

// 	var err error
// 	var sleepMilliseconds int
// 	if sleepMilliseconds, err = strconv.Atoi(os.Args[2]); err != nil {
// 		panic(err)
// 	}

// 	var appendTimestamp bool
// 	if appendTimestamp, err = strconv.ParseBool(os.Args[4]); err != nil {
// 		panic(err)
// 	}

// 	delimiter := "|"

// 	scanner := bufio.NewScanner(os.Stdin)
// 	for scanner.Scan() {
// 		line := scanner.Text()
// 		if err = scanner.Err(); err != nil {
// 			panic(err)
// 		}

// 		if appendTimestamp {
// 			ts := time.Now()
// 			line += delimiter + fmt.Sprint(ts.UTC().Format(time.RFC3339Nano))
// 		}
// 		fmt.Println(line)
// 		time.Sleep(time.Duration(sleepMilliseconds) * time.Millisecond)
// 	}
// }

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

	Throttle(sleepMilliseconds, appendTimestamp, delimiter)
}
