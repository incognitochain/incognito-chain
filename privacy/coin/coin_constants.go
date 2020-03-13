package coin

import "github.com/incognitochain/incognito-chain/privacy/operation"

const (
	MaxSizeInfoCoin         = 255
	PedersenPrivateKeyIndex = operation.PedersenPrivateKeyIndex
	PedersenValueIndex      = operation.PedersenValueIndex
	PedersenSndIndex        = operation.PedersenSndIndex
	PedersenShardIDIndex    = operation.PedersenShardIDIndex
	PedersenRandomnessIndex = operation.PedersenRandomnessIndex
)

var PedCom operation.PedersenCommitment = operation.PedCom

func getMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}
