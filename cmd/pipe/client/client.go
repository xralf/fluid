package main

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"slices"

	"github.com/gorilla/websocket"
)

var (
	stdin = bufio.NewReader(os.Stdin)
)

func main() {
	usage := fmt.Sprintf("usage: %v <server> (in|out), e.g., %v localhost:5801 in\n", os.Args[0], os.Args[0])

	if len(os.Args) < 3 {
		fmt.Print(usage)
		return
	}
	server := os.Args[1]
	path := os.Args[2]

	allowedPaths := []string{"in", "out"}
	if !slices.Contains(allowedPaths, path) {
		fmt.Print(usage)
		return
	}
	fmt.Println("Connecting to:", server, "at", path)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	input := make(chan string, 1)
	go getInput(input)
	url := url.URL{Scheme: "ws", Host: server, Path: path}

	var conn *websocket.Conn
	var err error
	if conn, _, err = websocket.DefaultDialer.Dial(url.String(), nil); err != nil {
		log.Println("Error:", err)
		return
	}
	defer conn.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		var message []byte
		for {
			if _, message, err = conn.ReadMessage(); err != nil {
				log.Println("ReadMessage() error:", err)
				return
			}

			log.Printf("Received: %s", message)
		}
	}()

	for {
		select {
		case <-done:
			return
		case t := <-input:
			if err = conn.WriteMessage(websocket.TextMessage, []byte(t)); err != nil {
				log.Println("Write error:", err)
				return
			}
			go getInput(input)
		case <-interrupt:
			log.Println("Caught interrupt signal - quitting!")
			if err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
				log.Println("Write close error:", err)
				return
			}
			return
		}
	}
}

func getInput(input chan string) {
	var line string
	var err error
	if line, err = stdin.ReadString('\n'); err != nil {
		log.Println(err)
		return
	}
	input <- line
}
