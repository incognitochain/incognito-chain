package blsbft

import "github.com/incognitochain/incognito-chain/common"

func (e BLSBFT) LoadUserKey(string) error {
	return nil
}
func (e BLSBFT) GetUserPublicKey() string {
	return ""
}
func (e BLSBFT) GetUserPrivateKey() string {
	return ""
}

func (e BLSBFT) SignData(data []byte) (string, error) {
	return "", nil
}
func (e BLSBFT) ValidateAggregatedSig(dataHash *common.Hash, aggSig string, validatorPubkeyList []string) error {
	return nil
}
func (e BLSBFT) ValidateSingleSig(dataHash *common.Hash, sig string, pubkey string) error {
	return nil
}
