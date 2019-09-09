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
	for _, key := range strKeyList {
		var keyStruct CommitteePublicKey
		if err := keyStruct.FromBase58(key); err != nil {
			return nil, err
		}
		result = append(result, keyStruct)
	}
	return result, nil
}
