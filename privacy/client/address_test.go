package client

import (
	"fmt"
	"testing"
)

func TestRandSpendingKey(t *testing.T) {
	a := [64]byte{192}
	if a[len(a)-1] > 0x0F {
		t.Errorf("SpendingKey first 4 bytes are not 0: %x", a)
	}
}

func TestReceivingKey(t *testing.T) {
	a := [32]byte{1, 2, 3}
	var ask SpendingKey
	copy(ask[:], a[:])
	pkenc := GenReceivingKey(ask)
	e := [32]byte{0, 155, 240, 217, 29, 179, 89, 156, 13, 30, 125, 31, 108, 85, 76, 38, 151, 249, 116, 112, 139, 98, 32, 138, 213, 140, 105, 91, 153, 5, 101, 113}

	for i, v := range e {
		if v != pkenc[i] {
			t.Errorf("Rk incorrect:\nExpected: %x\n Received: %x\n", e, pkenc)
		}
	}
}

func TestPaymentAddress(t *testing.T) {
	fmt.Println("TestPaymentAddress")
	a := [32]byte{1, 2, 3, 4, 5, 6}
	var ask SpendingKey
	copy(ask[:], a[:])
	addr := GenPaymentAddress(ask)
	expSpendingAddress := [...]byte{114, 113, 96, 254, 168, 25, 103, 142, 89, 177, 31, 92, 44, 151, 129, 185, 144, 154, 61, 208, 249, 213, 2, 135, 60, 6, 67, 42, 57, 5, 59, 135}
	expTransmissionKey := [...]byte{151, 101, 142, 174, 19, 254, 217, 234, 63, 192, 81, 135, 96, 114, 181, 206, 51, 134, 131, 166, 30, 106, 238, 242, 67, 64, 116, 37, 39, 52, 34, 19}

	for i, v := range expSpendingAddress {
		if v != addr.Apk[i] {
			t.Errorf("SpendingAddress incorrect:\nExpected: %x\n Received: %x\n", expSpendingAddress, addr.Apk)
		}
	}

	for i, v := range expTransmissionKey {
		if v != addr.Pkenc[i] {
			t.Errorf("SpendingAddress incorrect:\nExpected: %x\n Received: %x\n", expTransmissionKey, addr.Pkenc)
		}
	}
}
