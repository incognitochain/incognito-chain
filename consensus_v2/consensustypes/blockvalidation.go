package consensustypes

import (
	"encoding/json"

	portalprocessv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portalprocess"
)

type ValidationData struct {
	ProducerBLSSig []byte
	ProducerBriSig []byte
	ValidatiorsIdx []int
	AggSig         []byte
	BridgeSig      [][]byte
	PortalSig      []*portalprocessv4.PortalSig
}

func DecodeValidationData(data string) (*ValidationData, error) {
	var valData ValidationData
	err := json.Unmarshal([]byte(data), &valData)
	if err != nil {
		return nil, err
	}
	return &valData, nil
}

func EncodeValidationData(validationData ValidationData) (string, error) {
	result, err := json.Marshal(validationData)
	if err != nil {
		return "", err
	}
	return string(result), nil
}
