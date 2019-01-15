package blockchain

import (
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

func (self *BlockChain) VerifyPreSignShardBlock(block *ShardBlock, shardId byte) error {
	self.chainLock.Lock()
	defer self.chainLock.Unlock()
	//========Verify block only
	Logger.log.Infof("Verify block for signing process %d, with hash %+v", block.Header.Height, *block.Hash())
	if err := self.VerifyPreProcessingShardBlock(block, shardId); err != nil {
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
	if err := shardBestState.VerifyBestStateWithShardBlock(block, false, shardId); err != nil {
		return err
	}
	//========Update best state with new block
	if err := shardBestState.Update(block); err != nil {
		return err
	}
	//========Post verififcation: verify new beaconstate with corresponding block
	if err := shardBestState.VerifyPostProcessingShardBlock(block, shardId); err != nil {
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

	if err := self.BestState.Shard[block.Header.ShardID].Update(block); err != nil {
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
	return nil
}

func (self *BlockChain) InsertShardBlock(block *ShardBlock) error {
	blockHash := block.Hash().String()
	Logger.log.Infof("Insert block %+v", blockHash)

	if err := self.ValidateShardBlockSignature(block); err != nil {
		return err
	}

	return self.ProcessStoreShardBlock(block)
}

func (self *BlockChain) VerifyPreProcessingShardBlock(block *ShardBlock, shardID byte) error {
	/* Verify Pre-prosessing data
	This function DOES NOT verify new block with best state
	DO NOT USE THIS with GENESIS BLOCK
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
	*/
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

	return nil
}

func (self *BestStateShard) VerifyPostProcessingShardBlock(block *ShardBlock, shardId byte) error {
	return nil
}

func (self *BestStateShard) Update(block *ShardBlock) error {
	self.BestBeaconHash = block.Header.BeaconHash
	self.BestShardBlockHash = *block.Hash()
	self.ShardHeight = block.Header.Height
	self.BeaconHeight = block.Header.BeaconHeight

	self.ShardProposerIdx = common.IndexOfStr(block.Header.Producer, self.ShardCommittee)

	//self.TotalTxns += uint64(len(block.Body.Transactions))
	//self.NumTxns = uint64(len(block.Body.Transactions))
	return nil
}

func (self *BestStateShard) VerifyBestStateWithShardBlock(block *ShardBlock, isSign bool, shardId byte) error {
	return nil
}

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
