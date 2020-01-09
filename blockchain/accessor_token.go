package blockchain

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

// ListCustomToken - return all custom token which existed in network
func (blockchain *BlockChain) ListPrivacyCustomTokenV2(shardID byte) (map[common.Hash]*statedb.TokenState, error) {
	tokenStates, err := statedb.ListPrivacyToken(blockchain.GetTransactionStateDB(shardID))
	if err != nil {
		return nil, err
	}
	return tokenStates, nil
}
func (blockchain *BlockChain) GetAllCoinIDV2(shardID byte) ([]common.Hash, error) {
	tokenIDs := []common.Hash{}
	tokenStates, err := blockchain.ListPrivacyCustomTokenV2(shardID)
	if err != nil {
		return nil, err
	}
	for k, _ := range tokenStates {
		tokenIDs = append(tokenIDs, k)
	}
	return tokenIDs, nil
}

// Check Privacy Custom token ID is existed
func (blockchain *BlockChain) PrivacyCustomTokenIDExistedV2(tokenID *common.Hash, shardID byte) bool {
	return statedb.PrivacyTokenIDExisted(blockchain.GetTransactionStateDB(shardID), *tokenID)
}
