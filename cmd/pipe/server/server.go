package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/websocket"
)

var (
	inConn, outConn   *websocket.Conn
	inReady, outReady chan struct{}
)

func main() {
	usage := fmt.Sprintf("usage: %v <input-port> <output-port>, e.g., %v 7001 7002\n", os.Args[0], os.Args[0])

	if len(os.Args) != 3 {
		fmt.Print(usage)
		return
	}

	var inPort, outPort int
	var err error

	if inPort, err = strconv.Atoi(os.Args[1]); err != nil {
		fmt.Print(usage)
		panic(err)
	}
	if outPort, err = strconv.Atoi(os.Args[2]); err != nil {
		fmt.Print(usage)
		panic(err)
	}

	run(inPort, outPort)
}

func run(inPort, outPort int) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	inReady = make(chan struct{})
	outReady = make(chan struct{})

	go listen()

	http.HandleFunc("/in", func(w http.ResponseWriter, r *http.Request) {
		// Prevent this error: "websocket: request origin not allowed by Upgrader.CheckOrigin"
		// See: https://github.com/gorilla/websocket/blob/main/server.go
		upgrader.CheckOrigin = nil
		r.Header["Origin"] = []string{}

		var err error
		if inConn, err = upgrader.Upgrade(w, r, nil); err != nil {
			log.Println(err)
			return
		}
		inReady <- struct{}{}
	})

	http.HandleFunc("/out", func(w http.ResponseWriter, r *http.Request) {
		upgrader.CheckOrigin = nil
		r.Header["Origin"] = []string{}

		var err error
		if outConn, err = upgrader.Upgrade(w, r, nil); err != nil {
			log.Println(err)
			return
		}
		outReady <- struct{}{}
	})

	inAddr := fmt.Sprintf(":%d", inPort)
	outAddr := fmt.Sprintf(":%d", outPort)

	go http.ListenAndServe(inAddr, nil)
	http.ListenAndServe(outAddr, nil)
}

func listen() {
	for i := 0; i < 2; i++ {
		select {
		case <-inReady:
			log.Println("Pipe input client connected.")
		case <-outReady:
			log.Println("Pipe output client connected.")
		}
	}

	log.Println("All clients are ready!  The reader client can send some data now.")

	go func() {
		var messageType int
		var messageContent []byte
		var err error
		for {
			if messageType, messageContent, err = inConn.ReadMessage(); err != nil {
				log.Println(err)
				return
			}
			if err = outConn.WriteMessage(messageType, messageContent); err != nil {
				log.Println(err)
				return
			}
		}
	}()
}
