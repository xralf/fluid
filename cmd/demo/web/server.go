// websockets.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/websocket"
)

func main() {
	usage := fmt.Sprintf("usage: %v <port>, e.g., %v 8090\n", os.Args[0], os.Args[0])

	if len(os.Args) != 2 {
		fmt.Print(usage)
		return
	}

	var port int
	var err error

	if port, err = strconv.Atoi(os.Args[1]); err != nil {
		fmt.Print(usage)
		panic(err)
	}

	run(port)
}

func run(port int) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		var conn *websocket.Conn
		var err error
		if conn, err = upgrader.Upgrade(w, r, nil); err != nil {
			log.Println(err)
			return
		}

		var msgType int
		var msg []byte
		for {
			// Read message from browser
			if msgType, msg, err = conn.ReadMessage(); err != nil {
				log.Println(err)
				return
			}

			// Print the message to the console
			fmt.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))

			// Write message back to browser
			if err = conn.WriteMessage(msgType, msg); err != nil {
				log.Println(err)
				return
			}
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "dashboard.html")
	})

	addr := fmt.Sprintf(":%d", port)
	http.ListenAndServe(addr, nil)
}
