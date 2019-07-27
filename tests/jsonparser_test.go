package main

import (
	"log"
	"testing"
)

func TestReadFile(t *testing.T) {
	res, err := readfile("./testsdata/transaction.json")
	if err != nil {
		t.Fatal()
	}
	log.Println(res)
}