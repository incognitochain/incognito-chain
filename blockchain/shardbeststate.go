package blockchain

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"sync"
	
	"github.com/incognitochain/incognito-chain/cashec"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/transaction"
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
	//fmt.Println("Shard BestState/ Get Public Key Role, Found IN Shard COMMITTEES", found)
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

	if bestStateShard, ok := bestStateShardMap[shardID]; !ok {
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

func (blockchain *BlockChain) ValidateBlockWithPrevShardBestState(block *ShardBlock) error {
	prevBST, err := blockchain.config.DataBase.FetchPrevBestState(false, block.Header.ShardID)
	if err != nil {
		return err
	}
	shardBestState := BestStateShard{}
	if err := json.Unmarshal(prevBST, &shardBestState); err != nil {
		return err
	}

	blkHash := block.Header.Hash()
	producerPk := base58.Base58Check{}.Encode(block.Header.ProducerAddress.Pk, common.ZeroByte)
	err = cashec.ValidateDataB58(producerPk, block.ProducerSig, blkHash.GetBytes())
	if err != nil {
		return NewBlockChainError(ProducerError, errors.New("Producer's sig not match"))
	}
	//verify producer
	producerPosition := (shardBestState.ShardProposerIdx + block.Header.Round) % len(shardBestState.ShardCommittee)
	tempProducer := shardBestState.ShardCommittee[producerPosition]
	if strings.Compare(tempProducer, producerPk) != 0 {
		return NewBlockChainError(ProducerError, errors.New("Producer should be should be :"+tempProducer))
	}
	//if block.Header.Version != VERSION {
	//	return NewBlockChainError(VersionError, errors.New("Version should be :"+strconv.Itoa(VERSION)))
	//}
	// Verify parent hash exist or not
	//prevBlockHash := block.Header.PrevBlockHash
	//parentBlockData, err := blockchain.config.DataBase.FetchBlock(prevBlockHash)
	//if err != nil {
	//	return NewBlockChainError(DBError, err)
	//}
	//parentBlock := ShardBlock{}
	//json.Unmarshal(parentBlockData, &parentBlock)
	//// Verify block height with parent block
	//if parentBlock.Header.Height+1 != block.Header.Height {
	//	return NewBlockChainError(BlockHeightError, errors.New("block height of new block should be :"+strconv.Itoa(int(block.Header.Height+1))))
	//}
	return nil
}

//This only happen if user is a shard committee member.
func (blockchain *BlockChain) RevertShardState(shardID byte) error {
	//Steps:
	// 1. Restore current beststate to previous beststate
	// 2. Set pool shardstate
	// 3. Delete newly inserted block
	// 4. Remove incoming crossShardBlks
	// 5. Delete txs and its related stuff (ex: txview) belong to block

	blockchain.chainLock.Lock()
	defer blockchain.chainLock.Unlock()
	currentBestState := blockchain.BestState.Shard[shardID]
	currentBestStateBlk := currentBestState.BestBlock

	prevBST, err := blockchain.config.DataBase.FetchPrevBestState(false, shardID)
	if err != nil {
		return err
	}
	shardBestState := BestStateShard{}
	if err := json.Unmarshal(prevBST, &shardBestState); err != nil {
		return err
	}

	err = blockchain.DeleteIncomingCrossShard(currentBestStateBlk)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}

	for _, tx := range currentBestState.BestBlock.Body.Transactions {
		if err := blockchain.config.DataBase.DeleteTransactionIndex(*tx.Hash()); err != nil {
			return err
		}
	}

	if err := blockchain.restoreFromTxViewPoint(currentBestStateBlk); err != nil {
		return err
	}

	if err := blockchain.restoreFromCrossTxViewPoint(currentBestStateBlk); err != nil {
		return err
	}

	// DeleteIncomingCrossShard
	blockchain.config.DataBase.DeleteBlock(currentBestStateBlk.Header.Hash(), currentBestStateBlk.Header.Height, shardID)
	blockchain.BestState.Shard[shardID] = &shardBestState
	if err := blockchain.StoreShardBestState(shardID); err != nil {
		return err
	}
	return nil
}

func (blockchain *BlockChain) BackupCurrentShardState(block *ShardBlock) error {

	//Steps:
	// 1. Backup beststate
	// 2.	Backup data that will be modify by new block data

	tempMarshal, err := json.Marshal(blockchain.BestState.Shard[block.Header.ShardID])
	if err != nil {
		return NewBlockChainError(UnmashallJsonBlockError, err)
	}

	if err := blockchain.config.DataBase.StorePrevBestState(tempMarshal, false, block.Header.ShardID); err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}

	if err := blockchain.createBackupFromTxViewPoint(block); err != nil {
		return err
	}

	if err := blockchain.createBackupFromCrossTxViewPoint(block); err != nil {
		return err
	}

	return nil
}

func (blockchain *BlockChain) createBackupFromTxViewPoint(block *ShardBlock) error {
	// Fetch data from block into tx View point
	view := NewTxViewPoint(block.Header.ShardID)
	err := view.fetchTxViewPointFromBlock(blockchain.config.DataBase, block)
	if err != nil {
		return err
	}

	// check privacy custom token
	backupedView := make(map[string]bool)
	for _, privacyCustomTokenSubView := range view.privacyCustomTokenViewPoint {
		if ok := backupedView[privacyCustomTokenSubView.tokenID.String()]; !ok {
			err = blockchain.backupSerialNumbersFromTxViewPoint(*privacyCustomTokenSubView)
			if err != nil {
				return err
			}

			err = blockchain.backupCommitmentsFromTxViewPoint(*privacyCustomTokenSubView)
			if err != nil {
				return err
			}
			backupedView[privacyCustomTokenSubView.tokenID.String()] = true
		}

	}
	err = blockchain.backupSerialNumbersFromTxViewPoint(*view)
	if err != nil {
		return err
	}

	err = blockchain.backupCommitmentsFromTxViewPoint(*view)
	if err != nil {
		return err
	}

	return nil
}

func (blockchain *BlockChain) createBackupFromCrossTxViewPoint(block *ShardBlock) error {
	view := NewTxViewPoint(block.Header.ShardID)
	err := view.fetchCrossTransactionViewPointFromBlock(blockchain.config.DataBase, block)

	for _, privacyCustomTokenSubView := range view.privacyCustomTokenViewPoint {
		err = blockchain.backupCommitmentsFromTxViewPoint(*privacyCustomTokenSubView)
		if err != nil {
			return err
		}
	}
	err = blockchain.backupCommitmentsFromTxViewPoint(*view)
	if err != nil {
		return err
	}

	return nil
}

func (blockchain *BlockChain) backupSerialNumbersFromTxViewPoint(view TxViewPoint) error {
	err := blockchain.config.DataBase.BackupSerialNumbersLen(*view.tokenID, view.shardID)
	if err != nil {
		return err
	}
	return nil
}

func (blockchain *BlockChain) backupCommitmentsFromTxViewPoint(view TxViewPoint) error {

	// commitment
	keys := make([]string, 0, len(view.mapCommitments))
	for k := range view.mapCommitments {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		pubkey := k
		pubkeyBytes, _, err := base58.Base58Check{}.Decode(pubkey)
		if err != nil {
			return err
		}
		lastByte := pubkeyBytes[len(pubkeyBytes)-1]
		pubkeyShardID := common.GetShardIDFromLastByte(lastByte)
		if pubkeyShardID == view.shardID {
			err = blockchain.config.DataBase.BackupCommitmentsOfPubkey(*view.tokenID, view.shardID, pubkeyBytes)
			if err != nil {
				return err
			}
		}
	}

	// outputs
	keys = make([]string, 0, len(view.mapOutputCoins))
	for k := range view.mapOutputCoins {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// for _, k := range keys {
	// 	pubkey := k

	// 	pubkeyBytes, _, err := base58.Base58Check{}.Decode(pubkey)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	lastByte := pubkeyBytes[len(pubkeyBytes)-1]
	// 	pubkeyShardID := common.GetShardIDFromLastByte(lastByte)
	// 	if pubkeyShardID == view.shardID {
	// 		err = blockchain.config.DataBase.BackupOutputCoin(*view.tokenID, pubkeyBytes, pubkeyShardID)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// }
	return nil
}

func (blockchain *BlockChain) restoreFromTxViewPoint(block *ShardBlock) error {
	// Fetch data from block into tx View point
	view := NewTxViewPoint(block.Header.ShardID)
	err := view.fetchTxViewPointFromBlock(blockchain.config.DataBase, block)
	if err != nil {
		return err
	}

	// check normal custom token
	for indexTx, customTokenTx := range view.customTokenTxs {
		switch customTokenTx.TxTokenData.Type {
		case transaction.CustomTokenInit:
			{
				err = blockchain.config.DataBase.DeleteCustomToken(customTokenTx.TxTokenData.PropertyID)
				if err != nil {
					return err
				}
			}
		case transaction.CustomTokenCrossShard:
			{
				err = blockchain.config.DataBase.DeleteCustomToken(customTokenTx.TxTokenData.PropertyID)
				if err != nil {
					return err
				}
			}
		}
		err = blockchain.config.DataBase.DeleteCustomTokenTx(customTokenTx.TxTokenData.PropertyID, indexTx, block.Header.ShardID, block.Header.Height)
		if err != nil {
			return err
		}

	}

	// check privacy custom token
	for indexTx, privacyCustomTokenSubView := range view.privacyCustomTokenViewPoint {
		privacyCustomTokenTx := view.privacyCustomTokenTxs[indexTx]
		switch privacyCustomTokenTx.TxTokenPrivacyData.Type {
		case transaction.CustomTokenInit:
			{
				err = blockchain.config.DataBase.DeletePrivacyCustomToken(privacyCustomTokenTx.TxTokenPrivacyData.PropertyID)
				if err != nil {
					return err
				}
			}
		}
		err = blockchain.config.DataBase.DeletePrivacyCustomTokenTx(privacyCustomTokenTx.TxTokenPrivacyData.PropertyID, indexTx, block.Header.ShardID, block.Header.Height)
		if err != nil {
			return err
		}

		err = blockchain.restoreSerialNumbersFromTxViewPoint(*privacyCustomTokenSubView)
		if err != nil {
			return err
		}

		err = blockchain.restoreCommitmentsFromTxViewPoint(*privacyCustomTokenSubView, block.Header.ShardID)
		if err != nil {
			return err
		}
	}

	err = blockchain.restoreSerialNumbersFromTxViewPoint(*view)
	if err != nil {
		return err
	}

	err = blockchain.restoreCommitmentsFromTxViewPoint(*view, block.Header.ShardID)
	if err != nil {
		return err
	}

	return nil
}

func (blockchain *BlockChain) restoreFromCrossTxViewPoint(block *ShardBlock) error {
	view := NewTxViewPoint(block.Header.ShardID)
	err := view.fetchCrossTransactionViewPointFromBlock(blockchain.config.DataBase, block)

	for _, privacyCustomTokenSubView := range view.privacyCustomTokenViewPoint {
		tokenID := privacyCustomTokenSubView.tokenID
		if err := blockchain.config.DataBase.DeletePrivacyCustomTokenCrossShard(*tokenID); err != nil {
			return err
		}
		err = blockchain.restoreCommitmentsFromTxViewPoint(*privacyCustomTokenSubView, block.Header.ShardID)
		if err != nil {
			return err
		}
	}

	err = blockchain.restoreCommitmentsFromTxViewPoint(*view, block.Header.ShardID)
	if err != nil {
		return err
	}
	return nil
}

func (blockchain *BlockChain) restoreSerialNumbersFromTxViewPoint(view TxViewPoint) error {
	err := blockchain.config.DataBase.RestoreSerialNumber(*view.tokenID, view.shardID, view.listSerialNumbers)
	if err != nil {
		return err
	}
	return nil
}

func (blockchain *BlockChain) restoreCommitmentsFromTxViewPoint(view TxViewPoint, shardID byte) error {

	// commitment
	keys := make([]string, 0, len(view.mapCommitments))
	for k := range view.mapCommitments {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		pubkey := k
		item1 := view.mapCommitments[k]
		pubkeyBytes, _, err := base58.Base58Check{}.Decode(pubkey)
		if err != nil {
			return err
		}
		lastByte := pubkeyBytes[len(pubkeyBytes)-1]
		pubkeyShardID := common.GetShardIDFromLastByte(lastByte)
		if pubkeyShardID == view.shardID {
			err = blockchain.config.DataBase.RestoreCommitmentsOfPubkey(*view.tokenID, view.shardID, pubkeyBytes, item1)
			if err != nil {
				return err
			}
		}
	}

	// outputs
	for _, k := range keys {
		publicKey := k
		publicKeyBytes, _, err := base58.Base58Check{}.Decode(publicKey)
		if err != nil {
			return err
		}
		lastByte := publicKeyBytes[len(publicKeyBytes)-1]
		publicKeyShardID := common.GetShardIDFromLastByte(lastByte)
		if publicKeyShardID == shardID {
			outputCoinArray := view.mapOutputCoins[k]
			outputCoinBytesArray := make([][]byte, 0)
			for _, outputCoin := range outputCoinArray {
				outputCoinBytesArray = append(outputCoinBytesArray, outputCoin.Bytes())
			}
			err = blockchain.config.DataBase.DeleteOutputCoin(*view.tokenID, publicKeyBytes, outputCoinBytesArray, publicKeyShardID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
