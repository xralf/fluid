package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

func init() {
	log.SetOutput(os.Stdout) // Don't log to stderr by default
	//log.SetOutput(os.Stderr)
}

func main() {
	infiniteDialog()
	//acceleratingTicker()
}

// infiniteDialog waits for user input on STDIN and intermixes the lines entered with randlomly generated text lines.
// Every few moments, the writer function prints all messages comprising of the user input the bogus rando stuff gets printed.
//
// The reason for this function is to have a "thing" is a process that continuoaly reads from STDIn and WRITES to STDOUT.
func infiniteDialog() {
	quit := make(chan bool)
	msg := make(chan string)

	go func() { // writer function: waits for words produced by either the stdin reader function or the random word function
		millis := 1
		ticker := time.NewTicker(time.Duration(millis) * time.Millisecond)

		lines := "\n"
		for {
			select {
			case <-ticker.C:
				log.Print(lines)
				lines = "\n"
				ticker.Stop()
				millis = 1 + rand.Intn(4000) // average 2 seconds
				ticker = time.NewTicker(time.Duration(millis) * time.Millisecond)
			case line := <-msg:
				lines += line + "\n"
			case <-quit:
				ticker.Stop()
				log.Println("...reader ticker stopped!")
				return
			}
		}
	}()

	go func() { // stdin reader function: adds whatever is coming from the outside world through the msg channel
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			msg <- scanner.Text() + " <----------------------------- YOU wrote this!"
		}

		if err := scanner.Err(); err != nil {
			log.Println(err)
		}
	}()

	func() { // random word generator, simulates some stdin-like traffic for the msg channel
		millis := 1
		ticker := time.NewTicker(time.Duration(millis) * time.Millisecond)
		for range 100 {
			select {
			case <-ticker.C:
				msg <- "Random wrote: " + RandomWord(3)
				ticker.Stop()
				millis = 1 + rand.Intn(2000) // average 1s
				ticker = time.NewTicker(time.Duration(millis) * time.Millisecond)
			case <-quit:
				ticker.Stop()
				log.Println("...writer ticker stopped!")
				return
			}
		}
	}()

	time.Sleep(5 * time.Second)
	log.Println("stopping ticker...")
	quit <- true
	time.Sleep(500 * time.Millisecond) // just to see quit messages
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func RandomWord(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func acceleratingTicker() {
	start_interval := float64(1000)
	quit := make(chan bool)

	go func() {
		ticker := time.NewTicker(time.Duration(start_interval) * time.Millisecond)
		counter := 1.0

		for {
			select {
			case <-ticker.C:
				log.Println("ticker accelerating to " + fmt.Sprint(start_interval/counter) + " ms")
				ticker.Stop()
				ticker = time.NewTicker(time.Duration(start_interval/counter) * time.Millisecond)
				counter++
			case <-quit:
				ticker.Stop()
				log.Println("..ticker stopped!")
				return
			}
		}
	}()

	time.Sleep(5 * time.Second)

	log.Println("stopping ticker...")
	quit <- true

	time.Sleep(500 * time.Millisecond) // just to see quit messages
}
