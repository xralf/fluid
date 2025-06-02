package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
)

func main() {
	var out bytes.Buffer
	var err error

	cmd := exec.Command("make", "all", "ENGINE_DEST=/tmp")
	cmd.Dir = "../fluid"
	cmd.Stdout = &out
	if err = cmd.Run(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Output: %s\n", out.String())
}
