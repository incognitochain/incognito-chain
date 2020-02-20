package wrapper

import (
	"fmt"
	"testing"
)

type StTest struct {
	X string
	Y byte
	Z int
}

func TestWrapper(t *testing.T) {
	w, err := NewWrapper()
	if err != nil {
		t.Error(err)
	}
	oData := &StTest{
		X: "aaaaaaaa",
		Y: 0,
		Z: 9,
	}
	e, err := w.EnCom(oData)
	if err != nil {
		t.Error(err)
	}
	d := new(StTest)
	err = w.DeCom(e, d)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(len(e), d)
	e2, err := w.EnCom(oData)
	if err != nil {
		t.Error(err)
	}
	d2 := new(StTest)
	err = w.DeCom(e2, d2)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(len(e2), d2)
}
