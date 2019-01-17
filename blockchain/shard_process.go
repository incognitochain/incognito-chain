package blockchain

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/ninjadotorg/constant/metadata"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/privacy"
)

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
	if strings.Compare(self.BestState.Beacon.BestBlockHash.String(), block.Header.PrevBlockHash.String()) == 0 {
		tempMarshal, err := json.Marshal(self.BestState.Beacon)
		if err != nil {
			return NewBlockChainError(UnmashallJsonBlockError, err)
		}
		json.Unmarshal(tempMarshal, &shardBestState)
	}
	// if no match best state found then block is unknown
	if reflect.DeepEqual(shardBestState, BestStateShard{}) {
		return NewBlockChainError(BeaconError, errors.New("Beacon Block does not match with any Beacon State in cache or in Database"))
	}
	// Verify block with previous best state
	// not verify agg signature in this function
	if err := shardBestState.VerifyBestStateWithShardBlock(block, false, shardID); err != nil {
		return err
	}
	//========Update best state with new block
	//TODO: update beacon block
	if err := shardBestState.Update(block, nil); err != nil {
		return err
	}
	//========Post verififcation: verify new beaconstate with corresponding block
	if err := shardBestState.VerifyPostProcessingShardBlock(block, shardID); err != nil {
		return err
	}
	Logger.log.Infof("Block %d, with hash %+v is VALID for signing", block.Header.Height, *block.Hash())
	return nil
}
func (self *BlockChain) ValidateShardBlockSignature(block *ShardBlock) error {
	// get shard id
	shardID := block.Header.ShardID
	// get best state shard committee corresponding to shardID
	bestStateShardCommittee := self.BestState.Shard[shardID].ShardCommittee

	pubKeysR := []*privacy.PublicKey{}
	for _, index := range block.ValidatorsIdx[0] {
		pubkeyBytes, _, err := base58.Base58Check{}.Decode(bestStateShardCommittee[index])
		if err != nil {
			return errors.New("Error in convert Public key from string to byte")
		}
		pubKey := privacy.PublicKey{}
		pubKey = pubkeyBytes
		pubKeysR = append(pubKeysR, &pubKey)
	}
	pubKeysAggSig := []*privacy.PublicKey{}
	for _, index := range block.ValidatorsIdx[1] {
		pubkeyBytes, _, err := base58.Base58Check{}.Decode(bestStateShardCommittee[index])
		if err != nil {
			return errors.New("Error in convert Public key from string to byte")
		}
		pubKey := privacy.PublicKey{}
		pubKey = pubkeyBytes
		pubKeysAggSig = append(pubKeysAggSig, &pubKey)
	}
	RCombined := new(privacy.EllipticPoint)
	RCombined.Set(big.NewInt(0), big.NewInt(0))
	Rbytesarr, byteVersion, err := base58.Base58Check{}.Decode(block.R)
	if (err != nil) || (byteVersion != byte(0x00)) {
		return err
	}
	err = RCombined.Decompress(Rbytesarr)
	if err != nil {
		return err
	}

	aggSig, _, err := base58.Base58Check{}.Decode(block.AggregatedSig)
	if err != nil {
		return errors.New("Error in convert aggregated signature from string to byte")
	}
	schnMultiSig := &privacy.SchnMultiSig{}
	schnMultiSig.SetBytes(aggSig)
	blockHash := block.Header.Hash()
	if schnMultiSig.VerifyMultiSig(blockHash.GetBytes(), pubKeysR, pubKeysAggSig, RCombined) == false {
		return errors.New("Invalid Agg signature")
	}
	return nil
}

func (self *BlockChain) ProcessStoreShardBlock(block *ShardBlock) error {
	blockHash := block.Hash().String()
	Logger.log.Infof("Process store block %+v", blockHash)

	if err := self.BestState.Shard[block.Header.ShardID].Update(block, nil); err != nil {
		return err
	}

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
		Logger.log.Infof("Transaction in block with hash", blockHash, "and index", index, ":", tx)
	}
	err := self.StoreCrossShard(block)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}
	return nil
}

func (self *BlockChain) StoreCrossShard(block *ShardBlock) error {
	crossShardMap, _ := block.Body.ExtractCrossShardMap()
	for crossShard, crossBlks := range crossShardMap {
		for _, crossBlk := range crossBlks {
			self.config.DataBase.StoreCrossShard(block.Header.ShardID, crossShard, block.Header.Height, &crossBlk)
		}
	}
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
	if err := self.BestState.Shard[shardID].Update(block, nil); err != nil {
		return err
	}

	Logger.log.Infof("SHARD %+v | Verify Post Processing Block %+v \n", block.Header.ShardID, *block.Hash())
	//========Post verififcation: verify new beaconstate with corresponding block
	if err := self.BestState.Shard[shardID].VerifyPostProcessingShardBlock(block, shardID); err != nil {
		return err
	}
	//========Store new Beaconblock and new Beacon bestState
	self.ProcessStoreShardBlock(block)
	Logger.log.Infof("SHARD %+v | Finish Insert new block %d, with hash %x", block.Header.ShardID, block.Header.Height, *block.Hash())
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
//TODO: define where to verify beacon info
- BeaconHeight
- BeaconHash
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
	if block.Header.Height%EPOCH == 0 && parentBlock.Header.Epoch != block.Header.Epoch-1 {
		return NewBlockChainError(EpochError, errors.New("Block height and Epoch is not compatiable"))
	}
	// Verify timestamp with parent block
	if block.Header.Timestamp <= parentBlock.Header.Timestamp {
		return NewBlockChainError(TimestampError, errors.New("Timestamp of new block can't equal to parent block"))
	}
	// Verify transaction root
	txMerkle := Merkle{}.BuildMerkleTreeStore(block.Body.Transactions)
	txRoot := txMerkle[len(txMerkle)-1]
	if bytes.Compare(block.Header.TxRoot.GetBytes(), txRoot.GetBytes()) != 0 {
		return NewBlockChainError(HashError, errors.New("Can't Verify Transaction Root"))
	}
	// Verify ShardTx Root
	shardTxRoot := CreateMerkleRootShard(block.Body.Transactions)
	if bytes.Compare(block.Header.ShardTxRoot.GetBytes(), shardTxRoot.GetBytes()) != 0 {
		return NewBlockChainError(HashError, errors.New("Can't Verify CrossShardTransaction Root"))
	}
	// Verify Crossoutput coin
	if !VerifyMerkleCrossOutputCoin(block.Body.CrossOutputCoin, block.Header.CrossOutputCoinRoot) {
		return NewBlockChainError(HashError, errors.New("Can't Verify CrossOutputCoin Root"))
	}
	// Verify Action
	actions := CreateShardActionFromTransaction(self.config.DataBase, block.Body.Transactions, block.Header.ShardID)
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
	if len(block.ValidatorsIdx) < (len(self.ShardCommittee) >> 1) {
		return NewBlockChainError(SignatureError, errors.New("Block validators and Beacon committee is not compatible"))
	}
	if isVerifySig {
		pubKeysR := []*privacy.PublicKey{}
		for _, index := range block.ValidatorsIdx[0] {
			pubkeyBytes, _, err := base58.Base58Check{}.Decode(self.ShardCommittee[index])
			if err != nil {
				return errors.New("Error in convert Public key from string to byte")
			}
			pubKey := privacy.PublicKey{}
			pubKey = pubkeyBytes
			pubKeysR = append(pubKeysR, &pubKey)
		}
		pubKeysAggSig := []*privacy.PublicKey{}
		for _, index := range block.ValidatorsIdx[1] {
			pubkeyBytes, _, err := base58.Base58Check{}.Decode(self.ShardCommittee[index])
			if err != nil {
				return errors.New("Error in convert Public key from string to byte")
			}
			pubKey := privacy.PublicKey{}
			pubKey = pubkeyBytes
			pubKeysAggSig = append(pubKeysAggSig, &pubKey)
		}
		RCombined := new(privacy.EllipticPoint)
		RCombined.Set(big.NewInt(0), big.NewInt(0))
		Rbytesarr, byteVersion, err := base58.Base58Check{}.Decode(block.R)
		if (err != nil) || (byteVersion != common.ZeroByte) {
			return err
		}
		err = RCombined.Decompress(Rbytesarr)
		if err != nil {
			return err
		}

		aggSig, _, err := base58.Base58Check{}.Decode(block.AggregatedSig)
		if err != nil {
			return errors.New("Error in convert aggregated signature from string to byte")
		}
		schnMultiSig := &privacy.SchnMultiSig{}
		schnMultiSig.SetBytes(aggSig)
		blockHash := block.Header.Hash()
		if schnMultiSig.VerifyMultiSig(blockHash.GetBytes(), pubKeysR, pubKeysAggSig, RCombined) == false {
			return errors.New("Invalid Agg signature")
		}
	}
	//=============End Verify Aggegrate signature
	if self.ShardHeight+1 != block.Header.Height {
		return NewBlockChainError(BlockHeightError, errors.New("Block height of new block should be : "+strconv.Itoa(int(block.Header.Height+1))))
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
	self.BestBeaconHash = block.Header.BeaconHash
	self.BestShardBlock = block
	self.ShardHeight = block.Header.Height
	prevBeaconHeight := self.BeaconHeight
	self.BeaconHeight = block.Header.BeaconHeight
	self.ShardProposerIdx = common.IndexOfStr(block.Header.Producer, self.ShardCommittee)
	// Add pending validator
	for _, beaconBlock := range beaconBlocks {
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
	// Swap committee
	if block.Header.BeaconHeight%EPOCH < prevBeaconHeight {
		self.ShardPendingValidator, self.ShardCommittee, shardSwapedCommittees, shardNewCommittees, err = SwapValidator(self.ShardPendingValidator, self.ShardCommittee, COMMITEES, OFFSET)
		if err != nil {
			Logger.log.Errorf("SHARD %+v | Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
			return NewBlockChainError(UnExpectedError, err)
		}
		Logger.log.Infof("SHARD %+v | Swap: Out committee %+v", block.Header.ShardID, shardSwapedCommittees)
		Logger.log.Infof("SHARD %+v | Swap: In committee %+v", block.Header.ShardID, shardNewCommittees)
	}
	Logger.log.Debugf("SHARD %+v | Finish update Beststate with new Block with height %+v at hash %+v", block.Header.ShardID, block.Header.Height, block.Hash())
	return nil
}
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
		return NewBlockChainError(HashError, errors.New("Error verify Pendinging validator root"))
	}
	Logger.log.Debugf("SHARD %+v | Finish VerifyPostProcessing Block with height %+v at hash %+v", block.Header.ShardID, block.Header.Height, block.Hash())
	return nil
}

//=====================Util for shard====================
func CreateMerkleRootShard(txList []metadata.Transaction) common.Hash {
	//calculate output coin hash for each shard
	outputCoinHash := getOutCoinHashEachShard(txList)
	// calculate merkle data : 1 2 3 4 12 34 1234
	merkleData := outputCoinHash
	if len(merkleData)%2 == 1 {
		merkleData = append(merkleData, common.HashH([]byte{}))
	}

	cursor := 0
	for {
		v1 := merkleData[cursor]
		v2 := merkleData[cursor+1]
		merkleData = append(merkleData, common.HashH(append(v1.GetBytes(), v2.GetBytes()...)))
		cursor += 2
		if cursor >= len(merkleData)-1 {
			break
		}
	}
	merkleShardRoot := merkleData[len(merkleData)-1]
	return merkleShardRoot
}

func CreateMerkleCrossOutputCoin(crossOutputCoins []CrossOutputCoin) (*common.Hash, error) {
	crossOutputCoinHashes := []*common.Hash{}
	for _, value := range crossOutputCoins {
		hash := value.Hash()
		hashByte := hash.GetBytes()
		newHash, err := common.Hash{}.NewHash(hashByte)
		if err != nil {
			return &common.Hash{}, NewBlockChainError(HashError, err)
		}
		crossOutputCoinHashes = append(crossOutputCoinHashes, newHash)
	}
	merkle := Merkle{}
	merkleTree := merkle.BuildMerkleTreeOfHashs(crossOutputCoinHashes)
	return merkleTree[len(merkleTree)-1], nil
}

func VerifyMerkleCrossOutputCoin(crossOutputCoins []CrossOutputCoin, rootHash common.Hash) bool {
	res, err := CreateMerkleCrossOutputCoin(crossOutputCoins)
	if err != nil {
		return false
	}
	hashByte := rootHash.GetBytes()
	newHash, err := common.Hash{}.NewHash(hashByte)
	return newHash.IsEqual(res)
}
