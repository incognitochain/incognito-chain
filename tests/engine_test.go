package main

import (
	"log"
	"testing"
)

func TestExecuteTest(t *testing.T) {
	res, err := executeTest("./testsdata/sample.json")
	log.Println(res,err)
}