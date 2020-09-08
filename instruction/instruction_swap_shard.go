package instruction

import "github.com/incognitochain/incognito-chain/incognitokey"

type SwapShardInstruction struct {
	InPublicKeys        []string
	InPublicKeyStructs  []incognitokey.CommitteePublicKey
	OutPublicKeys       []string
	OutPublicKeyStructs []incognitokey.CommitteePublicKey
	ChainID             int
	Height              uint64
}
