package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
)

/*
	// This function should receives block in consensus round
	// It verify validity of this function before sign it
	// This should be verify in the first round of consensus

	Step:
	1. Verify Pre proccessing data
	2. Retrieve beststate for new block, store in local variable
	3. Update: process local beststate with new block
	4. Verify Post processing: updated local beststate and newblock

	Return:
	- No error: valid and can be sign
	- Error: invalid new block
*/
func (blockchain *BlockChain) VerifyPreSignBeaconBlock(block *BeaconBlock, isCommittee bool) error {
	blockchain.chainLock.Lock()
	defer blockchain.chainLock.Unlock()
	//========Verify block only
	Logger.log.Infof("Verify block for signing process %d, with hash %+v", block.Header.Height, *block.Hash())
	if err := blockchain.VerifyPreProcessingBeaconBlock(block, isCommittee); err != nil {
		return err
	}
	//========Verify block with previous best state
	// Get Beststate of previous block == previous best state
	// Clone best state value into new variable
	beaconBestState := BestStateBeacon{}
	// check with current final best state
	// New block must be compatible with current best state
	if strings.Compare(blockchain.BestState.Beacon.BestBlockHash.String(), block.Header.PrevBlockHash.String()) == 0 {
		tempMarshal, err := json.Marshal(blockchain.BestState.Beacon)
		if err != nil {
			return NewBlockChainError(UnmashallJsonBlockError, err)
		}
		json.Unmarshal(tempMarshal, &beaconBestState)
	}
	// if no match best state found then block is unknown
	if reflect.DeepEqual(beaconBestState, BestStateBeacon{}) {
		return NewBlockChainError(BeaconError, errors.New("beacon Block does not match with any Beacon State in cache or in Database"))
	}
	// Verify block with previous best state
	// not verify agg signature in this function
	if err := beaconBestState.VerifyBestStateWithBeaconBlock(block, false); err != nil {
		return err
	}
	//========Update best state with new block
	snapShotBeaconCommittee := beaconBestState.BeaconCommittee
	if err := beaconBestState.Update(block); err != nil {
		return err
	}
	//========Post verififcation: verify new beaconstate with corresponding block
	if err := beaconBestState.VerifyPostProcessingBeaconBlock(block, snapShotBeaconCommittee); err != nil {
		return err
	}
	Logger.log.Infof("Block %d, with hash %+v is VALID for signing", block.Header.Height, *block.Hash())
	return nil
}

func (blockchain *BlockChain) InsertBeaconBlock(block *BeaconBlock, isCommittee bool) error {
	blockchain.chainLock.Lock()
	defer blockchain.chainLock.Unlock()
	Logger.log.Infof("Begin Insert new block %d, with hash %+v \n", block.Header.Height, *block.Hash())
	// fmt.Printf("Beacon block %+v \n", block)
	Logger.log.Infof("Verify Pre Processing Beacon Block %+v \n", *block.Hash())
	if err := blockchain.VerifyPreProcessingBeaconBlock(block, isCommittee); err != nil {
		return err
	}
	//========Verify block with previous best state
	// check with current final best state
	// block can only be insert if it match the current best state
	if strings.Compare(blockchain.BestState.Beacon.BestBlockHash.String(), block.Header.PrevBlockHash.String()) != 0 {
		return NewBlockChainError(BeaconError, errors.New("beacon Block does not match with any Beacon State in cache or in Database"))
	}
	// fmt.Printf("BeaconBest state %+v \n", blockchain.BestState.Beacon)
	Logger.log.Infof("Verify BestState with Beacon Block %+v \n", *block.Hash())
	// Verify block with previous best state
	if err := blockchain.BestState.Beacon.VerifyBestStateWithBeaconBlock(block, true); err != nil {
		return err
	}
	Logger.log.Infof("Update BestState with Beacon Block %+v \n", *block.Hash())
	//========Update best state with new block
	snapShotBeaconCommittee := blockchain.BestState.Beacon.BeaconCommittee
	if err := blockchain.BestState.Beacon.Update(block); err != nil {
		return err
	}
	Logger.log.Infof("Verify Post Processing Beacon Block %+v \n", *block.Hash())
	//========Post verififcation: verify new beaconstate with corresponding block
	if err := blockchain.BestState.Beacon.VerifyPostProcessingBeaconBlock(block, snapShotBeaconCommittee); err != nil {
		return err
	}
	//========Store new Beaconblock and new Beacon bestState in cache
	Logger.log.Infof("Store Beacon BestState %+v \n", *block.Hash())
	if err := blockchain.config.DataBase.StoreBeaconBestState(blockchain.BestState.Beacon); err != nil {
		return err
	}
	Logger.log.Infof("Store Beacon Block %+v \n", *block.Hash())
	if err := blockchain.config.DataBase.StoreBeaconBlock(block); err != nil {
		return err
	}
	blockHash := block.Hash()
	if err := blockchain.config.DataBase.StoreBeaconBlockIndex(blockHash, block.Header.Height); err != nil {
		return err
	}
	for shardID, shardStates := range block.Body.ShardState {
		for _, shardState := range shardStates {
			blockchain.config.DataBase.StoreAcceptedShardToBeacon(shardID, block.Header.Height, &shardState.Hash)
		}
	}
	if err := blockchain.config.DataBase.StoreBeaconCommitteeByHeight(block.Header.Height, blockchain.BestState.Beacon.ShardCommittee); err != nil {
		return err
	}
	//=========Remove shard block in beacon pool
	Logger.log.Infof("Remove block from pool %+v \n", *block.Hash())
	blockchain.config.ShardToBeaconPool.SetShardState(blockchain.BestState.Beacon.BestShardHeight)

	Logger.log.Infof("Finish Insert new block %d, with hash %+v", block.Header.Height, *block.Hash())
	return nil
}

/* Verify Pre-prosessing data
This function DOES NOT verify new block with best state
DO NOT USE THIS with GENESIS BLOCK
- Producer validity
- version
- parent hash
- Height = parent hash + 1
- Epoch = blockHeight % Epoch ? Parent Epoch + 1
- Timestamp can not excess some limit
- Instruction hash
- ShardStateHash
- ShardState is sorted?
FOR CURRENT COMMITTEES ONLY
	- Is shardState existed in pool
*/
func (blockchain *BlockChain) VerifyPreProcessingBeaconBlock(block *BeaconBlock, isCommittee bool) error {
	//verify producer
	producerPosition := (blockchain.BestState.Beacon.BeaconProposerIdx + 1) % len(blockchain.BestState.Beacon.BeaconCommittee)
	tempProducer := blockchain.BestState.Beacon.BeaconCommittee[producerPosition]
	if strings.Compare(tempProducer, block.Header.Producer) != 0 {
		return NewBlockChainError(ProducerError, errors.New("Producer should be should be :"+tempProducer))
	}
	//verify version
	if block.Header.Version != VERSION {
		return NewBlockChainError(VersionError, errors.New("Version should be :"+strconv.Itoa(VERSION)))
	}
	prevBlockHash := block.Header.PrevBlockHash
	// Verify parent hash exist or not
	parentBlock, err := blockchain.config.DataBase.FetchBeaconBlock(&prevBlockHash)
	if err != nil {
		return NewBlockChainError(DBError, err)
	}
	parentBlockInterface := NewBeaconBlock()
	json.Unmarshal(parentBlock, &parentBlockInterface)
	// Verify block height with parent block
	if parentBlockInterface.Header.Height+1 != block.Header.Height {
		return NewBlockChainError(BlockHeightError, errors.New("block height of new block should be :"+strconv.Itoa(int(block.Header.Height+1))))
	}
	// Verify epoch with parent block
	if (block.Header.Height != 1) && (block.Header.Height%common.EPOCH == 1) && (parentBlockInterface.Header.Epoch != block.Header.Epoch-1) {
		return NewBlockChainError(EpochError, errors.New("lock height and Epoch is not compatiable"))
	}
	// Verify timestamp with parent block
	if block.Header.Timestamp <= parentBlockInterface.Header.Timestamp {
		return NewBlockChainError(TimestampError, errors.New("timestamp of new block can't equal to parent block"))
	}

	if !VerifyHashFromShardState(block.Body.ShardState, block.Header.ShardStateHash) {
		return NewBlockChainError(ShardStateHashError, errors.New("shard state hash is not correct"))
	}

	tempInstructionArr := []string{}
	for _, strs := range block.Body.Instructions {
		tempInstructionArr = append(tempInstructionArr, strs...)
	}
	if !VerifyHashFromStringArray(tempInstructionArr, block.Header.InstructionHash) {
		return NewBlockChainError(InstructionHashError, errors.New("instruction hash is not correct"))
	}
	// Shard state must in right format
	// state[i].Height must less than state[i+1].Height and state[i+1].Height - state[i].Height = 1
	for _, shardStates := range block.Body.ShardState {
		for i := 0; i < len(shardStates)-2; i++ {
			if shardStates[i+1].Height-shardStates[i].Height != 1 {
				return NewBlockChainError(ShardStateError, errors.New("shardstates are not in right format"))
			}
		}
	}
	// if pool does not have one of needed block, fail to verify
	if isCommittee {
		allShardBlocks := blockchain.config.ShardToBeaconPool.GetValidPendingBlock()
		for shardID, shardBlocks := range allShardBlocks {
			shardBlocks = shardBlocks[:len(block.Body.ShardState[shardID])]
			shardStates := block.Body.ShardState[shardID]
			for index, shardState := range shardStates {
				if shardBlocks[index].Header.Height != shardState.Height {
					return NewBlockChainError(ShardStateError, errors.New("shardstate fail to verify with ShardToBeacon Block in pool"))
				}
				blockHash := shardBlocks[index].Header.Hash()
				if strings.Compare(blockHash.String(), shardState.Hash.String()) != 0 {
					return NewBlockChainError(ShardStateError, errors.New("shardstate fail to verify with ShardToBeacon Block in pool"))
				}
				if !reflect.DeepEqual(shardBlocks[index].Header.CrossShards, shardState.CrossShard) {
					return NewBlockChainError(ShardStateError, errors.New("shardstate fail to verify with ShardToBeacon Block in pool"))
				}
			}
			// Only accept block in one epoch
			for index, shardBlock := range shardBlocks {
				currentCommittee := blockchain.BestState.Beacon.ShardCommittee[shardID]
				currentPendingValidator := blockchain.BestState.Beacon.ShardPendingValidator[shardID]
				hash := shardBlock.Header.Hash()
				err := ValidateAggSignature(shardBlock.ValidatorsIdx, currentCommittee, shardBlock.AggregatedSig, shardBlock.R, &hash)
				if index == 0 && err != nil {
					currentCommittee, _, _, _, err = SwapValidator(currentPendingValidator, currentCommittee, blockchain.BestState.Beacon.ShardCommitteeSize, common.OFFSET)
					if err != nil {
						return NewBlockChainError(ShardStateError, errors.New("shardstate fail to verify with ShardToBeacon Block in pool"))
					}
					err = ValidateAggSignature(shardBlock.ValidatorsIdx, currentCommittee, shardBlock.AggregatedSig, shardBlock.R, &hash)
					if err != nil {
						return NewBlockChainError(ShardStateError, errors.New("shardstate fail to verify with ShardToBeacon Block in pool"))
					}
				}
				if index != 0 && err != nil {
					return NewBlockChainError(ShardStateError, errors.New("shardstate fail to verify with ShardToBeacon Block in pool"))
				}
			}
		}
	}

	return nil
}

/*
	This function will verify the validation of a block with some best state in cache or current best state
	Get beacon state of this block
	For example, new blockHeight is 91 then beacon state of this block must have height 90
	OR new block has previous has is beacon best block hash
	- Committee length and validatorIndex length
	- Producer + sig
	- Has parent hash is current best block hash in best state
	- Height
	- Epoch
	- staker
	- ShardState
*/
func (bestStateBeacon *BestStateBeacon) VerifyBestStateWithBeaconBlock(block *BeaconBlock, isVerifySig bool) error {
	//=============Verify aggegrate signature
	if isVerifySig {
		// ValidatorIdx must > Number of Beacon Committee / 2 AND Number of Beacon Committee > 3
		if len(block.ValidatorsIdx) <= (len(bestStateBeacon.BeaconCommittee)>>1) && len(bestStateBeacon.BeaconCommittee) > 3 {
			return NewBlockChainError(SignatureError, errors.New("block validators and Beacon committee is not compatible"))
		}
		err := ValidateAggSignature(block.ValidatorsIdx, bestStateBeacon.BeaconCommittee, block.AggregatedSig, block.R, block.Hash())
		if err != nil {
			return NewBlockChainError(SignatureError, err)
		}
	}
	//=============End Verify Aggegrate signature
	if bestStateBeacon.BeaconHeight+1 != block.Header.Height {
		return NewBlockChainError(BlockHeightError, errors.New("block height of new block should be :"+strconv.Itoa(int(block.Header.Height+1))))
	}
	if !bytes.Equal(bestStateBeacon.BestBlockHash.GetBytes(), block.Header.PrevBlockHash.GetBytes()) {
		return NewBlockChainError(BlockHeightError, errors.New("previous us block should be :"+bestStateBeacon.BestBlockHash.String()))
	}
	if block.Header.Height%common.EPOCH == 1 && bestStateBeacon.Epoch+1 != block.Header.Epoch {
		return NewBlockChainError(EpochError, errors.New("block height and Epoch is not compatiable"))
	}
	if block.Header.Height%common.EPOCH != 1 && bestStateBeacon.Epoch != block.Header.Epoch {
		return NewBlockChainError(EpochError, errors.New("block height and Epoch is not compatiable"))
	}
	//=============Verify Stakers
	newBeaconCandidate, newShardCandidate := GetStakingCandidate(*block)
	if !reflect.DeepEqual(newBeaconCandidate, []string{}) {
		validBeaconCandidate := bestStateBeacon.GetValidStakers(newBeaconCandidate)
		if !reflect.DeepEqual(validBeaconCandidate, newBeaconCandidate) {
			return NewBlockChainError(CandidateError, errors.New("beacon candidate list is INVALID"))
		}
	}
	if !reflect.DeepEqual(newShardCandidate, []string{}) {
		validShardCandidate := bestStateBeacon.GetValidStakers(newShardCandidate)
		if !reflect.DeepEqual(validShardCandidate, newShardCandidate) {
			return NewBlockChainError(CandidateError, errors.New("shard candidate list is INVALID"))
		}
	}
	//=============End Verify Stakers
	// Verify shard state
	for shardID, shardStates := range block.Body.ShardState {
		// Do not check this condition with first minted block (genesis block height = 1)
		if bestStateBeacon.BeaconHeight != 2 {
			fmt.Printf("Beacon Process/Check ShardStates with BestState Current Shard Height %+v \n", bestStateBeacon.AllShardState[shardID][len(bestStateBeacon.AllShardState[shardID])-1].Height)
			fmt.Printf("Beacon Process/Check ShardStates with BestState FirstShardHeight %+v \n", shardStates[0].Height)
			if shardStates[0].Height-bestStateBeacon.AllShardState[shardID][len(bestStateBeacon.AllShardState[shardID])-1].Height != 1 {
				return NewBlockChainError(ShardStateError, errors.New("Shardstates are not compatible with beacon best state"))
			}
		}
	}
	return nil
}

/* Verify Post-processing data
- Validator root: BeaconCommittee + BeaconPendingValidator
- Beacon Candidate root: CandidateBeaconWaitingForCurrentRandom + CandidateBeaconWaitingForNextRandom
- Shard Candidate root: CandidateShardWaitingForCurrentRandom + CandidateShardWaitingForNextRandom
- Shard Validator root: ShardCommittee + ShardPendingValidator
- Random number if have in instruction
*/
func (bestStateBeacon *BestStateBeacon) VerifyPostProcessingBeaconBlock(block *BeaconBlock, snapShotBeaconCommittee []string) error {
	var (
		strs []string
		isOk bool
	)
	//=============Verify producer signature
	producerPubkey := snapShotBeaconCommittee[bestStateBeacon.BeaconProposerIdx]
	blockHash := block.Header.Hash()
	if err := cashec.ValidateDataB58(producerPubkey, block.ProducerSig, blockHash.GetBytes()); err != nil {
		return NewBlockChainError(SignatureError, err)
	}
	//=============End Verify producer signature
	strs = append(strs, bestStateBeacon.BeaconCommittee...)
	strs = append(strs, bestStateBeacon.BeaconPendingValidator...)
	isOk = VerifyHashFromStringArray(strs, block.Header.ValidatorsRoot)
	if !isOk {
		return NewBlockChainError(HashError, errors.New("error verify Beacon Validator root"))
	}

	strs = []string{}
	strs = append(strs, bestStateBeacon.CandidateBeaconWaitingForCurrentRandom...)
	strs = append(strs, bestStateBeacon.CandidateBeaconWaitingForNextRandom...)
	isOk = VerifyHashFromStringArray(strs, block.Header.BeaconCandidateRoot)
	if !isOk {
		return NewBlockChainError(HashError, errors.New("error verify Beacon Candidate root"))
	}

	strs = []string{}
	strs = append(strs, bestStateBeacon.CandidateShardWaitingForCurrentRandom...)
	strs = append(strs, bestStateBeacon.CandidateShardWaitingForNextRandom...)
	isOk = VerifyHashFromStringArray(strs, block.Header.ShardCandidateRoot)
	if !isOk {
		return NewBlockChainError(HashError, errors.New("error verify Shard Candidate root"))
	}

	isOk = VerifyHashFromMapByteString(bestStateBeacon.ShardPendingValidator, bestStateBeacon.ShardCommittee, block.Header.ShardValidatorsRoot)
	if !isOk {
		return NewBlockChainError(HashError, errors.New("error verify shard validator root"))
	}

	// COMMENT FOR TESTING
	// instructions := block.Body.Instructions
	// for _, l := range instructions {
	// 	if l[0] == "random" {
	// 		temp, err := strconv.Atoi(l[3])
	// 		if err != nil {
	// 			Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
	// 			return NewBlockChainError(UnExpectedError, err)
	// 		}
	// 		isOk, err = btcapi.VerifyNonceWithTimestamp(bestStateBeacon.CurrentRandomTimeStamp, int64(temp))
	// 		Logger.log.Infof("Verify Random number %+v", isOk)
	// 		if err != nil {
	// 			Logger.log.Error("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
	// 			return NewBlockChainError(UnExpectedError, err)
	// 		}
	// 		if !isOk {
	// 			return NewBlockChainError(RandomError, errors.New("Error verify random number"))
	// 		}
	// 	}
	// }
	return nil
}

/*
	Update Beststate with new Block
*/
func (bestStateBeacon *BestStateBeacon) Update(newBlock *BeaconBlock) error {
	newBeaconCandidate := []string{}
	newShardCandidate := []string{}
	// Logger.log.Infof("Start processing new block at height %d, with hash %+v", newBlock.Header.Height, *newBlock.Hash())
	if newBlock == nil {
		return errors.New("null pointer")
	}
	// signal of random parameter from beacon block
	randomFlag := false
	// update BestShardHash, BestBlock, BestBlockHash
	bestStateBeacon.PrevBestBlockHash = bestStateBeacon.BestBlockHash
	bestStateBeacon.BestBlockHash = *newBlock.Hash()
	bestStateBeacon.BestBlock = newBlock
	bestStateBeacon.Epoch = newBlock.Header.Epoch
	bestStateBeacon.BeaconHeight = newBlock.Header.Height
	bestStateBeacon.BeaconProposerIdx = common.IndexOfStr(newBlock.Header.Producer, bestStateBeacon.BeaconCommittee)

	allShardState := newBlock.Body.ShardState
	if bestStateBeacon.AllShardState == nil {
		bestStateBeacon.AllShardState = make(map[byte][]ShardState)
		for index := 0; index < common.MAX_SHARD_NUMBER; index++ {
			bestStateBeacon.AllShardState[byte(index)] = []ShardState{
				ShardState{
					Height: 1,
				},
			}
		}
	}
	if bestStateBeacon.BestShardHash == nil {
		bestStateBeacon.BestShardHash = make(map[byte]common.Hash)
	}
	if bestStateBeacon.BestShardHeight == nil {
		bestStateBeacon.BestShardHeight = make(map[byte]uint64)
	}
	// Update new best new block hash
	for shardID, shardStates := range allShardState {
		bestStateBeacon.BestShardHash[shardID] = shardStates[len(shardStates)-1].Hash
		bestStateBeacon.BestShardHeight[shardID] = shardStates[len(shardStates)-1].Height
		if _, ok := bestStateBeacon.AllShardState[shardID]; !ok {
			bestStateBeacon.AllShardState[shardID] = []ShardState{}
		}
		bestStateBeacon.AllShardState[shardID] = append(bestStateBeacon.AllShardState[shardID], shardStates...)
	}

	// update param
	instructions := newBlock.Body.Instructions
	for _, l := range instructions {
		// For stability instructions
		err := bestStateBeacon.processStabilityInstruction(l)
		if err != nil {
			fmt.Println(err)
		}

		if l[0] == "set" {
			bestStateBeacon.Params[l[1]] = l[2]
		}
		if l[0] == "del" {
			delete(bestStateBeacon.Params, l[1])
		}
		if l[0] == "swap" {
			fmt.Println("---------------============= SWAP", l)
			// format
			// ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "shard" "shardID"]
			// ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "beacon"]
			inPubkeys := strings.Split(l[1], ",")
			outPubkeys := strings.Split(l[2], ",")
			fmt.Println("---------------============= SWAP l1", l[1])
			fmt.Println("---------------============= SWAP l2", l[2])
			fmt.Println("---------------============= SWAP inPubkeys", inPubkeys)
			fmt.Println("---------------============= SWAP outPubkeys", outPubkeys)
			if l[3] == "shard" {
				temp, err := strconv.Atoi(l[4])
				if err != nil {
					Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
					return NewBlockChainError(UnExpectedError, err)
				}
				shardID := byte(temp)
				// delete in public key out of sharding pending validator list
				if len(l[1]) > 0 {
					fmt.Println("Beacon Process/Update Before, ShardPendingValidator", bestStateBeacon.ShardPendingValidator[shardID])
					bestStateBeacon.ShardPendingValidator[shardID], err = RemoveValidator(bestStateBeacon.ShardPendingValidator[shardID], inPubkeys)
					fmt.Println("Beacon Process/Update After, ShardPendingValidator", bestStateBeacon.ShardPendingValidator[shardID])
					if err != nil {
						Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
						return NewBlockChainError(UnExpectedError, err)
					}
					// append in public key to committees
					bestStateBeacon.ShardCommittee[shardID] = append(bestStateBeacon.ShardCommittee[shardID], inPubkeys...)
					fmt.Println("Beacon Process/Update Add New, ShardCommitees", bestStateBeacon.ShardCommittee[shardID])
				}
				// delete out public key out of current committees
				if len(l[2]) > 0 {
					bestStateBeacon.ShardCommittee[shardID], err = RemoveValidator(bestStateBeacon.ShardCommittee[shardID], outPubkeys)
					fmt.Println("Beacon Process/Update Remove Old, ShardCommitees", bestStateBeacon.ShardCommittee[shardID])
					if err != nil {
						Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
						return NewBlockChainError(UnExpectedError, err)
					}
				}
			} else if l[3] == "beacon" {
				var err error
				if len(l[1]) > 0 {
					bestStateBeacon.BeaconPendingValidator, err = RemoveValidator(bestStateBeacon.BeaconPendingValidator, inPubkeys)
					if err != nil {
						Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
						return NewBlockChainError(UnExpectedError, err)
					}
					bestStateBeacon.BeaconCommittee = append(bestStateBeacon.BeaconCommittee, inPubkeys...)
				}
				if len(l[2]) > 0 {
					bestStateBeacon.BeaconCommittee, err = RemoveValidator(bestStateBeacon.BeaconCommittee, outPubkeys)
					if err != nil {
						Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
						return NewBlockChainError(UnExpectedError, err)
					}
				}
			}
		}
		// ["random" "{nonce}" "{blockheight}" "{timestamp}" "{bitcoinTimestamp}"]
		if l[0] == "random" {
			temp, err := strconv.Atoi(l[1])
			if err != nil {
				Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
				return NewBlockChainError(UnExpectedError, err)
			}
			bestStateBeacon.CurrentRandomNumber = int64(temp)
			Logger.log.Info("Random number found %+v", bestStateBeacon.CurrentRandomNumber)
			randomFlag = true
		}
		// Update candidate
		// get staking candidate list and store
		// store new staking candidate
		if l[0] == "stake" && l[2] == "beacon" {
			beacon := strings.Split(l[1], ",")
			newBeaconCandidate = append(newBeaconCandidate, beacon...)
		}
		if l[0] == "stake" && l[2] == "shard" {
			shard := strings.Split(l[1], ",")
			newShardCandidate = append(newShardCandidate, shard...)
		}
	}
	if bestStateBeacon.BeaconHeight == 1 {
		// Assign committee with genesis block
		Logger.log.Infof("Proccessing Genesis Block")
		//Test with 1 member
		bestStateBeacon.BeaconCommittee = make([]string, bestStateBeacon.BeaconCommitteeSize)
		copy(bestStateBeacon.BeaconCommittee, newBeaconCandidate[:bestStateBeacon.BeaconCommitteeSize])
		for shardID := 0; shardID < bestStateBeacon.ActiveShards; shardID++ {
			bestStateBeacon.ShardCommittee[byte(shardID)] = append(bestStateBeacon.ShardCommittee[byte(shardID)], newShardCandidate[shardID*bestStateBeacon.ShardCommitteeSize:(shardID+1)*bestStateBeacon.ShardCommitteeSize]...)
		}
		bestStateBeacon.Epoch = 1
	} else {
		bestStateBeacon.CandidateBeaconWaitingForNextRandom = append(bestStateBeacon.CandidateBeaconWaitingForNextRandom, newBeaconCandidate...)
		bestStateBeacon.CandidateShardWaitingForNextRandom = append(bestStateBeacon.CandidateShardWaitingForNextRandom, newShardCandidate...)
		fmt.Println("Beacon Process/Before: CandidateShardWaitingForNextRandom: ", bestStateBeacon.CandidateShardWaitingForNextRandom)
	}

	if bestStateBeacon.BeaconHeight%common.EPOCH == 1 && bestStateBeacon.BeaconHeight != 1 {
		bestStateBeacon.IsGetRandomNumber = false
		// Begin of each epoch
	} else if bestStateBeacon.BeaconHeight%common.EPOCH < common.RANDOM_TIME {
		// Before get random from bitcoin
	} else if bestStateBeacon.BeaconHeight%common.EPOCH >= common.RANDOM_TIME {
		// After get random from bitcoin
		if bestStateBeacon.BeaconHeight%common.EPOCH == common.RANDOM_TIME {
			// snapshot candidate list
			bestStateBeacon.CandidateShardWaitingForCurrentRandom = bestStateBeacon.CandidateShardWaitingForNextRandom
			bestStateBeacon.CandidateBeaconWaitingForCurrentRandom = bestStateBeacon.CandidateBeaconWaitingForNextRandom
			fmt.Println("==================Beacon Process: Snapshot candidate====================")
			fmt.Println("Beacon Process: CandidateShardWaitingForCurrentRandom: ", bestStateBeacon.CandidateShardWaitingForCurrentRandom)
			fmt.Println("Beacon Process: CandidateBeaconWaitingForCurrentRandom: ", bestStateBeacon.CandidateBeaconWaitingForCurrentRandom)
			// reset candidate list
			bestStateBeacon.CandidateShardWaitingForNextRandom = []string{}
			bestStateBeacon.CandidateBeaconWaitingForNextRandom = []string{}
			fmt.Println("Beacon Process/After: CandidateShardWaitingForNextRandom: ", bestStateBeacon.CandidateShardWaitingForNextRandom)
			fmt.Println("Beacon Process/After: CandidateBeaconWaitingForCurrentRandom: ", bestStateBeacon.CandidateBeaconWaitingForCurrentRandom)
			// assign random timestamp
			bestStateBeacon.CurrentRandomTimeStamp = newBlock.Header.Timestamp
		}
		// if get new random number
		// Assign candidate to shard
		// assign CandidateShardWaitingForCurrentRandom to ShardPendingValidator with CurrentRandom
		if randomFlag {
			bestStateBeacon.IsGetRandomNumber = true
			fmt.Println("Beacon Process/Update/RandomFlag: Shard Candidate Waiting for Current Random Number", bestStateBeacon.CandidateShardWaitingForCurrentRandom)
			err := AssignValidatorShard(bestStateBeacon.ShardPendingValidator, bestStateBeacon.CandidateShardWaitingForCurrentRandom, bestStateBeacon.CurrentRandomNumber, bestStateBeacon.ActiveShards)
			if err != nil {
				Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
				return NewBlockChainError(UnExpectedError, err)
			}
			// delete CandidateShardWaitingForCurrentRandom list
			bestStateBeacon.CandidateShardWaitingForCurrentRandom = []string{}
			fmt.Println("Beacon Process/Update/RandomFalg: Shard Pending Validator", bestStateBeacon.ShardPendingValidator)
			// Shuffle candidate
			// shuffle CandidateBeaconWaitingForCurrentRandom with current random number
			fmt.Println("Beacon Process/Update/RandomFlag: Beacon Candidate Waiting for Current Random Number", bestStateBeacon.CandidateBeaconWaitingForCurrentRandom)
			newBeaconPendingValidator, err := ShuffleCandidate(bestStateBeacon.CandidateBeaconWaitingForCurrentRandom, bestStateBeacon.CurrentRandomNumber)
			fmt.Println("Beacon Process/Update/RandomFalg: NewBeaconPendingValidator", newBeaconPendingValidator)
			if err != nil {
				Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
				return NewBlockChainError(UnExpectedError, err)
			}
			bestStateBeacon.CandidateBeaconWaitingForCurrentRandom = []string{}
			bestStateBeacon.BeaconPendingValidator = append(bestStateBeacon.BeaconPendingValidator, newBeaconPendingValidator...)
			fmt.Println("Beacon Process/Update/RandomFalg: Beacon Pending Validator", bestStateBeacon.BeaconPendingValidator)
			if err != nil {
				return err
			}
		}
	} else if bestStateBeacon.BeaconHeight%common.EPOCH == 0 {
		// At the end of each epoch, eg: block 200, 400, 600 with epoch is 200 blocks
		// Swap pending validator in committees, pop some of public key in committees out
		// ONLY SWAP FOR BEACON
		// SHARD WILL SWAP ITblockchain
		var (
			beaconSwapedCommittees []string
			beaconNewCommittees    []string
			err                    error
		)
		bestStateBeacon.BeaconPendingValidator, bestStateBeacon.BeaconCommittee, beaconSwapedCommittees, beaconNewCommittees, err = SwapValidator(bestStateBeacon.BeaconPendingValidator, bestStateBeacon.BeaconCommittee, bestStateBeacon.BeaconCommitteeSize, common.OFFSET)
		if err != nil {
			Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
			return NewBlockChainError(UnExpectedError, err)
		}
		Logger.log.Info("Swap: Out committee %+v", beaconSwapedCommittees)
		Logger.log.Info("Swap: In committee %+v", beaconNewCommittees)
	}
	return nil
}

//===================================Util for Beacon=============================
func GetStakingCandidate(beaconBlock BeaconBlock) ([]string, []string) {
	beacon := []string{}
	shard := []string{}
	beaconBlockBody := beaconBlock.Body
	for _, v := range beaconBlockBody.Instructions {
		if v[0] == "stake" && v[2] == "beacon" {
			beacon = strings.Split(v[1], ",")
		}
		if v[0] == "stake" && v[2] == "shard" {
			shard = strings.Split(v[1], ",")
		}
	}

	return beacon, shard
}

// Assumption:
// validator and candidate public key encode as base58 string
// assume that candidates are already been checked
// Check validation of candidate in transaction
func AssignValidator(candidates []string, rand int64, activeShards int) (map[byte][]string, error) {
	pendingValidators := make(map[byte][]string)
	for _, candidate := range candidates {
		shardID := calculateCandidateShardID(candidate, rand, activeShards)
		pendingValidators[shardID] = append(pendingValidators[shardID], candidate)
	}
	return pendingValidators, nil
}

// AssignValidatorShard, param for better convenice than AssignValidator
func AssignValidatorShard(currentCandidates map[byte][]string, shardCandidates []string, rand int64, activeShards int) error {
	for _, candidate := range shardCandidates {
		shardID := calculateCandidateShardID(candidate, rand, activeShards)
		currentCandidates[shardID] = append(currentCandidates[shardID], candidate)
	}
	return nil
}

func VerifyValidator(candidate string, rand int64, shardID byte, activeShards int) (bool, error) {
	res := calculateCandidateShardID(candidate, rand, activeShards)
	if shardID == res {
		return true, nil
	} else {
		return false, nil
	}
}

// Formula ShardID: LSB[hash(candidatePubKey+randomNumber)]
// Last byte of hash(candidatePubKey+randomNumber)
func calculateCandidateShardID(candidate string, rand int64, activeShards int) (shardID byte) {

	seed := candidate + strconv.Itoa(int(rand))
	hash := sha256.Sum256([]byte(seed))
	// fmt.Println("Candidate public key", candidate)
	// fmt.Println("Hash of candidate serialized pubkey and random number", hash)
	// fmt.Printf("\"%d\",\n", hash[len(hash)-1])
	// fmt.Println("Shard to be assign", hash[len(hash)-1])
	shardID = byte(int(hash[len(hash)-1]) % activeShards)
	return shardID
}

// consider these list as queue structure
// unqueue a number of validator out of currentValidators list
// enqueue a number of validator into currentValidators list <=> unqueue a number of validator out of pendingValidators list
// return value: #1 remaining pendingValidators, #2 new currentValidators #3 swapped out validator, #4 incoming validator #5 error
func SwapValidator(pendingValidators []string, currentValidators []string, maxCommittee int, offset int) ([]string, []string, []string, []string, error) {
	if maxCommittee < 0 || offset < 0 {
		panic("committee can't be zero")
	}
	if offset == 0 {
		return []string{}, pendingValidators, currentValidators, []string{}, errors.New("can't not swap 0 validator")
	}
	// if number of pending validator is less or equal than offset, set offset equal to number of pending validator
	if offset > len(pendingValidators) {
		offset = len(pendingValidators)
	}
	// if swap offset = 0 then do nothing
	if offset == 0 {
		return pendingValidators, currentValidators, []string{}, []string{}, errors.New("no pending validator for swapping")
	}
	if offset > maxCommittee {
		return pendingValidators, currentValidators, []string{}, []string{}, errors.New("trying to swap too many validator")
	}
	tempValidators := []string{}
	swapValidator := []string{}
	// if len(currentValidator) < maxCommittee then push validator until it is full
	if len(currentValidators) < maxCommittee {
		diff := maxCommittee - len(currentValidators)
		if diff >= offset {
			tempValidators = append(tempValidators, pendingValidators[:offset]...)
			currentValidators = append(currentValidators, tempValidators...)
			pendingValidators = pendingValidators[offset:]
			return pendingValidators, currentValidators, swapValidator, tempValidators, nil
		} else {
			offset -= diff
			tempValidators := append(tempValidators, pendingValidators[:diff]...)
			pendingValidators = pendingValidators[diff:]
			currentValidators = append(currentValidators, tempValidators...)
		}
	}
	fmt.Println("Swap Validator/Before: pendingValidators", pendingValidators)
	fmt.Println("Swap Validator/Before: currentValidators", currentValidators)
	fmt.Println("Swap Validator: offset", offset)
	// out pubkey: swapped out validator
	swapValidator = append(swapValidator, currentValidators[:offset]...)
	// unqueue validator with index from 0 to offset-1 from currentValidators list
	currentValidators = currentValidators[offset:]
	// in pubkey: unqueue validator with index from 0 to offset-1 from pendingValidators list
	tempValidators = append(tempValidators, pendingValidators[:offset]...)
	// enqueue new validator to the remaning of current validators list
	currentValidators = append(currentValidators, pendingValidators[:offset]...)
	// save new pending validators list
	pendingValidators = pendingValidators[offset:]
	fmt.Println("Swap Validator: pendingValidators", pendingValidators)
	fmt.Println("Swap Validator: currentValidators", currentValidators)
	fmt.Println("Swap Validator: swapValidator", swapValidator)
	fmt.Println("Swap Validator: tempValidators", tempValidators)
	if len(currentValidators) > maxCommittee {
		panic("Length of current validator greater than max committee in Swap validator ")
	}
	return pendingValidators, currentValidators, swapValidator, tempValidators, nil
}

// return: #param1: validator list after remove
// in parameter: #param1: list of full validator
// in parameter: #param2: list of removed validator
// removed validators list must be a subset of full validator list and it must be first in the list
func RemoveValidator(validators []string, removedValidators []string) ([]string, error) {
	// if number of pending validator is less or equal than offset, set offset equal to number of pending validator
	if len(removedValidators) > len(validators) {
		return validators, errors.New("trying to remove too many validators")
	}

	for index, validator := range removedValidators {
		if strings.Compare(validators[index], validator) == 0 {
			validators = validators[1:]
		} else {
			return validators, errors.New("remove Validator with Wrong Format")
		}
	}
	return validators, nil
}

/*
	Shuffle Candidate:
		Candidate Value Concatenate with Random Number
		Then Hash and Obtain Hash Value
		Sort Hash Value Then Re-arrange Candidate corresponding to Hash Value
*/
func ShuffleCandidate(candidates []string, rand int64) ([]string, error) {
	fmt.Println("Beacon Process/Shuffle Candidate: Candidate Before Sort ", candidates)
	hashes := []string{}
	m := make(map[string]string)
	sortedCandidate := []string{}
	for _, candidate := range candidates {
		seed := candidate + strconv.Itoa(int(rand))
		hash := sha256.Sum256([]byte(seed))
		hashes = append(hashes, string(hash[:32]))
		m[string(hash[:32])] = candidate
	}
	sort.Strings(hashes)
	for _, candidate := range m {
		sortedCandidate = append(sortedCandidate, candidate)
	}
	fmt.Println("Beacon Process/Shuffle Candidate: Candidate After Sort ", sortedCandidate)
	return sortedCandidate, nil
}

//=====================ARCHIVE=========================

/*
Insert new block into beaconchain
1. Verify Block
	1.1 Verify Block (block height, parent block,...)
	1.2 Validate Block after process (root hash, random number,...)
2. Update: Process block
	2.1 Process BestStateBeacon
	2.2 Store BestStateBeacon
3. Store Block
*/

// func (blockchain *BlockChain) ConnectBlockBeacon(block *BeaconBlock) error {
// 	blockchain.chainLock.Lock()
// 	defer blockchain.chainLock.Unlock()
// 	blockHash := block.Hash().String()

// 	Logger.log.Infof("Insert block %+v to Blockchain", blockHash)

// 	//===================Verify============================
// 	Logger.log.Infof("Verify Pre-Process block %+v to Blockchain", blockHash)

// 	err := blockchain.VerifyPreProcessingBeaconBlock(block)
// 	if err != nil {
// 		Logger.log.Error("Error update best state for block", block, "in beacon chain")
// 		return NewBlockChainError(UnExpectedError, err)
// 	}

// 	//===================Post-Verify == Validation============================
// 	Logger.log.Infof("Verify Post-Process block %+v to Blockchain", blockHash)
// 	err = blockchain.VerifyPostProcessingBeaconBlock(block)
// 	if err != nil {
// 		Logger.log.Error("Error Verify Post-Processing block", block, "in beacon chain")
// 		return NewBlockChainError(UnExpectedError, err)
// 	}

// 	//===================Process============================
// 	Logger.log.Infof("Process block %+v", blockHash)

// 	Logger.log.Infof("Process BeaconBestState block %+v", blockHash)
// 	// Process best state or not and store beststate
// 	err = blockchain.BestState.Beacon.Update(block)
// 	if err != nil {
// 		Logger.log.Error("Error update best state for block", block, "in beacon chain")
// 		return NewBlockChainError(UnExpectedError, err)
// 	}
// 	//===================Store Block and BestState in cache======================
// 	return nil
// }
// // Maybe accept new block (new block created from consensus)
// /*
// 	1. Verify Signature
// 	2. Verify Pre Processing
// 	3.
// 		- Load (from cache or database) beststate corressponding to block
// 		- Stored loaded beststate in local variable
// 		- verify loaded beststate with new block
// 	4. Update local beststate (process local beststate with new block)
// 	5. Verify Post Processing (verify updated local beststate with new block)
// 	6. Store in cache
// 		- updated local beststate
// 		- new block
// 	7. Acceptblock
// 		- Store in DB Previous Block of new Block
// 	    - Update final beststate (blockchain.BestState.BestStateBeacon.Beacon) with previous block
// 	    - Store just updated final beststate in DB
// */
// // This function return key to retrive new block and new beststate in cache
// func (blockchain *BlockChain) MaybeAcceptBeaconBlock(block *BeaconBlock) (string, error) {
// 	blockchain.chainLock.Lock()
// 	defer blockchain.chainLock.Unlock()
// 	Logger.log.Infof("Maybe accept new block %d, with hash %+v", block.Header.Height, *block.Hash())
// 	if err := blockchain.VerifyPreProcessingBeaconBlock(block); err != nil {
// 		return "", err
// 	}
// 	//========Verify block with previous best state
// 	// Get Beststate of previous block == previous best state
// 	// Clone best state value into new variable
// 	beaconBestState := BestStateBeacon{}
// 	// check with current final best state
// 	if strings.Compare(blockchain.BestState.Beacon.BestBlockHash.String(), block.Header.PrevBlockHash.String()) == 0 {
// 		tempMarshal, err := json.Marshal(blockchain.BestState.Beacon)
// 		if err != nil {
// 			return "", NewBlockChainError(UnmashallJsonBlockError, err)
// 		}
// 		json.Unmarshal(tempMarshal, &beaconBestState)
// 	} else {
// 		// check with current cache best state
// 		var err error
// 		beaconBestState, err = blockchain.GetMaybeAcceptBeaconBestState(block.Header.PrevBlockHash.String())
// 		if err != nil {
// 			return "", err
// 		}
// 	}
// 	// if no match best state found then block is unknown
// 	if reflect.DeepEqual(beaconBestState, BestStateBeacon{}) {
// 		return "", NewBlockChainError(BeaconError, errors.New("Beacon Block does not match with any Beacon State in cache or in Database"))
// 	}

// 	// beaconBestState.lock.Lock()
// 	// defer beaconBestState.lock.Unlock()

// 	// Verify block with previous best state
// 	if err := beaconBestState.VerifyBestStateWithBeaconBlock(block, true); err != nil {
// 		return "", err
// 	}

// 	//========Update best state with new block
// 	if err := beaconBestState.Update(block); err != nil {
// 		return "", err
// 	}
// 	//========Post verififcation: verify new beaconstate with corresponding block
// 	if err := beaconBestState.VerifyPostProcessingBeaconBlock(block); err != nil {
// 		return "", err
// 	}

// 	//========Store new Beaconblock and new Beacon bestState in cache
// 	_, err := blockchain.StoreMaybeAcceptBeaconBeststate(beaconBestState)
// 	if err != nil {
// 		return "", err
// 	}
// 	keyBL, err := blockchain.StoreMaybeAcceptBeaconBlock(*block)
// 	if err != nil {
// 		return "", err
// 	}
// 	//=========Remove beacon block
// 	blockchain.config.ShardToBeaconPool.RemovePendingBlock(beaconBestState.BestShardHeight)
// 	//=========Accept previous if new block is valid
// 	if err := blockchain.AcceptBeaconBlock(&block.Header.PrevBlockHash); err != nil {
// 		return "", err
// 	}
// 	Logger.log.Infof("New maybe accepted VALID block %d, with hash %x", block.Header.Height, *block.Hash())
// 	return keyBL, nil
// }

// //Store block & state offcial
// //lock sync.Mutex blockchain before call accept beacon block
// func (blockchain *BlockChain) AcceptBeaconBlock(blockHash *common.Hash) error {
// 	// blockchain.chainLock.Lock()
// 	// defer blockchain.chainLock.Unlock()
// 	// This function make sure if stored block at height 91, then best state height at 90
// 	beaconBlock, err := blockchain.GetMaybeAcceptBeaconBlock(blockHash.String())
// 	if err != nil {
// 		Logger.log.Errorf("Can't find block %+v to accept", blockHash)
// 		return err
// 	}
// 	Logger.log.Infof("Accept block %d, with hash %+v", beaconBlock.Header.Height, blockHash)
// 	err = blockchain.BestState.Beacon.Update(&beaconBlock)
// 	if err != nil {
// 		return err
// 	}

// 	// beaconBestState, err := blockchain.GetMaybeAcceptBeaconBestState(blockHash.String())
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	// if !reflect.DeepEqual(beaconBestState, blockchain.BestState.Beacon) {
// 	// 	Logger.log.Error("Current best state and stored block %+v are not compatible", blockHash)
// 	// 	return NewBlockChainError(BeaconError, errors.New("Current best state and stored block are not compatible"))
// 	// }
// 	//===================Store Block============================
// 	Logger.log.Infof("Store Beacon block %+v", blockHash)
// 	if err := blockchain.config.DataBase.StoreBeaconBlock(beaconBlock); err != nil {
// 		Logger.log.Error("Error store beacon block", blockHash, "in beacon chain")
// 		return err
// 	}

// 	//===================Store State============================
// 	Logger.log.Infof("Store BeaconBestState block %+v", blockHash)
// 	//Process stored block with current best state

// 	if err := blockchain.config.DataBase.StoreBeaconBestState(blockchain.BestState.Beacon); err != nil {
// 		Logger.log.Error("Error Store best state for block", blockHash, "in beacon chain")
// 		return NewBlockChainError(UnExpectedError, err)
// 	}
// 	Logger.log.Infof("Accepted block %+v", blockHash)
// 	return nil
// }
