package blockchain

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
)

func (blockchain *BlockChain) verifyMinerCreatedTxBeforeGettingInBlock(
	insts [][]string,
	txs []metadata.Transaction,
	shardID byte,
) ([]metadata.Transaction, error) {

	mintData := new(metadata.MintData)
	mintData.Txs = txs
	mintData.TxsUsed =  make([]int, len(txs))
	mintData.Insts = insts
	mintData.InstsUsed = make([]int, len(insts))

	accumulatedValues := &metadata.AccumulatedValues{
		UniqETHTxsUsed:   [][]byte{},
		DBridgeTokenPair: map[string][]byte{},
		CBridgeTokens:    []*common.Hash{},
	}

	invalidTxs := []metadata.Transaction{}

	mintData.ReturnStaking = make(map[string]bool)
	mintData.WithdrawReward = make(map[string]bool)

	for _, tx := range txs {
		fmt.Println("Mint Data", mintData.WithdrawReward)
		shardViewRetriever := blockchain.GetBestStateShard(shardID)
		beaconViewRetriever := blockchain.GetBeaconBestState()
		ok, err := tx.VerifyMinerCreatedTxBeforeGettingInBlock(mintData, shardID, blockchain, accumulatedValues, shardViewRetriever, beaconViewRetriever)
		if err != nil {
			return nil, err
		}
		if !ok {
			invalidTxs = append(invalidTxs, tx)
		}
	}
	if len(invalidTxs) > 0 {
		return invalidTxs, nil
	}
	return invalidTxs, nil
}
