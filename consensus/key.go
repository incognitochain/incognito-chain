package consensus

import (
	"errors"
	"strings"
)

func (engine *Engine) LoadMiningKeys(keysString string) error {
	engine.userMiningPublicKeys = make(map[string]string)
	keys := strings.Split(keysString, "|")
	for _, key := range keys {
		keyParts := strings.Split(key, ":")
		if len(keyParts) == 2 {
			if _, ok := AvailableConsensus[keyParts[0]]; ok {
				err := AvailableConsensus[keyParts[0]].LoadUserKey(keyParts[1])
				if err != nil {
					panic(err)
				}
				engine.userMiningPublicKeys[keyParts[0]] = AvailableConsensus[keyParts[0]].GetUserPublicKey()
			} else {
				return errors.New("Consensus type for this key isn't exist " + keyParts[0])
			}
		}
	}
	return nil
}
func (engined *Engine) GetCurrentMiningPublicKey() (publickey string, keyType string) {
	if engined != nil && engined.CurrentMiningChain != "" {
		if _, ok := engined.ChainConsensusList[engined.CurrentMiningChain]; ok {
			return engined.ChainConsensusList[engined.CurrentMiningChain].GetUserPublicKey(), engined.ChainConsensusList[engined.CurrentMiningChain].GetConsensusName()
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
