//nolint:revive // skip linter for this package name
package privacy_v2

import (
	"errors"

	"github.com/incognitochain/incognito-chain/privacy/env"
	"github.com/incognitochain/incognito-chain/privacy/operation"
)

func (proof PaymentProofV2) VerifyV2(vEnv env.ValidationEnviroment, fee uint64) (bool, error) {
	hasConfidentialAsset := vEnv.HasCA()
	inputCoins := proof.GetInputCoins()
	dupMap := make(map[[operation.Ed25519KeySize]byte]bool)
	for _, coin := range inputCoins {
		identifier := coin.GetKeyImage().ToBytes()
		_, exists := dupMap[identifier]
		if exists {
			return false, errors.New("Duplicate input coin in PaymentProofV2")
		}
		dupMap[identifier] = true
	}

	if !hasConfidentialAsset {
		return proof.verifyHasNoCA(false)
	}
	return proof.verifyHasConfidentialAsset(false)
}
