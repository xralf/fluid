package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/xralf/fluid/pkg/model"
)

// func Run(port int, executablePath string) {
func Run(job model.Job) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		websocket, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Websocket Connected!")
		listen(websocket, job)
	})
	addr := fmt.Sprintf(":%d", job.Pipe1IngressPort) // example ":9001"
	http.ListenAndServe(addr, nil)
}

func listen(conn *websocket.Conn, job model.Job) {
	var messageType int
	var messageContent []byte
	var err error

	//cmd := exec.Command("./cmd.sh")
	//cmd := exec.Command("cmd/ws-engine/server/cmd.sh")
	//cmd := exec.Command(executablePath)
	cmd := exec.Command(job.EnginePath, "-p", job.BinaryPlanPath, "-x", strconv.Itoa(job.ExitAfterSeconds))
	cmd.Stderr = os.Stderr

	var cmdStdin io.WriteCloser
	if cmdStdin, err = cmd.StdinPipe(); err != nil {
		log.Fatalf("Error obtaining stdin: %s", err.Error())
	}

	go func() { // writer: websocket -> command
		for {
			if messageType, messageContent, err = conn.ReadMessage(); err != nil {
				log.Println(err)
				return
			}
			fmt.Fprint(cmdStdin, string(messageContent))
		}
	}()

	var cmdStdout io.ReadCloser
	if cmdStdout, err = cmd.StdoutPipe(); err != nil {
		log.Fatalf("Error obtaining stdout: %s", err.Error())
	}
	stdoutReader := bufio.NewReader(cmdStdout)

	go func(reader io.Reader) { // reader: command -> websocket
		bytes := make([]byte, 1024)
		var n int
		var err error
		for {
			if n, err = reader.Read(bytes); err != nil {
				if err == io.EOF {
					break
				}
				log.Println(err)
				return
			}

			// Remove the last symbol if it's a newline
			if n > 0 {
				if bytes[n-1] == '\n' {
					bytes = bytes[:n-1]
				}

				if err = conn.WriteMessage(messageType, bytes); err != nil {
					log.Println(err)
					return
				}
			}
		}
	}(stdoutReader)

	if err := cmd.Start(); err != nil {
		log.Fatalf("Error starting program: %s, %s", cmd.Path, err.Error())
	}

	cmd.Wait()
}
