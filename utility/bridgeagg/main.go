package main

import (
	"encoding/base64"
	"fmt"
	"os"
)

func main() {
	contentBytes, err := base64.StdEncoding.DecodeString(os.Args[1])
	if err != nil {
		panic(err)
	}
	fmt.Println(string(contentBytes))
}
