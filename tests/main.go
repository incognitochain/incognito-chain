package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	switch os.Args[1] {
	case "all":
		fmt.Println("abc")
	case "client":
		cmd := exec.Command("go test -run", "client_test.go")
		err := cmd.Run()
		if err != nil {
			log.Println("Failed to run test client", err)
		} else {
			log.Println("Begin to run get test client")
		}
	default:
		log.Println("Please choose the right test to run")
	}
}