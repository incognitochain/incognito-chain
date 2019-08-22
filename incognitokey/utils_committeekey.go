package incognitokey

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

func CommitteeKeyListToString(keyList []CommitteePublicKey) []string {
	result := []string{}
	for _, key := range keyList {
		keyStr, _ := key.ToBase58()
		result = append(result, keyStr)
	}
	return result
}

func CommitteeBase58KeyListToStruct(strKeyList []string) []CommitteePublicKey {
	result := []CommitteePublicKey{}
	for _, key := range strKeyList {
		var keyStruct CommitteePublicKey
		keyStruct.FromBase58(key)
		result = append(result, keyStruct)
	}
	return result
}
