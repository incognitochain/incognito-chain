package incognitokey

import (
	"encoding/json"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
)

func ExtractPublickeysFromCommitteeKeyList(keyList []CommitteePublicKey, keyType string) ([]string, error) {
	result := []string{}
	for _, keySet := range keyList {
		key := keySet.GetMiningKeyBase58(keyType)
		if key != "" {
			result = append(result, key)
		}
	}
	return result, nil
}

func CommitteeKeyListToString(keyList []CommitteePublicKey) ([]string, error) {
	result := []string{}
	for _, key := range keyList {
		keyStr, err := key.ToBase58()
		if err != nil {
			return nil, err
		}
		result = append(result, keyStr)
	}
	return result, nil
}

func CommitteeBase58KeyListToStruct(strKeyList []string) ([]CommitteePublicKey, error) {
	if len(strKeyList) == 0 {
		return []CommitteePublicKey{}, nil
	}
	if len(strKeyList) == 1 && len(strKeyList[0]) == 0 {
		return []CommitteePublicKey{}, nil
	}
	result := []CommitteePublicKey{}
	keyStruct := new(CommitteePublicKey)
	for _, key := range strKeyList {

		if err := keyStruct.FromString(key); err != nil {
			return nil, err
		}
		result = append(result, *keyStruct)
	}
	return result, nil
}

func Equal(keyString1 string, keyString2 string) bool {
	var pubKey1 CommitteePublicKey
	var pubKey2 CommitteePublicKey
	keyBytes1, ver, err := base58.Base58Check{}.Decode(keyString1)
	if (ver != common.ZeroByte) || (err != nil) {
		// errors.New("wrong input")
		return false
	}
	keyBytes2, ver, err := base58.Base58Check{}.Decode(keyString2)
	if (ver != common.ZeroByte) || (err != nil) {
		// return errors.New("wrong input")
		return false
	}
	err = json.Unmarshal(keyBytes1, pubKey1)
	if err != nil {
		// return errors.New("wrong input")
		return false
	}
	err = json.Unmarshal(keyBytes2, pubKey2)
	if err != nil {
		// return errors.New("wrong input")
		return false
	}
	if reflect.DeepEqual(pubKey1, pubKey2) {
		return true
	}
	return false
}

func IsOneMiner(keyString1 string, keyString2 string) bool {
	var pubKey1 CommitteePublicKey
	var pubKey2 CommitteePublicKey
	keyBytes1, ver, err := base58.Base58Check{}.Decode(keyString1)
	if (ver != common.ZeroByte) || (err != nil) {
		// errors.New("wrong input")
		return false
	}
	keyBytes2, ver, err := base58.Base58Check{}.Decode(keyString2)
	if (ver != common.ZeroByte) || (err != nil) {
		// return errors.New("wrong input")
		return false
	}
	err = json.Unmarshal(keyBytes1, pubKey1)
	if err != nil {
		// return errors.New("wrong input")
		return false
	}
	err = json.Unmarshal(keyBytes2, pubKey2)
	if err != nil {
		// return errors.New("wrong input")
		return false
	}
	if reflect.DeepEqual(pubKey1.MiningPubKey, pubKey2.MiningPubKey) {
		return true
	}
	return false
}
