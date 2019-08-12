package consensus

import (
	"errors"
	"fmt"
	"strings"
)

func (engine *Engine) LoadMiningKeys(keysString string) error {
	keys := strings.Split(keysString, "|")
	for _, key := range keys {
		keyParts := strings.Split(key, ":")
		if len(keyParts) == 2 {
			if _, ok := engine.MiningKeys[keyParts[0]]; !ok {
				if _, ok := AvailableConsensus[keyParts[0]]; ok {
					engine.MiningKeys[keyParts[0]] = keyParts[1]
				} else {
					return errors.New("Consensus type for this key isn't exist")
				}
			} else {
				return errors.New("Only one mining key per consensus type")
			}
		}
	}
	fmt.Println(engine.MiningKeys)
	panic(1)
	return nil
}
func (engine *Engine) GetMiningPublicKey() (publickey string, keyType string) {
	return "", ""
}
func (engine *Engine) SignDataWithMiningKey(data []byte) (string, error) {
	return "", nil
}
func (engine *Engine) VerifyValidationData(data []byte, validationData string, consensusType string) error {
	return nil
}
