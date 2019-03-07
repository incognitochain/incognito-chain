package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
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
}

func (shardBlock *CrossShardBlock) Hash() *common.Hash {
	hash := shardBlock.Header.Hash()
	return &hash
}

func (shardBlock *ShardToBeaconBlock) Hash() *common.Hash {
	hash := shardBlock.Header.Hash()
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
	beaconBlocks, err := FetchBeaconBlockFromHeight(bc.config.DataBase, bc.BestState.Shard[block.Header.ShardID].BeaconHeight+1, block.Header.BeaconHeight)
	if err != nil {
		Logger.log.Error(err)
		panic(err)
		return nil
	}
	instructions := CreateShardInstructionsFromTransactionAndIns(blk.Body.Transactions, bc, blk.Header.ShardID, blk.Header.ProducerAddress, blk.Header.Height, beaconBlocks)
	if len(instructions) > 0 {
		fmt.Printf("[db] buildActionReq to send to beacon\n")
	}
	block.Instructions = append(block.Instructions, instructions...)
	return &block
}

func (blk *ShardBlock) CreateAllCrossShardBlock(activeShards int) map[byte]*CrossShardBlock {
	allCrossShard := make(map[byte]*CrossShardBlock)
	fmt.Println("########################## 1")
	if activeShards == 1 {
		return allCrossShard
	}
	fmt.Println("########################## 2")
	for i := 0; i < activeShards; i++ {
		if byte(i) != blk.Header.ShardID {
			fmt.Println("########################## 3")
			crossShard, err := blk.CreateCrossShardBlock(byte(i))
			fmt.Printf("Create CrossShardBlock from Shard %+v to Shard %+v: %+v \n", blk.Header.ShardID, i, crossShard)
			if crossShard != nil && err == nil {
				allCrossShard[byte(i)] = crossShard
			}
			fmt.Println("########################## 4")
		}
	}
	fmt.Println("########################## 5")
	return allCrossShard
}

func (block *ShardBlock) CreateCrossShardBlock(shardID byte) (*CrossShardBlock, error) {
	fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@ 1")
	crossShard := &CrossShardBlock{}
	crossOutputCoin := getOutCoinCrossShard(block.Body.Transactions, shardID)
	crossTxTokenData := getTxTokenDataCrossShard(block.Body.Transactions, shardID)
	if len(crossOutputCoin) == 0 && len(crossTxTokenData) == 0 {
		return nil, nil
	}
	fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@ 2")
	merklePathShard, merkleShardRoot := GetMerklePathCrossShard(block.Body.Transactions, shardID)
	fmt.Println("CreateCrossShardBlock/Shard Tx Root", merkleShardRoot)
	if merkleShardRoot != block.Header.ShardTxRoot {
		fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@ 2 ERROR")
		return crossShard, NewBlockChainError(CrossShardBlockError, errors.New("MerkleRootShard mismatch"))
	}
	fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@ 3")
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
	crossShard.ToShardID = shardID
	fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@ 4")
	return crossShard, nil
}
