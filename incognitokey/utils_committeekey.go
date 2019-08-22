package incognitokey

func ExtractPublickeysFromCommitteeKeyList(keyList []CommitteePubKey, keyType string) ([]string, error) {
	result := []string{}
	for _, keySet := range keyList {
		key := keySet.GetMiningKeyBase58(keyType)
		if key != "" {
			result = append(result, key)
		}
	}
	return result, nil
}

func CommitteeKeyListToString(keyList []CommitteePubKey) []string {
	result := []string{}
	for _, key := range keyList {
		keyStr, _ := key.ToBase58()
		result = append(result, keyStr)
	}
	return result
}

func CommitteeBase58KeyListToStruct(strKeyList []string) []CommitteePubKey {
	result := []CommitteePubKey{}
	for _, key := range strKeyList {
		var keyStruct CommitteePubKey
		keyStruct.FromBase58(key)
		result = append(result, keyStruct)
	}
	return result
}
