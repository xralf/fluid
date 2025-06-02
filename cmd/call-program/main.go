package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

func main() {
	test1()
	//test2()
}

func test1() {
	cmd := exec.Command("./my-echo.sh")

	cmd.Stderr = os.Stderr

	var stdin io.WriteCloser
	var err error
	if stdin, err = cmd.StdinPipe(); err != nil {
		log.Fatalf("Error obtaining stdin: %s", err.Error())
	}

	var stdout io.ReadCloser
	if stdout, err = cmd.StdoutPipe(); err != nil {
		log.Fatalf("Error obtaining stdout: %s", err.Error())
	}

	reader := bufio.NewReader(stdout)
	go func(reader io.Reader) {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			log.Printf("Reading from subprocess: %s", scanner.Text())
			stdin.Write([]byte("blah\n"))
		}
	}(reader)

	if err := cmd.Start(); err != nil {
		log.Fatalf("Error starting program: %s, %s", cmd.Path, err.Error())
	}

	cmd.Wait()
}

func test2() {
	cmd := exec.Command("./my-echo.sh")

	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if nil != err {
		log.Fatalf("Error obtaining stdin: %s", err.Error())
	}

	stdout, err := cmd.StdoutPipe()
	if nil != err {
		log.Fatalf("Error obtaining stdout: %s", err.Error())
	}

	reader := bufio.NewReader(stdout)

	go func(reader io.Reader) {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			log.Printf("Reading from subprocess: %s", scanner.Text())
			stdin.Write([]byte("some sample text\n"))
		}
	}(reader)

	if err := cmd.Start(); nil != err {
		log.Fatalf("Error starting program: %s, %s", cmd.Path, err.Error())
	}

	cmd.Wait()
}

func test3() {
	//cmd := exec.Command("ls", "-lah")
	cmd := exec.Command("htop")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	var err error
	if err = cmd.Run(); err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	outStr, errStr := stdout.String(), stderr.String()
	fmt.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)
}
