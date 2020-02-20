package blockchain

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

func (blockchain *BlockChain) ListAllPrivacyCustomTokenAndPRV() (map[common.Hash]*statedb.TokenState, error) {
	tokenStates := make(map[common.Hash]*statedb.TokenState)
	for i := 0; i < blockchain.BestState.Beacon.ActiveShards; i++ {
		shardID := byte(i)
		m, err := blockchain.ListPrivacyCustomTokenAndPRVByShardID(shardID)
		if err != nil {
			return nil, err
		}
		for newK, newV := range m {
			if v, ok := tokenStates[newK]; !ok {
				tokenStates[newK] = newV
				tokenStates[newK].AddTxs([]common.Hash{newV.InitTx()})
			} else {
				if v.PropertyName() == "" && newV.PropertyName() != "" {
					v.SetPropertyName(newV.PropertyName())
				}
				if v.PropertySymbol() == "" && newV.PropertySymbol() != "" {
					v.SetPropertySymbol(newV.PropertySymbol())
				}
				v.AddTxs([]common.Hash{newV.InitTx()})
				v.AddTxs(newV.Txs())
			}
		}
	}
	return tokenStates, nil
}

// ListCustomToken - return all custom token which existed in network
func (blockchain *BlockChain) ListPrivacyCustomTokenAndPRVByShardID(shardID byte) (map[common.Hash]*statedb.TokenState, error) {
	tokenStates, err := statedb.ListPrivacyToken(blockchain.BestState.Shard[shardID].GetCopiedTransactionStateDB())
	if err != nil {
		return nil, err
	}
	return tokenStates, nil
}

func (blockchain *BlockChain) ListPrivacyTokenAndBridgeTokenAndPRVByShardID(shardID byte) ([]common.Hash, error) {
	tokenIDs := []common.Hash{}
	tokenStates, err := blockchain.ListPrivacyCustomTokenAndPRVByShardID(shardID)
	if err != nil {
		return nil, err
	}
	for k, _ := range tokenStates {
		tokenIDs = append(tokenIDs, k)
	}
	brigdeTokenIDs, _, err := blockchain.GetAllBridgeTokens()
	if err != nil {
		return nil, err
	}
	for _, bridgeTokenID := range brigdeTokenIDs {
		if _, found := tokenStates[bridgeTokenID]; !found {
			tokenIDs = append(tokenIDs, bridgeTokenID)
		}
	}
	return tokenIDs, nil
}

// Check Privacy Custom token ID is existed
func (blockchain *BlockChain) PrivacyCustomTokenIDExistedV2(tokenID *common.Hash, shardID byte) bool {
	return statedb.PrivacyTokenIDExisted(blockchain.BestState.Shard[shardID].GetCopiedTransactionStateDB(), *tokenID)
}

func (blockchain *BlockChain) GetAllBridgeTokens() ([]common.Hash, []*rawdbv2.BridgeTokenInfo, error) {
	bridgeTokenIDs := []common.Hash{}
	allBridgeTokens := []*rawdbv2.BridgeTokenInfo{}
	bridgeStateDB := blockchain.BestState.Beacon.GetCopiedFeatureStateDB()
	allBridgeTokensBytes, err := statedb.GetAllBridgeTokens(bridgeStateDB)
	if err != nil {
		return bridgeTokenIDs, allBridgeTokens, err
	}
	err = json.Unmarshal(allBridgeTokensBytes, &allBridgeTokens)
	if err != nil {
		return bridgeTokenIDs, allBridgeTokens, err
	}
	for _, bridgeTokenInfo := range allBridgeTokens {
		bridgeTokenIDs = append(bridgeTokenIDs, *bridgeTokenInfo.TokenID)
	}
	return bridgeTokenIDs, allBridgeTokens, nil
}
