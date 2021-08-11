package pdexv3

import "github.com/incognitochain/incognito-chain/common"

type AcceptWithdrawLiquidity struct {
	tokenID     common.Hash
	tokenAmount uint64
	txReqID     common.Hash
	shardID     byte
}
