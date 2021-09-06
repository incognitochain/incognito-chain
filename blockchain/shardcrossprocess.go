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
	if err := blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(crossShardBlock, committees); err != nil {
		return NewBlockChainError(SignatureError, err)
	}
	if ok := types.VerifyCrossShardBlockUTXO(crossShardBlock); !ok {
		return NewBlockChainError(HashError, errors.New("Fail to verify Merkle Path Shard"))
	}
	return nil
}
