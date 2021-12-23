package env

import "github.com/incognitochain/incognito-chain/common"

type ValidationEnviroment interface {
	IsPrivacy() bool
	IsConfimed() bool
	TxType() string
	TxAction() int
	ShardID() int
	ShardHeight() uint64
	BeaconHeight() uint64
	ConfirmedTime() int64
	Version() int
	SigPubKey() []byte
	HasCA() bool
	TokenID() common.Hash
	DBData() [][]byte
}
