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
func IsInBase58ShortFormat(strKeyList []string) bool {
	tempStruct, err := CommitteeBase58KeyListToStruct(strKeyList)
	if err != nil {
		return false
	}
	tempString, err := CommitteeKeyListToString(tempStruct)
	if len(tempString) != len(strKeyList) {
		return false
	}
	for index, value := range tempString {
		if value != strKeyList[index] {
			return false
		}
	}
	return true
}

func ConvertToBase58ShortFormat(strKeyList []string) ([]string, error) {
	tempStruct, err := CommitteeBase58KeyListToStruct(strKeyList)
	if err != nil {
		return []string{}, err
	}
	tempString, err := CommitteeKeyListToString(tempStruct)
	if err != nil {
		return []string{}, err
	}
	return tempString, nil
}
