package mock

import "github.com/incognitochain/incognito-chain/blockchain"

type Fee struct {
}

func (f *Fee) RegisterBlock(block *blockchain.ShardBlock) error {
	return nil
}
