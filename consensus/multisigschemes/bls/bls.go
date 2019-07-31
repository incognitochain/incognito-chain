package bls

import "github.com/incognitochain/incognito-chain/common"

func (keyset *KeySet) SignData(privKey string, dataHash common.Hash) string {
	return ""
}

func ValidateAggSig(dataHash common.Hash, aggSig string, validatorPubkeyList []string) error {
	return nil
}
func ValidateSingleSig(dataHash *common.Hash, sig string, pubkey string) error {
	return nil
}
func AggregateSig() string {
	return ""
}
