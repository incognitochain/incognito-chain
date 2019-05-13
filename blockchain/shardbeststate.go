package blockchain

import (
	"encoding/binary"
	"fmt"
	"sort"
	"sync"

	"github.com/constant-money/constant-chain/common"
)

// BestState houses information about the current best block and other info
// related to the state of the main chain as it exists from the point of view of
// the current best block.
//
// The BestSnapshot method can be used to obtain access to this information
// in a concurrent safe manner and the data will not be changed out from under
// the caller when chain state changes occur as the function name implies.
// However, the returned snapshot must be treated as immutable since it is
// shared by all callers.

type BestStateShard struct {
	BestBlockHash          common.Hash       `json:"BestBlockHash"` // hash of block.
	BestBlock              *ShardBlock       `json:"BestBlock"`     // block data
	BestBeaconHash         common.Hash       `json:"BestBeaconHash"`
	BeaconHeight           uint64            `json:"BeaconHeight"`
	ShardID                byte              `json:"ShardID"`
	Epoch                  uint64            `json:"Epoch"`
	ShardHeight            uint64            `json:"ShardHeight"`
	ShardCommitteeSize     int               `json:"ShardCommitteeSize"`
	ShardProposerIdx       int               `json:"ShardProposerIdx"`
	ShardCommittee         []string          `json:"ShardCommittee"`
	ShardPendingValidator  []string          `json:"ShardPendingValidator"`
	BestCrossShard         map[byte]uint64   `json:"BestCrossShard"` // Best cross shard block by heigh
	StakingTx              map[string]string `json:"StakingTx"`
	NumTxns                uint64            `json:"NumTxns"`                // The number of txns in the block.
	TotalTxns              uint64            `json:"TotalTxns"`              // The total number of txns in the chain.
	TotalTxnsExcludeSalary uint64            `json:"TotalTxnsExcludeSalary"` // for testing and benchmark
	ActiveShards           int               `json:"ActiveShards"`

	MetricBlockHeight uint64
	lock              sync.RWMutex
}

// Get role of a public key base on best state shard
func (bestStateShard *BestStateShard) GetBytes() []byte {
	res := []byte{}
	res = append(res, bestStateShard.BestBlockHash.GetBytes()...)
	res = append(res, bestStateShard.BestBlock.Hash().GetBytes()...)
	res = append(res, bestStateShard.BestBeaconHash.GetBytes()...)
	beaconHeightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(beaconHeightBytes, bestStateShard.BeaconHeight)
	res = append(res, beaconHeightBytes...)
	res = append(res, bestStateShard.ShardID)
	epochBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(epochBytes, bestStateShard.Epoch)
	res = append(res, epochBytes...)
	shardHeightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(shardHeightBytes, bestStateShard.ShardHeight)
	res = append(res, shardHeightBytes...)
	shardCommitteeSizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(shardCommitteeSizeBytes, uint32(bestStateShard.ShardCommitteeSize))
	res = append(res, shardCommitteeSizeBytes...)
	proposerIdxBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(proposerIdxBytes, uint32(bestStateShard.ShardProposerIdx))
	res = append(res, proposerIdxBytes...)
	for _, value := range bestStateShard.ShardCommittee {
		res = append(res, []byte(value)...)
	}
	for _, value := range bestStateShard.ShardPendingValidator {
		res = append(res, []byte(value)...)
	}

	keys := []int{}
	for k := range bestStateShard.BestCrossShard {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		value := bestStateShard.BestCrossShard[byte(shardID)]
		valueBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(valueBytes, value)
		res = append(res, valueBytes...)
	}

	keystr := []string{}
	for _, k := range bestStateShard.StakingTx {
		keystr = append(keystr, k)
	}
	sort.Strings(keystr)
	for key, value := range bestStateShard.StakingTx {
		res = append(res, []byte(key)...)
		res = append(res, []byte(value)...)
	}

	numTxnsBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(numTxnsBytes, bestStateShard.NumTxns)
	res = append(res, numTxnsBytes...)
	totalTxnsBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(totalTxnsBytes, bestStateShard.TotalTxns)
	res = append(res, totalTxnsBytes...)
	activeShardsBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(activeShardsBytes, uint32(bestStateShard.ActiveShards))
	res = append(res, activeShardsBytes...)

	return res
}
func (bestStateShard *BestStateShard) Hash() common.Hash {
	bestStateShard.lock.RLock()
	defer bestStateShard.lock.RUnlock()
	return common.HashH(bestStateShard.GetBytes())
}
func (bestStateShard *BestStateShard) GetPubkeyRole(pubkey string, round int) string {
	// fmt.Println("Shard BestState/ BEST STATE", bestStateShard)
	found := common.IndexOfStr(pubkey, bestStateShard.ShardCommittee)
	fmt.Println("Shard BestState/ Get Public Key Role, Found IN Shard COMMITTEES", found)
	if found > -1 {
		tmpID := (bestStateShard.ShardProposerIdx + round) % len(bestStateShard.ShardCommittee)
		if found == tmpID {
			// fmt.Printf("Shard BestState/ Get Public Key Role, ROLE %+v , Shard %+v \n", common.PROPOSER_ROLE, bestStateShard.ShardID)
			return common.PROPOSER_ROLE
		} else {
			// fmt.Printf("Shard BestState/ Get Public Key Role, ROLE %+v , Shard %+v \n", common.VALIDATOR_ROLE, bestStateShard.ShardID)
			return common.VALIDATOR_ROLE
		}

	}

	found = common.IndexOfStr(pubkey, bestStateShard.ShardPendingValidator)
	if found > -1 {
		// fmt.Printf("Shard BestState/ Get Public Key Role, ROLE %+v , Shard %+v \n", common.PENDING_ROLE, bestStateShard.ShardID)
		return common.PENDING_ROLE
	}

	return common.EmptyString
}

var bestStateShardMap = make(map[byte]*BestStateShard)

func GetBestStateShard(shardID byte) *BestStateShard {

	if bestStateShard, ok := bestStateShardMap[shardID]; ok != true {
		bestStateShardMap[shardID] = &BestStateShard{}
		bestStateShardMap[shardID].ShardID = shardID
		return bestStateShardMap[shardID]
	} else {
		return bestStateShard
	}
}

func SetBestStateShard(shardID byte, beststateShard *BestStateShard) {
	bestStateShardMap[shardID] = beststateShard
}

func InitBestStateShard(shardID byte, netparam *Params) *BestStateShard {
	bestStateShard := GetBestStateShard(shardID)

	bestStateShard.BestBlockHash.SetBytes(make([]byte, 32))
	bestStateShard.BestBeaconHash.SetBytes(make([]byte, 32))
	bestStateShard.BestBlock = nil
	bestStateShard.ShardCommittee = []string{}
	bestStateShard.ShardCommitteeSize = netparam.ShardCommitteeSize
	bestStateShard.ShardPendingValidator = []string{}
	bestStateShard.ActiveShards = netparam.ActiveShards
	bestStateShard.BestCrossShard = make(map[byte]uint64)
	bestStateShard.StakingTx = make(map[string]string)
	bestStateShard.ShardHeight = 1
	bestStateShard.BeaconHeight = 1

	return bestStateShard
}

//This only happen if user is a shard committee member.
func (blockchain *BlockChain) RevertShardState(shardID byte) {
	//Steps:
	// 1. Load backup beststate
	// 2. Set pool shardstate

}

func (blockchain *BlockChain) SaveCurrentShardState(block *ShardBlock) error {

	//Steps:
	// 1. Backup beststate
	// 2.

	// tempMarshal, err := json.Marshal(blockchain.BestState.Shard[block.Header.ShardID])
	// if err != nil {
	// 	return NewBlockChainError(UnmashallJsonBlockError, err)
	// }

	// go func() {
	// 	errCh <- blockchain.ProcessLoanForBlock(block)
	// }()

	// go func() {
	// 	errCh <- blockchain.processTradeBondTx(block)
	// }()

	// for _, tx := range block.Body.Transactions {
	// 	var err error

	// 	switch tx.GetMetadataType() {
	// 	case metadata.BuyFromGOVRequestMeta:
	// 		err = blockchain.processBuyBondTx(tx)

	// 	case metadata.BuyBackRequestMeta:
	// 		err = blockchain.processSellBondTx(tx)
	// 	case metadata.LoanUnlockMeta:
	// 		{
	// 			tx := tx.(*transaction.Tx)
	// 			meta := tx.GetMetadata().(*metadata.LoanUnlock)
	// 			err = blockchain.config.DataBase.StoreLoanWithdrawed(meta.LoanID)
	// 		}
	// 	}

	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// go func() {
	// 	// Process stability stand-alone instructions
	// 	errCh <- blockchain.ProcessStandAloneInstructions(block)
	// }()
	// for _, inst := range block.Body.Instructions {
	// 	if len(inst) < 2 {
	// 		continue
	// 	}
	// 	var err error
	// 	switch inst[0] {
	// 	case strconv.Itoa(component.ConfirmBuySellRequestMeta):
	// 		err = blockchain.processConfirmBuySellInst(inst)
	// 	case strconv.Itoa(component.ConfirmBuyBackRequestMeta):
	// 		err = blockchain.processConfirmBuyBackInst(inst)
	// 	}
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// Store metadata instruction to local state
	// for _, beaconBlock := range beaconBlocks {
	// 	instructions := beaconBlock.Body.Instructions
	// 	for _, inst := range instructions {
	// 		err := blockchain.StoreMetadataInstructions(inst, shardID)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// }

	// if len(inst) < 2 {
	// 	return nil // Not error, just not stability instruction
	// }
	// switch inst[0] {
	// case strconv.Itoa(metadata.IssuingRequestMeta):
	// 	return bc.storeIssuingResponseInstruction(inst, shardID)
	// case strconv.Itoa(metadata.ContractingRequestMeta):
	// 	return bc.storeContractingResponseInstruction(inst, shardID)
	// }

	// //========Store new  Shard block and new shard bestState
	// err = blockchain.ProcessStoreShardBlock(block)
	// if err != nil {
	// 	return err
	// }
	// Logger.log.Infof("SHARD %+v | Finish Insert new block %d, with hash %+v", block.Header.ShardID, block.Header.Height, *block.Hash())
	// return nil

	return nil
}
