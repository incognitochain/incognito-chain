package coin

import "github.com/incognitochain/incognito-chain/privacy/pedersen"

const (
	MaxSizeInfoCoin         = 255
	PedersenPrivateKeyIndex = pedersen.PedersenPrivateKeyIndex
	PedersenValueIndex      = pedersen.PedersenValueIndex
	PedersenSndIndex        = pedersen.PedersenSndIndex
	PedersenShardIDIndex    = pedersen.PedersenShardIDIndex
	PedersenRandomnessIndex = pedersen.PedersenRandomnessIndex
)

var PedCom pedersen.PedersenCommitment = pedersen.PedCom
