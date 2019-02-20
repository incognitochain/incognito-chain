package blockchain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
)

func (self *BlockChain) StoreMetadata(tx metadata.Transaction) error {
	switch tx.GetMetadataType() {
	}
	return nil
}

func (self *BlockChain) VerifyPreSignShardBlock(block *ShardBlock, shardID byte) error {
	self.chainLock.Lock()
	defer self.chainLock.Unlock()
	//========Verify block only
	Logger.log.Infof("Verify block for signing process %d, with hash %+v", block.Header.Height, *block.Hash())
	if err := self.VerifyPreProcessingShardBlock(block, shardID); err != nil {
		return err
	}
	//========Verify block with previous best state
	// Get Beststate of previous block == previous best state
	// Clone best state value into new variable
	shardBestState := BestStateShard{}
	// check with current final best state
	// New block must be compatible with current best state
	if strings.Compare(self.BestState.Shard[shardID].BestShardBlockHash.String(), block.Header.PrevBlockHash.String()) == 0 {
		tempMarshal, err := json.Marshal(self.BestState.Shard[shardID])
		if err != nil {
			return NewBlockChainError(UnmashallJsonBlockError, err)
		}
		json.Unmarshal(tempMarshal, &shardBestState)
	}
	// if no match best state found then block is unknown
	if reflect.DeepEqual(shardBestState, BestStateShard{}) {
		return NewBlockChainError(ShardError, errors.New("Shard Block does not match with any Shard State in cache or in Database"))
	}
	// Verify block with previous best state
	// not verify agg signature in this function
	prevBeaconHeight := shardBestState.BeaconHeight
	beaconBlocks, err := FetchBeaconBlockFromHeight(self.config.DataBase, prevBeaconHeight+1, block.Header.BeaconHeight)
	if err != nil {
		return err
	}
	if err := shardBestState.VerifyBestStateWithShardBlock(block, false, shardID); err != nil {
		return err
	}
	//========Update best state with new block
	if err := shardBestState.Update(block, beaconBlocks); err != nil {
		return err
	}
	//========Post verififcation: verify new beaconstate with corresponding block
	if err := shardBestState.VerifyPostProcessingShardBlock(block, shardID); err != nil {
		return err
	}
	Logger.log.Infof("Block %d, with hash %+v is VALID for signing", block.Header.Height, *block.Hash())
	return nil
}

func (self *BlockChain) ProcessStoreShardBlock(block *ShardBlock) error {
	blockHash := block.Hash().String()
	Logger.log.Debugf("Process store block %+v", blockHash)

	if err := self.StoreShardBlock(block); err != nil {
		return err
	}

	if err := self.StoreShardBlockIndex(block); err != nil {
		return err
	}

	if err := self.StoreShardBestState(block.Header.ShardID); err != nil {
		return err
	}

	// Process transaction db
	if len(block.Body.Transactions) < 1 {
		Logger.log.Infof("No transaction in this block")
	} else {
		Logger.log.Infof("Number of transaction in this block %d", len(block.Body.Transactions))
	}

	// TODO: Check: store output coin?
	if err := self.CreateAndSaveTxViewPointFromBlock(block); err != nil {
		return err
	}

	for index, tx := range block.Body.Transactions {
		if tx.GetType() == common.TxCustomTokenPrivacyType {
			_ = 1
			//TODO: do what???
		}

		if err := self.StoreTransactionIndex(tx.Hash(), block.Hash(), index); err != nil {
			Logger.log.Error("ERROR", err, "Transaction in block with hash", blockHash, "and index", index, ":", tx)
			return NewBlockChainError(UnExpectedError, err)
		}
		Logger.log.Debugf("Transaction in block with hash", blockHash, "and index", index)

		// Store metadata if needed
		if tx.GetMetadata() != nil {
			if err := self.StoreMetadata(tx); err != nil {
				return err
			}
		}
	}
	err := self.StoreIncomingCrossShard(block)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}
	// err = self.StoreOutgoingCrossShard(block)
	// if err != nil {
	// 	return NewBlockChainError(UnExpectedError, err)
	// }
	//TODO: store most recent proccess cross shard block
	return nil
}

func (self *BlockChain) InsertShardBlock(block *ShardBlock) error {
	self.chainLock.Lock()
	defer self.chainLock.Unlock()
	shardID := block.Header.ShardID
	Logger.log.Infof("SHARD %+v | Begin Insert new block height %+v at hash %+v", block.Header.ShardID, block.Header.Height, block.Hash())
	Logger.log.Infof("SHARD %+v | Verify Pre Processing  Block %+v \n", block.Header.ShardID, *block.Hash())
	if err := self.VerifyPreProcessingShardBlock(block, shardID); err != nil {
		return err
	}
	//========Verify block with previous best state
	// check with current final best state
	// block can only be insert if it match the current best state
	if strings.Compare(self.BestState.Shard[shardID].BestShardBlockHash.String(), block.Header.PrevBlockHash.String()) != 0 {
		return NewBlockChainError(BeaconError, errors.New("Beacon Block does not match with any Beacon State in cache or in Database"))
	}
	// fmt.Printf("BeaconBest state %+v \n", self.BestState.Beacon)
	Logger.log.Infof("SHARD %+v | Verify BestState with Block %+v \n", block.Header.ShardID, *block.Hash())
	// Verify block with previous best state
	if err := self.BestState.Shard[shardID].VerifyBestStateWithShardBlock(block, true, shardID); err != nil {
		return err
	}

	Logger.log.Infof("SHARD %+v | Update BestState with Block %+v \n", block.Header.ShardID, *block.Hash())
	//========Update best state with new block
	prevBeaconHeight := self.BestState.Shard[shardID].BeaconHeight
	beaconBlocks, err := FetchBeaconBlockFromHeight(self.config.DataBase, prevBeaconHeight+1, block.Header.BeaconHeight)
	if err != nil {
		return err
	}
	if err := self.BestState.Shard[shardID].Update(block, beaconBlocks); err != nil {
		return err
	}

	Logger.log.Infof("SHARD %+v | Verify Post Processing Block %+v \n", block.Header.ShardID, *block.Hash())
	//========Post verififcation: verify new beaconstate with corresponding block
	if err := self.BestState.Shard[shardID].VerifyPostProcessingShardBlock(block, shardID); err != nil {
		return err
	}
	//========Store new Beaconblock and new Beacon bestState
	self.ProcessStoreShardBlock(block)

	// Process stability tx
	err = self.ProcessLoanForBlock(block)
	if err != nil {
		return err
	}
	err = self.ProcessDividendForBlock(block)
	if err != nil {
		return err
	}
	err = self.ProcessVotingForBlock(block)
	if err != nil {
		return err
	}

	// Store metadata instruction to local state
	for _, beaconBlock := range beaconBlocks {
		instructions := beaconBlock.Body.Instructions
		for _, inst := range instructions {
			err := self.StoreMetadataInstructions(inst, shardID)
			if err != nil {
				return err
			}
		}
	}

	//TODO: Remove cross shard block in pool
	Logger.log.Infof("SHARD %+v | Finish Insert new block %d, with hash %+v", block.Header.ShardID, block.Header.Height, *block.Hash())
	return nil
}

/* Verify Pre-prosessing data
This function DOES NOT verify new block with best state
DO NOT USE THIS with GENESIS BLOCK
- Producer
- ShardID: of received block same shardID with input
- Version
- Parent hash
- Height = parent hash + 1
- Epoch = blockHeight % Epoch ? Parent Epoch + 1
- Timestamp can not excess some limit
- TxRoot
- ShardTxRoot
- CrossOutputCoinRoot
- ActionsRoot
- BeaconHeight
- BeaconHash
- Swap instruction
*/
func (self *BlockChain) VerifyPreProcessingShardBlock(block *ShardBlock, shardID byte) error {
	Logger.log.Debugf("SHARD %+v | Begin VerifyPreProcessingShardBlock Block with height %+v at hash %+v", block.Header.ShardID, block.Header.Height, block.Hash())
	if block.Header.ShardID != shardID {
		return NewBlockChainError(ShardIDError, errors.New("Shard should be :"+strconv.Itoa(int(shardID))))
	}
	if block.Header.Version != VERSION {
		return NewBlockChainError(VersionError, errors.New("Version should be :"+strconv.Itoa(VERSION)))
	}
	// Verify parent hash exist or not
	prevBlockHash := block.Header.PrevBlockHash
	parentBlockData, err := self.config.DataBase.FetchBlock(&prevBlockHash)
	if err != nil {
		return NewBlockChainError(DBError, err)
	}
	parentBlock := ShardBlock{}
	json.Unmarshal(parentBlockData, &parentBlock)
	// Verify block height with parent block
	if parentBlock.Header.Height+1 != block.Header.Height {
		return NewBlockChainError(BlockHeightError, errors.New("Block height of new block should be :"+strconv.Itoa(int(block.Header.Height+1))))
	}
	// Verify epoch with parent block
	// if block.Header.Height%EPOCH == 0 && parentBlock.Header.Epoch != block.Header.Epoch-1 {
	// 	return NewBlockChainError(EpochError, errors.New("Block height and Epoch is not compatiable"))
	// }
	// Verify timestamp with parent block
	if block.Header.Timestamp <= parentBlock.Header.Timestamp {
		return NewBlockChainError(TimestampError, errors.New("Timestamp of new block can't equal to parent block"))
	}
	// Verify transaction root
	txMerkle := Merkle{}.BuildMerkleTreeStore(block.Body.Transactions)
	txRoot := txMerkle[len(txMerkle)-1]

	if bytes.Compare(block.Header.TxRoot.GetBytes(), txRoot.GetBytes()) != 0 {
		fmt.Println()
		test, _ := json.Marshal(block.Body.Transactions[0])
		fmt.Println(len(block.Body.Transactions), string(test))
		fmt.Println()
		return NewBlockChainError(HashError, errors.New("Can't Verify Transaction Root"))
	}
	// Verify ShardTx Root
	shardTxRoot := block.Body.CalcMerkleRootShard()

	if bytes.Compare(block.Header.ShardTxRoot.GetBytes(), shardTxRoot.GetBytes()) != 0 {
		return NewBlockChainError(HashError, errors.New("Can't Verify CrossShardTransaction Root"))
	}
	// Verify Crossoutput coin
	if !VerifyMerkleCrossOutputCoin(block.Body.CrossOutputCoin, block.Header.CrossOutputCoinRoot) {
		return NewBlockChainError(HashError, errors.New("Can't Verify CrossOutputCoin Root"))
	}
	//Verify transaction
	for _, tx := range block.Body.Transactions {
		if !tx.ValidateTxByItself(tx.IsPrivacy(), self.config.DataBase, self, shardID) {
			return NewBlockChainError(TransactionError, errors.New("Can't Validate transaction"))
		}
	}
	// Verify Action
	actions := CreateShardActionFromTransaction(block.Body.Transactions, self, shardID)
	action := []string{}
	for _, value := range actions {
		action = append(action, value...)
	}
	for _, value := range block.Body.Instructions {
		action = append(action, value...)
	}
	isOk := VerifyHashFromStringArray(action, block.Header.ActionsRoot)
	if !isOk {
		return NewBlockChainError(HashError, errors.New("Error verify action root"))
	}
	//Get beacon hash by height in db
	//If hash not found then fail to verify
	beaconHash, err := self.config.DataBase.GetBeaconBlockHashByIndex(block.Header.BeaconHeight)
	if err != nil {
		return err
	}
	//Hash in db must be equal to hash in shard block
	newHash, err := common.Hash{}.NewHash(block.Header.BeaconHash.GetBytes())
	if err != nil {
		return NewBlockChainError(HashError, err)
	}
	if newHash.IsEqual(beaconHash) == false {
		return NewBlockChainError(BeaconError, errors.New("Beacon block height and beacon block hash are not compatible in Database"))
	}
	// Swap instruction
	for _, l := range block.Body.Instructions {
		if l[0] == "swap" {
			if l[3] != "shard" || l[4] != strconv.Itoa(int(shardID)) {
				return NewBlockChainError(InstructionError, errors.New("Swap instruction is invalid"))
			}
		}
	}
	//TODO: UNCOMMENT To verify Cross Shard Block
	// // Get cross shard block from pool
	// crossShardMap := make(map[byte][]CrossShardBlock)
	// bestShardHeight := self.BestState.Beacon.BestShardHeight
	// allCrossShardBlock := self.config.CrossShardPool.GetBlock(bestShardHeight)
	// oneShardCrossShardBlocks := allCrossShardBlock[shardID]
	// currentBestCrossShard := self.BestState.Shard[shardID].BestCrossShard
	// for _, blk := range oneShardCrossShardBlocks {
	// 	crossShardMap[blk.Header.ShardID] = append(crossShardMap[blk.Header.ShardID], blk)
	// }
	// for crossShardID, crossShardBlocks := range crossShardMap {
	// 	sort.SliceStable(crossShardBlocks[:], func(i, j int) bool {
	// 		return crossShardBlocks[i].Header.Height < crossShardBlocks[j].Header.Height
	// 	})
	// 	// compare cross shard block with received cross output coin
	// 	crossOutputCoins := block.Body.CrossOutputCoin[crossShardID]
	// 	for _, crossOutputCoin := range crossOutputCoins {
	// 		found := false
	// 		for _, crossShardBlock := range crossShardBlocks {
	// 			if crossOutputCoin.BlockHeight == crossShardBlock.Header.Height {
	// 				found = true
	// 				break
	// 			}
	// 		}
	// 		if !found {
	// 			return NewBlockChainError(ShardStateError, errors.New("No CrossOutputCoin can't be found from any CrossShardBlock in pool"))
	// 		}
	// 	}
	// 	currentBestCrossShardForThisBlock := currentBestCrossShard[crossShardID]
	// 	for _, blk := range crossShardBlocks {
	// 		temp, err := self.config.DataBase.FetchBeaconCommitteeByHeight(blk.Header.BeaconHeight)
	// 		if err != nil {
	// 			return NewBlockChainError(UnmashallJsonBlockError, err)
	// 		}
	// 		shardCommittee := make(map[byte][]string)
	// 		json.Unmarshal(temp, &shardCommittee)
	// 		err = blk.VerifyCrossShardBlock(shardCommittee[shardID])
	// 		if err != nil {
	// 			return NewBlockChainError(SignatureError, err)
	// 		}
	// 		// Verify with bytemap in beacon
	// 		passed := false
	// 		for i := self.BestState.Shard[shardID].BeaconHeight + 1; i <= block.Header.BeaconHeight; i++ {
	// 			for shardToBeaconID, shardStates := range self.BestState.Beacon.AllShardState {
	// 				if crossShardID == shardToBeaconID {
	// 					// compare crossoutputcoin with bytemap in beacon
	// 					for i := int(currentBestCrossShardForThisBlock); i < len(shardStates); i++ {
	// 						if bytes.Contains(shardStates[i].CrossShard, []byte{shardID}) {
	// 							if shardStates[i].Height == blk.Header.Height {
	// 								continue
	// 							}
	// 							return NewBlockChainError(ShardStateError, errors.New("CrossOutput coin not in bytemap"))
	// 						}
	// 					}
	// 				}
	// 				if passed {
	// 					break
	// 				}
	// 			}
	// 		}
	// 	}
	// }
	Logger.log.Debugf("SHARD %+v | Finish VerifyPreProcessingShardBlock Block with height %+v at hash %+v", block.Header.ShardID, block.Header.Height, block.Hash())
	return nil
}

/*
	This function will verify the validation of a block with some best state in cache or current best state
	Get beacon state of this block
	For example, new blockHeight is 91 then beacon state of this block must have height 90
	OR new block has previous has is beacon best block hash
	- Producer
	- committee length and validatorIndex length
	- Producer + sig
	- Has parent hash is current best state best blockshard hash (compatible with current beststate)
	- Block Height
	- Beacon Height
	- Action root
*/
func (self *BestStateShard) VerifyBestStateWithShardBlock(block *ShardBlock, isVerifySig bool, shardID byte) error {
	Logger.log.Debugf("SHARD %+v | Begin VerifyBestStateWithShardBlock Block with height %+v at hash %+v", block.Header.ShardID, block.Header.Height, block.Hash())
	//TODO: define method to verify producer
	// Cal next producer
	// Verify next producer
	//=============Verify producer signature
	//==========TODO:UNCOMMENT to verify producer signature
	// producerPubkey := self.ShardCommittee[self.ShardProposerIdx]
	// blockHash := block.Header.Hash()
	// if err := cashec.ValidateDataB58(producerPubkey, block.ProducerSig, blockHash.GetBytes()); err != nil {
	// 	return NewBlockChainError(SignatureError, err)
	// }
	//=============End Verify producer signature
	//=============Verify aggegrate signature
	if isVerifySig {
		if len(block.ValidatorsIdx) < (len(self.ShardCommittee) >> 1) {
			return NewBlockChainError(SignatureError, errors.New("Block validators and Beacon committee is not compatible"))
		}
		ValidateAggSignature(block.ValidatorsIdx, self.ShardCommittee, block.AggregatedSig, block.R, block.Hash())
	}
	//=============End Verify Aggegrate signature
	if self.ShardHeight+1 != block.Header.Height {
		return NewBlockChainError(BlockHeightError, errors.New("Block height of new block should be : "+strconv.Itoa(int(self.ShardHeight+1))))
	}
	if block.Header.BeaconHeight < self.BeaconHeight {
		return NewBlockChainError(BlockHeightError, errors.New("Block contain invalid beacon height"))
	}
	Logger.log.Debugf("SHARD %+v | Finish VerifyBestStateWithShardBlock Block with height %+v at hash %+v", block.Header.ShardID, block.Header.Height, block.Hash())
	return nil
}

/*
	Update beststate with new block
		PrevShardBlockHash
		BestShardBlockHash
		BestBeaconHash
		BestShardBlock
		ShardHeight
		BeaconHeight
		ShardProposerIdx

		Add pending validator
		Swap shard committee if detect new epoch of beacon
*/
func (self *BestStateShard) Update(block *ShardBlock, beaconBlocks []*BeaconBlock) error {
	Logger.log.Debugf("SHARD %+v | Begin update Beststate with new Block with height %+v at hash %+v", block.Header.ShardID, block.Header.Height, block.Hash())
	var (
		err                   error
		shardSwapedCommittees []string
		shardNewCommittees    []string
	)
	self.PrevShardBlockHash = self.BestShardBlockHash
	self.BestShardBlockHash = *block.Hash()
	if block.Header.BeaconHeight == 1 {
		self.BestBeaconHash = *ChainTestParam.GenesisBeaconBlock.Hash()
	} else {
		self.BestBeaconHash = block.Header.BeaconHash
	}
	self.BestShardBlock = block
	self.ShardHeight = block.Header.Height
	self.Epoch = block.Header.Epoch
	self.BeaconHeight = block.Header.BeaconHeight
	self.ShardProposerIdx = common.IndexOfStr(block.Header.Producer, self.ShardCommittee)
	// Add pending validator
	for _, beaconBlock := range beaconBlocks {
		fmt.Println("ShardProcess/Update: BeaconBlock Height", beaconBlock.Header.Height)
		for _, l := range beaconBlock.Body.Instructions {
			if l[0] == "assign" && l[2] == "shard" {
				if l[3] == strconv.Itoa(int(block.Header.ShardID)) {
					Logger.log.Infof("SHARD %+v | Old ShardPendingValidatorList %+v", block.Header.ShardID, self.ShardPendingValidator)
					self.ShardPendingValidator = append(self.ShardPendingValidator, strings.Split(l[1], ",")...)
					Logger.log.Infof("SHARD %+v | New ShardPendingValidatorList %+v", block.Header.ShardID, self.ShardPendingValidator)
				}
			}
		}
	}
	fmt.Println("Shard Process/Update: ALL Instruction", block.Body.Instructions)
	// Swap committee
	for _, l := range block.Body.Instructions {
		fmt.Println("Shard Process/Update: Instruction", l)
		if l[0] == "swap" {
			fmt.Println("Shard Process/Update: ShardPendingValidator", self.ShardPendingValidator)
			fmt.Println("Shard Process/Update: ShardCommittee", self.ShardCommittee)
			self.ShardPendingValidator, self.ShardCommittee, shardSwapedCommittees, shardNewCommittees, err = SwapValidator(self.ShardPendingValidator, self.ShardCommittee, common.COMMITEES, common.OFFSET)
			if err != nil {
				Logger.log.Errorf("SHARD %+v | Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
				return NewBlockChainError(UnExpectedError, err)
			}
			swapedCommittees := strings.Split(l[2], ",")
			newCommittees := strings.Split(l[1], ",")
			if !reflect.DeepEqual(swapedCommittees, shardSwapedCommittees) {
				return NewBlockChainError(SwapError, errors.New("Invalid shard swapped committees"))
			}
			if !reflect.DeepEqual(newCommittees, shardNewCommittees) {
				return NewBlockChainError(SwapError, errors.New("Invalid shard new committees"))
			}
			Logger.log.Infof("SHARD %+v | Swap: Out committee %+v", block.Header.ShardID, shardSwapedCommittees)
			Logger.log.Infof("SHARD %+v | Swap: In committee %+v", block.Header.ShardID, shardNewCommittees)
		}
	}
	//Update best cross shard
	for shardID, crossShardBlock := range block.Body.CrossOutputCoin {
		self.BestCrossShard[shardID] = crossShardBlock[len(crossShardBlock)-1].BlockHeight
	}
	Logger.log.Debugf("SHARD %+v | Finish update Beststate with new Block with height %+v at hash %+v", block.Header.ShardID, block.Header.Height, block.Hash())
	return nil
}

/*
	VerifyPostProcessingShardBlock
	- commitee root
	- pending validator root
*/
func (self *BestStateShard) VerifyPostProcessingShardBlock(block *ShardBlock, shardID byte) error {
	var (
		isOk bool
	)
	Logger.log.Debugf("SHARD %+v | Begin VerifyPostProcessing Block with height %+v at hash %+v", block.Header.ShardID, block.Header.Height, block.Hash())
	isOk = VerifyHashFromStringArray(self.ShardCommittee, block.Header.CommitteeRoot)
	if !isOk {
		return NewBlockChainError(HashError, errors.New("Error verify Committee root"))
	}
	isOk = VerifyHashFromStringArray(self.ShardPendingValidator, block.Header.PendingValidatorRoot)
	if !isOk {
		return NewBlockChainError(HashError, errors.New("Error verify Pending validator root"))
	}
	Logger.log.Debugf("SHARD %+v | Finish VerifyPostProcessing Block with height %+v at hash %+v", block.Header.ShardID, block.Header.Height, block.Hash())
	return nil
}

//=====================Util for shard====================
//TODO: remove
//func CreateMerkleRootShard(txList []metadata.Transaction) common.Hash {
//	//calculate output coin hash for each shard
//	if len(txList) == 0 {
//		res, _ := GenerateZeroValueHash()
//		return res
//	}
//	outputCoinHash := getOutCoinHashEachShard(txList)
//	// calculate merkle data : 1 2 3 4 12 34 1234
//	merkleData := outputCoinHash
//	if len(merkleData)%2 == 1 {
//		merkleData = append(merkleData, common.HashH([]byte{}))
//	}
//
//	cursor := 0
//	for {
//		v1 := merkleData[cursor]
//		v2 := merkleData[cursor+1]
//		merkleData = append(merkleData, common.HashH(append(v1.GetBytes(), v2.GetBytes()...)))
//		cursor += 2
//		if cursor >= len(merkleData)-1 {
//			break
//		}
//	}
//	merkleShardRoot := merkleData[len(merkleData)-1]
//	return merkleShardRoot
//}

func CreateMerkleCrossOutputCoin(crossOutputCoins map[byte][]CrossOutputCoin) (*common.Hash, error) {
	if len(crossOutputCoins) == 0 {
		res, err := GenerateZeroValueHash()

		return &res, err
	}
	keys := []int{}
	crossOutputCoinHashes := []*common.Hash{}
	for k, _ := range crossOutputCoins {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		for _, value := range crossOutputCoins[byte(shardID)] {
			hash := value.Hash()
			hashByte := hash.GetBytes()
			newHash, err := common.Hash{}.NewHash(hashByte)
			if err != nil {
				return &common.Hash{}, NewBlockChainError(HashError, err)
			}
			crossOutputCoinHashes = append(crossOutputCoinHashes, newHash)
		}
	}
	merkle := Merkle{}
	merkleTree := merkle.BuildMerkleTreeOfHashs(crossOutputCoinHashes)
	return merkleTree[len(merkleTree)-1], nil
}

func VerifyMerkleCrossOutputCoin(crossOutputCoins map[byte][]CrossOutputCoin, rootHash common.Hash) bool {
	res, err := CreateMerkleCrossOutputCoin(crossOutputCoins)
	if err != nil {
		return false
	}
	hashByte := rootHash.GetBytes()
	newHash, err := common.Hash{}.NewHash(hashByte)
	return newHash.IsEqual(res)
}

func (self *BlockChain) StoreIncomingCrossShard(block *ShardBlock) error {
	crossShardMap, _ := block.Body.ExtractIncomingCrossShardMap()
	for crossShard, crossBlks := range crossShardMap {
		for _, crossBlk := range crossBlks {
			self.config.DataBase.StoreIncomingCrossShard(block.Header.ShardID, crossShard, block.Header.Height, &crossBlk)
		}
	}
	return nil
}

// func (self *BlockChain) StoreOutgoingCrossShard(block *ShardBlock) error {
// 	crossShardMap, _ := block.Body.ExtractOutgoingCrossShardMap()
// 	for crossShard, crossBlks := range crossShardMap {
// 		for _, crossBlk := range crossBlks {
// 			self.config.DataBase.StoreIncomingCrossShard(block.Header.ShardID, crossShard, block.Header.Height, &crossBlk)
// 		}
// 	}
// 	return nil
// }
