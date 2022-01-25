package incognitokey

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"testing"
)

func TestOTDepositKey_UnmarshalJSON(t *testing.T) {
	privateKey := common.RandBytes(32)
	tokenID := common.HashH(common.RandBytes(32)).String()

	for index := uint64(0); index < 100000; index++ {
		depositKey, err := GenerateOTDepositKeyFromPrivateKey(privateKey, tokenID, index)
		if err != nil {
			panic(err)
		}
		jsb, err := json.Marshal(depositKey)
		if err != nil {
			panic(err)
		}

		recoveredKey := new(OTDepositKey)
		err = json.Unmarshal(jsb, recoveredKey)
		if err != nil {
			panic(err)
		}

		if recoveredKey.Index != depositKey.Index ||
			!bytes.Equal(recoveredKey.PublicKey, depositKey.PublicKey) ||
			!bytes.Equal(recoveredKey.PrivateKey, depositKey.PrivateKey) {
			panic(fmt.Sprintf("recovered: %v, actual %v\n", recoveredKey, depositKey))
		}
	}
}

func TestKeySet_GenerateOTDepositKey(t *testing.T) {
	privateKey := []byte{207, 198, 249, 34, 212, 217, 135, 61, 100, 20, 255, 233, 108, 181, 120, 239, 22, 48, 142, 220, 11, 18, 29, 33, 153, 20, 220, 65, 143, 26, 88, 48}
	keySet := new(KeySet)
	err := keySet.InitFromPrivateKeyByte(privateKey)
	if err != nil {
		panic(err)
	}

	tokenID := "b832e5d3b1f01a4f0623f7fe91d6673461e1f5d37d91fe78c5c2e6183ff39696"
	for i := uint64(0); i < 100; i++ {
		otDepositKey, err := keySet.GenerateOTDepositKey(tokenID, i)
		if err != nil {
			panic(err)
		}
		jsb, _ := json.Marshal(otDepositKey)
		fmt.Printf("Index: %v, OTDepositKey: %v\n\n", i, string(jsb))
	}
}

func TestGenerateOTDepositKeyFromPrivateKey(t *testing.T) {
	privateKey := common.RandBytes(32)
	keySet := new(KeySet)
	err := keySet.InitFromPrivateKeyByte(privateKey)
	if err != nil {
		panic(err)
	}

	tokenID := "b832e5d3b1f01a4f0623f7fe91d6673461e1f5d37d91fe78c5c2e6183ff39696"
	for i := uint64(0); i < 1000000; i++ {
		if i%1000 == 0 {
			fmt.Printf("FINISHED %v\n", i)
		}
		otDepositKey, err := keySet.GenerateOTDepositKey(tokenID, i)
		if err != nil {
			panic(err)
		}
		jsb, _ := json.Marshal(otDepositKey)

		otDepositKeyFromPrivateKey, err := GenerateOTDepositKeyFromPrivateKey(privateKey, tokenID, i)
		if err != nil {
			panic(err)
		}
		jsb2, err := json.Marshal(otDepositKeyFromPrivateKey)
		if !bytes.Equal(jsb, jsb2) || len(jsb)+len(jsb2) == 0 {
			panic(fmt.Errorf("expected %v, got %v", jsb, jsb2))
		}
	}
}

func TestGenerateOTDepositKeyFromMasterDepositSeed(t *testing.T) {
	privateKey := common.RandBytes(32)
	keySet := new(KeySet)
	err := keySet.InitFromPrivateKeyByte(privateKey)
	if err != nil {
		panic(err)
	}

	tokenIDStr := "b832e5d3b1f01a4f0623f7fe91d6673461e1f5d37d91fe78c5c2e6183ff39696"
	tokenID, _ := new(common.Hash).NewHashFromStr(tokenIDStr)
	tmp := append([]byte(common.PortalV4DepositKeyGenSeed), tokenID[:]...)
	masterDepositSeed := common.SHA256(append(privateKey, tmp...))

	for i := uint64(0); i < 1000000; i++ {
		if i%1000 == 0 {
			fmt.Printf("FINISHED %v\n", i)
		}
		otDepositKey, err := keySet.GenerateOTDepositKey(tokenIDStr, i)
		if err != nil {
			panic(err)
		}
		jsb, _ := json.Marshal(otDepositKey)

		otDepositKeyFromPrivateKey, err := GenerateOTDepositKeyFromMasterDepositSeed(masterDepositSeed, i)
		if err != nil {
			panic(err)
		}
		jsb2, err := json.Marshal(otDepositKeyFromPrivateKey)
		if !bytes.Equal(jsb, jsb2) || len(jsb)+len(jsb2) == 0 {
			panic(fmt.Errorf("expected %v, got %v", jsb, jsb2))
		}
	}
}
