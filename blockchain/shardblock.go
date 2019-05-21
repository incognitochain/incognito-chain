package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
)

type ShardBlock struct {
	AggregatedSig string  `json:"AggregatedSig"`
	R             string  `json:"R"`
	ValidatorsIdx [][]int `json:"ValidatorsIdx"` //[0]: R | [1]:AggregatedSig
	ProducerSig   string  `json:"ProducerSig"`
	Body          ShardBody
	Header        ShardHeader
}

type ShardToBeaconBlock struct {
	AggregatedSig string  `json:"AggregatedSig"`
	R             string  `json:"R"`
	ValidatorsIdx [][]int `json:"ValidatorsIdx"` //[0]: R | [1]:AggregatedSig
	ProducerSig   string  `json:"ProducerSig"`

	Instructions [][]string
	Header       ShardHeader
}

type CrossShardBlock struct {
	AggregatedSig   string  `json:"AggregatedSig"`
	R               string  `json:"R"`
	ValidatorsIdx   [][]int `json:"ValidatorsIdx"` //[0]: R | [1]:AggregatedSig
	ProducerSig     string  `json:"ProducerSig"`
	Header          ShardHeader
	ToShardID       byte
	MerklePathShard []common.Hash
	// Cross Shard data for constant
	CrossOutputCoin []privacy.OutputCoin
	// Cross Shard Data for Custom Token Tx
	CrossTxTokenData []transaction.TxTokenData
	//TODO: add to hash
	CrossTxTokenPrivacyData []ContentCrossTokenPrivacyData
}

func (crossShardBlock *CrossShardBlock) Hash() *common.Hash {
	hash := crossShardBlock.Header.Hash()
	return &hash
}

func (shardToBeaconBlock *ShardToBeaconBlock) Hash() *common.Hash {
	hash := shardToBeaconBlock.Header.Hash()
	return &hash
}

func (shardBlock *ShardBlock) Hash() *common.Hash {
	hash := shardBlock.Header.Hash()
	return &hash
}

func (shardBlock *ShardBlock) UnmarshalJSON(data []byte) error {
	tempBlk := &struct {
		AggregatedSig string
		R             string `json:"R"`
		ValidatorsIdx [][]int
		ProducerSig   string
		Header        ShardHeader
		Body          *json.RawMessage
	}{}
	err := json.Unmarshal(data, &tempBlk)
	if err != nil {
		return NewBlockChainError(UnmashallJsonBlockError, err)
	}
	shardBlock.AggregatedSig = tempBlk.AggregatedSig
	shardBlock.R = tempBlk.R
	shardBlock.ValidatorsIdx = tempBlk.ValidatorsIdx
	shardBlock.ProducerSig = tempBlk.ProducerSig

	blkBody := ShardBody{}
	err = blkBody.UnmarshalJSON(*tempBlk.Body)
	if err != nil {
		return NewBlockChainError(UnmashallJsonBlockError, err)
	}
	shardBlock.Header = tempBlk.Header

	shardBlock.Body = blkBody
	return nil
}

// /*
// AddTransaction adds a new transaction into block
// */
// // #1 - tx
func (shardBlock *ShardBlock) AddTransaction(tx metadata.Transaction) error {
	if shardBlock.Body.Transactions == nil {
		return NewBlockChainError(UnExpectedError, errors.New("not init tx arrays"))
	}
	shardBlock.Body.Transactions = append(shardBlock.Body.Transactions, tx)
	return nil
}

func (blk *ShardBlock) CreateShardToBeaconBlock(bc *BlockChain) *ShardToBeaconBlock {
	block := ShardToBeaconBlock{}
	block.AggregatedSig = blk.AggregatedSig

	block.ValidatorsIdx = make([][]int, 2)                                           //multi-node
	block.ValidatorsIdx[0] = append(block.ValidatorsIdx[0], blk.ValidatorsIdx[0]...) //multi-node
	block.ValidatorsIdx[1] = append(block.ValidatorsIdx[1], blk.ValidatorsIdx[1]...) //multi-node

	block.R = blk.R
	block.ProducerSig = blk.ProducerSig
	block.Header = blk.Header
	block.Instructions = blk.Body.Instructions
	previousShardBlockByte, err := bc.config.DataBase.FetchBlock(&blk.Header.PrevBlockHash)
	if err != nil {
		Logger.log.Error(err)
		return nil
	}
	previousShardBlock := ShardBlock{}
	err = json.Unmarshal(previousShardBlockByte, &previousShardBlock)
	if err != nil {
		Logger.log.Error(err)
		return nil
	}
	beaconBlocks, err := FetchBeaconBlockFromHeight(bc.config.DataBase, previousShardBlock.Header.BeaconHeight+1, block.Header.BeaconHeight)
	// fmt.Println("[ndh] - newshardtobeacon", previousShardBlock.Header.BeaconHeight, block.Header.BeaconHeight)
	if err != nil {
		Logger.log.Error(err)
		fmt.Println("[ndh] - error in create shard to beacon", err)
		return nil
	}
	instructions, err := CreateShardInstructionsFromTransactionAndIns(blk.Body.Transactions, bc, blk.Header.ShardID, &blk.Header.ProducerAddress, blk.Header.Height, beaconBlocks, blk.Header.BeaconHeight)
	if err != nil {
		Logger.log.Error(err)
		fmt.Println("[ndh] - error in create shard to beacon", err)
		return nil
	}
	for _, inst := range instructions {
		if len(inst) != 0 {
			if inst[0] != "37" {
				fmt.Println("[ndh] - instruction to beacon: ", inst)
			}
		}
	}
	block.Instructions = append(block.Instructions, instructions...)
	return &block
}

func (blk *ShardBlock) CreateAllCrossShardBlock(activeShards int) map[byte]*CrossShardBlock {
	allCrossShard := make(map[byte]*CrossShardBlock)
	if activeShards == 1 {
		return allCrossShard
	}
	for i := 0; i < activeShards; i++ {
		shardID := common.GetShardIDFromLastByte(byte(i))
		if shardID != blk.Header.ShardID {
			crossShard, err := blk.CreateCrossShardBlock(shardID)
			//fmt.Printf("Create CrossShardBlock from Shard %+v to Shard %+v: %+v \n", blk.Header.ShardID, shardID, crossShard)
			if crossShard != nil && err == nil {
				allCrossShard[byte(i)] = crossShard
			}
		}
	}
	return allCrossShard
}

func (block *ShardBlock) CreateCrossShardBlock(shardID byte) (*CrossShardBlock, error) {
	crossShard := &CrossShardBlock{}
	crossOutputCoin, crossTxTokenData, crossCustomTokenPrivacyData := getCrossShardData(block.Body.Transactions, shardID)
	// Return nothing if nothing to cross
	if len(crossOutputCoin) == 0 && len(crossTxTokenData) == 0 && len(crossCustomTokenPrivacyData) == 0 {
		//fmt.Println("CreateCrossShardBlock no crossshard", block.Header.Height)
		return nil, NewBlockChainError(CrossShardBlockError, errors.New("No cross outputcoin"))
	}
	merklePathShard, merkleShardRoot := GetMerklePathCrossShard2(block.Body.Transactions, shardID)
	if merkleShardRoot != block.Header.ShardTxRoot {
		return crossShard, NewBlockChainError(CrossShardBlockError, errors.New("ShardTxRoot mismatch"))
	}
	//Copy signature and header
	crossShard.AggregatedSig = block.AggregatedSig

	crossShard.ValidatorsIdx = make([][]int, 2)                                                  //multi-node
	crossShard.ValidatorsIdx[0] = append(crossShard.ValidatorsIdx[0], block.ValidatorsIdx[0]...) //multi-node
	crossShard.ValidatorsIdx[1] = append(crossShard.ValidatorsIdx[1], block.ValidatorsIdx[1]...) //multi-node

	crossShard.R = block.R
	crossShard.ProducerSig = block.ProducerSig
	crossShard.Header = block.Header
	crossShard.MerklePathShard = merklePathShard
	crossShard.CrossOutputCoin = crossOutputCoin
	crossShard.CrossTxTokenData = crossTxTokenData
	crossShard.CrossTxTokenPrivacyData = crossCustomTokenPrivacyData
	crossShard.ToShardID = shardID
	return crossShard, nil
}
