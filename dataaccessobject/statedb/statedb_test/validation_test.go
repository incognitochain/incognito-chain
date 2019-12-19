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
	err := statedb.SoValidation.ValidatePaymentAddressSanity(str1)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range receiverPaymentAddress[1:] {
		err := statedb.SoValidation.ValidatePaymentAddressSanity(v)
		if err != nil {
			t.Fatal(err)
		}
	}
	err = statedb.SoValidation.ValidatePaymentAddressSanity(str2)
	if err == nil {
		t.Fatal(err)
	}
	err = statedb.SoValidation.ValidatePaymentAddressSanity(str3)
	if err == nil {
		t.Fatal(err)
	}
	err = statedb.SoValidation.ValidatePaymentAddressSanity(str4)
	if err == nil {
		t.Fatal(err)
	}
	err = statedb.SoValidation.ValidatePaymentAddressSanity(str5)
	if err == nil {
		t.Fatal(err)
	}
}

func TestValidation_ValidateIncognitoPublicKeySanity(t *testing.T) {
	str1 := incognitoPublicKeys[0]
	str2 := str1[1:]
	str3 := str1[2:]
	str4 := str1[:len(str1)-1]
	str5 := str1[:len(str1)-2]
	err := statedb.SoValidation.ValidateIncognitoPublicKeySanity(str1)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range incognitoPublicKeys[1:] {
		err := statedb.SoValidation.ValidateIncognitoPublicKeySanity(v)
		if err != nil {
			t.Fatal(err)
		}
	}
	err = statedb.SoValidation.ValidateIncognitoPublicKeySanity(str2)
	if err == nil {
		t.Fatal(err)
	}
	err = statedb.SoValidation.ValidateIncognitoPublicKeySanity(str3)
	if err == nil {
		t.Fatal(err)
	}
	err = statedb.SoValidation.ValidateIncognitoPublicKeySanity(str4)
	if err == nil {
		t.Fatal(err)
	}
	err = statedb.SoValidation.ValidateIncognitoPublicKeySanity(str5)
	if err == nil {
		t.Fatal(err)
	}
}
