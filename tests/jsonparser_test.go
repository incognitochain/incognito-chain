package main

import (
	"testing"
)

func TestReadFile(t *testing.T) {
	ok := readfile("./testsdata/transaction.json")
	if !ok {
		t.Fatal()
	}
}