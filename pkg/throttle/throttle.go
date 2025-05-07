package throttle

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

func Throttle(sleepMilliseconds int, appendTimestamp bool, delimiter string) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if err = scanner.Err(); err != nil {
			panic(err)
		}

		if appendTimestamp {
			ts := time.Now()
			line += delimiter + fmt.Sprint(ts.UTC().Format(time.RFC3339Nano))
		}
		fmt.Println(line)
		time.Sleep(time.Duration(sleepMilliseconds) * time.Millisecond)
	}
}
