package blockchain

import (
	"errors"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

// VerifyCrossShardBlock Verify CrossShard Block
//- Agg Signature
//- MerklePath
func VerifyCrossShardBlock(crossShardBlock *types.CrossShardBlock, blockchain *BlockChain, committees []incognitokey.CommitteePublicKey) error {
	shardBestState := blockchain.GetBestStateShard(crossShardBlock.Header.ShardID)
	tempShardBlock := types.NewShardBlock()
	tempShardBlock.Header.CommitteeFromBlock = crossShardBlock.Header.CommitteeFromBlock
	tempShardBlock.Header.ProposeTime = crossShardBlock.Header.ProposeTime
	_, committeesForSigning, err := shardBestState.getSigningCommittees(tempShardBlock, blockchain)
	if err != nil {
		return err
	}
	if err := blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(crossShardBlock, committeesForSigning); err != nil {
		return NewBlockChainError(SignatureError, err)
	}
	if ok := types.VerifyCrossShardBlockUTXO(crossShardBlock); !ok {
		return NewBlockChainError(HashError, errors.New("Fail to verify Merkle Path Shard"))
	}
	return nil
}
