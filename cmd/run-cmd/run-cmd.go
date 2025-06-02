package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

func main() {
	run()
	//runMaybeTooMuch()
}

func run() {
	var err error
	cmd := exec.Command("./cmd.sh")
	cmd.Stderr = os.Stderr

	var cmdStdin io.WriteCloser
	if cmdStdin, err = cmd.StdinPipe(); err != nil {
		log.Fatalf("Error obtaining stdin: %s", err.Error())
	}

	localStdin := bufio.NewReader(os.Stdin)
	go func() {
		for {
			s, _ := localStdin.ReadString('\n')
			s = s[:len(s)-1]
			fmt.Fprintln(cmdStdin, s)
		}
	}()

	if err := cmd.Start(); err != nil {
		log.Fatalf("Error starting program: %s, %s", cmd.Path, err.Error())
	}

	cmd.Wait()
}

func runAndAlsoShowStdout() {
	var err error
	cmd := exec.Command("./cmd.sh")
	cmd.Stderr = os.Stderr

	var cmdStdin io.WriteCloser
	if cmdStdin, err = cmd.StdinPipe(); err != nil {
		log.Fatalf("Error obtaining stdin: %s", err.Error())
	}

	var cmdStdout io.ReadCloser
	if cmdStdout, err = cmd.StdoutPipe(); err != nil {
		log.Fatalf("Error obtaining stdout: %s", err.Error())
	}

	stdoutReader := bufio.NewReader(cmdStdout)
	stdinReader := bufio.NewReader(os.Stdin)
	go func(reader io.Reader) {
		var s string
		for {
			s, _ = stdinReader.ReadString('\n')
			s = s[:len(s)-1]
			fmt.Fprintln(cmdStdin, s)
		}
	}(stdoutReader)

	if err := cmd.Start(); err != nil {
		log.Fatalf("Error starting program: %s, %s", cmd.Path, err.Error())
	}

	cmd.Wait()
}
