package statedb_test

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"testing"
)

func TestValidation_ValidatePaymentAddressSanity(t *testing.T) {
	str1 := receiverPaymentAddress[0]
	str2 := str1[1:]
	str3 := str1[2:]
	str4 := str1[:len(str1)-1]
	str5 := str1[:len(str1)-2]
	err := statedb.ValidatePaymentAddressSanity(str1)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range receiverPaymentAddress[1:] {
		err := statedb.ValidatePaymentAddressSanity(v)
		if err != nil {
			t.Fatal(err)
		}
	}
	err = statedb.ValidatePaymentAddressSanity(str2)
	if err == nil {
		t.Fatal(err)
	}
	err = statedb.ValidatePaymentAddressSanity(str3)
	if err == nil {
		t.Fatal(err)
	}
	err = statedb.ValidatePaymentAddressSanity(str4)
	if err == nil {
		t.Fatal(err)
	}
	err = statedb.ValidatePaymentAddressSanity(str5)
	if err == nil {
		t.Fatal(err)
	}
}

func TestValidation_ValidateIncognitoPublicKeySanity(t *testing.T) {
	str1 := incognitoPublicKey[0]
	str2 := str1[1:]
	str3 := str1[2:]
	str4 := str1[:len(str1)-1]
	str5 := str1[:len(str1)-2]
	err := statedb.ValidateIncognitoPublicKeySanity(str1)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range incognitoPublicKey[1:] {
		err := statedb.ValidateIncognitoPublicKeySanity(v)
		if err != nil {
			t.Fatal(err)
		}
	}
	err = statedb.ValidateIncognitoPublicKeySanity(str2)
	if err == nil {
		t.Fatal(err)
	}
	err = statedb.ValidateIncognitoPublicKeySanity(str3)
	if err == nil {
		t.Fatal(err)
	}
	err = statedb.ValidateIncognitoPublicKeySanity(str4)
	if err == nil {
		t.Fatal(err)
	}
	err = statedb.ValidateIncognitoPublicKeySanity(str5)
	if err == nil {
		t.Fatal(err)
	}
}
