package blsbft

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
)

type ValidationData struct { 
	Producer    string
	ProducerSig string
	AggSig      string
}

func DecodeValidationData(data string) (&ValidationData, error) {
	var valData ValidationData
	err := json.Unmarshal([]byte(data), &valData)
	if err != nil {
		return nil, err
	}
	return valData, nil
}

func EncodeValidationData(validationData ValidationData) (string,error){
	return json.Marshal(validationData)
}
func (e *BLSBFT) validatePreSignBlock(blockHash common.Hash, validationData string) error {
	valData, err := DecodeValidationData(validationData)
	return nil
}

func (e *BLSBFT) ValidateBlock(blockHash common.Hash, validationData string) error {
	valData, error := DecodeValidationData(validationData)
	return nil

}

func (e *BLSBFT) CreateValidationData(blockHash common.Hash,privateKey string, round int) ValidationData{
	var valData ValidationData
	return valData
}

func (e *BLSBFT) FinalizedValidationData(block chain.BlockInterface, sigs []string) error{
return nil
} 