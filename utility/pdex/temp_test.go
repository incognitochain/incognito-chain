package main

import (
	"testing"

	"github.com/incognitochain/incognito-chain/privacy"
)

func TestOTAReceiver(t *testing.T) {
	otaReceiver1 := privacy.OTAReceiver{}
	otaReceiver1.FromString("15vMwzt9wXbdf8KkNckPSXoB1pM2W27n1mdQspbHJXKZvFpTjziM86eaS2uC3bX6SVteNHdKXpd3jzDgew4iahyZimSb4NtDNXxdZQaugu5bBTK5bxwKJNtoPMHxnJMZweWDXQyCEMzJ6Xfe")
	otaReceiver2 := privacy.OTAReceiver{}
	otaReceiver2.FromString("16AfhmWxauFq9h8BcwFjaK2YUvwjfVYQn8pyQxjN86Pjx3NAUpEwh4ZKCyUfgdzdkuKTaXjJE4rjt2HsThBRMqCoLuGJdTPCuvKpurAFwrLGHtiTEw1YDL1ps4cb3P4EUcseNrfey2G6Nqwt")
	t.Logf("public key %s:", otaReceiver1.PublicKey.String())
	t.Logf("public key %s:", otaReceiver2.PublicKey.String())
}
