package main

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/MaxSchaefer/macos-log-stream/pkg/mls"
)

func main() {
	name := runtime.GOOS
	switch name {
	case "darwin":
		darwinLogWorker()
	case "linux":
		//linuxLogWorker(conn, ticker, done)
	default:
		panic(fmt.Errorf("operation system logs not supported"))
	}
}

func darwinLogWorker() {
	logs := mls.NewLogs()
	if err := logs.StartGathering(); err != nil {
		panic(err)
	}

	ticker := time.NewTicker(1 * time.Second)
	done := make(chan bool)

	go func() {
		for line := range logs.Channel {
			select {
			case <-done:
				return
			default:
				fmt.Printf("%s|%s|%s|%s|%s\n", toRFC3339Nano(line.Timestamp), strconv.Itoa(line.ProcessID), line.Subsystem, line.MessageType, line.ProcessImagePath)
			}
		}
	}()

	time.Sleep(3600 * time.Second)
	ticker.Stop()
	done <- true
	fmt.Println("Stopped due to timeout!")
}

// Input:  2024-02-20 18:20:33.898369-0800 (timestamp format of the "mls" package)
// Output: 2024-02-20T18:20:33.898369-08:00 (RFC3339Nano)
func toRFC3339Nano(s string) string {
	var t time.Time
	var err error
	if t, err = time.Parse("2006-01-02T15:04:05.999999999-0700", strings.Replace(s, " ", "T", -1)); err != nil {
		panic(err)
	}
	return t.Format(time.RFC3339Nano)
}
