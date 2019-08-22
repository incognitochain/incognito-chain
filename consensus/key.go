package consensus

import (
	"errors"
	"strings"

	"github.com/incognitochain/incognito-chain/incognitokey"
)

func (engine *Engine) LoadMiningKeys(keysString string) error {
	engine.userMiningPublicKeys = make(map[string]incognitokey.CommitteePubKey)
	keys := strings.Split(keysString, "|")
	for _, key := range keys {
		keyParts := strings.Split(key, ":")
		if len(keyParts) == 2 {
			if _, ok := AvailableConsensus[keyParts[0]]; ok {
				err := AvailableConsensus[keyParts[0]].LoadUserKey(keyParts[1])
				if err != nil {
					panic(err)
				}
				engine.userMiningPublicKeys[keyParts[0]] = *AvailableConsensus[keyParts[0]].GetUserPublicKey()
			} else {
				return errors.New("Consensus type for this key isn't exist " + keyParts[0])
			}
		}
	}
	return nil
}
func (engine *Engine) GetCurrentMiningPublicKey() (publickey string, keyType string) {
	if engine != nil && engine.CurrentMiningChain != "" {
		if _, ok := engine.ChainConsensusList[engine.CurrentMiningChain]; ok {
			keytype := engine.ChainConsensusList[engine.CurrentMiningChain].GetConsensusName()
			pubkey := engine.userMiningPublicKeys[keytype]
			return pubkey.GetMiningKeyBase58(keytype), keytype
		}
	}
	return "", ""
}
func (engine *Engine) SignDataWithMiningKey(data []byte) (string, error) {
	return "", nil
}
func (engine *Engine) VerifySignature(data []byte, validationData string, consensusType string) error {
	return nil
}
