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

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/common/base58"
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
	if err := beaconBestState.Update(block, blockchain); err != nil {
		return err
	}
	//========Post verififcation: verify new beaconstate with corresponding block
	if err := beaconBestState.VerifyPostProcessingBeaconBlock(block, snapShotBeaconCommittee); err != nil {
		return err
	}
	Logger.log.Infof("Block %d, with hash %+v is VALID for signing", block.Header.Height, *block.Hash())
	return nil
}

func (blockchain *BlockChain) InsertBeaconBlock(block *BeaconBlock, isValidated bool) error {
	blockchain.chainLock.Lock()
	defer blockchain.chainLock.Unlock()

	Logger.log.Infof("Check block existence for insert process %d, with hash %+v", block.Header.Height, *block.Hash())
	isExist, _ := blockchain.config.DataBase.HasBeaconBlock(block.Hash())
	if isExist {
		return NewBlockChainError(DuplicateBlockErr, errors.New("This block has been stored already"))
	}
	Logger.log.Infof("Begin Insert new block %d, with hash %+v \n", block.Header.Height, *block.Hash())
	if !isValidated {
		Logger.log.Infof("Verify Pre Processing Beacon Block %+v \n", *block.Hash())
		if err := blockchain.VerifyPreProcessingBeaconBlock(block, false); err != nil {
			return err
		}
	} else {
		Logger.log.Infof("BEACON %+v | SKIP Verify Pre Processing Block %+v \n", *block.Hash())
	}
	//========Verify block with previous best state
	// check with current final best state
	// block can only be insert if it match the current best state
	if strings.Compare(blockchain.BestState.Beacon.BestBlockHash.String(), block.Header.PrevBlockHash.String()) != 0 {
		return NewBlockChainError(BeaconError, errors.New("beacon Block does not match with any Beacon State in cache or in Database"))
	}
	// fmt.Printf("BeaconBest state %+v \n", blockchain.BestState.Beacon)
	if !isValidated {
		Logger.log.Infof("Verify BestState with Beacon Block %+v \n", *block.Hash())
		// Verify block with previous best state
		if err := blockchain.BestState.Beacon.VerifyBestStateWithBeaconBlock(block, true); err != nil {
			return err
		}
	} else {
		Logger.log.Infof("BEACON %+v | SKIP Verify BestState with Block %+v \n", *block.Hash())
	}
	Logger.log.Infof("Update BestState with Beacon Block %+v \n", *block.Hash())
	//========Update best state with new block
	snapShotBeaconCommittee := blockchain.BestState.Beacon.BeaconCommittee
	if err := blockchain.BestState.Beacon.Update(block, blockchain); err != nil {
		return err
	}
	if !isValidated {
		Logger.log.Infof("Verify Post Processing Beacon Block %+v \n", *block.Hash())
		//========Post verififcation: verify new beaconstate with corresponding block
		if err := blockchain.BestState.Beacon.VerifyPostProcessingBeaconBlock(block, snapShotBeaconCommittee); err != nil {
			return err
		}
	} else {
		Logger.log.Infof("BEACON %+v | SKIP Verify Post Processing Block %+v \n", *block.Hash())
	}

	for shardID, shardStates := range block.Body.ShardState {
		for _, shardState := range shardStates {
			blockchain.config.DataBase.StoreAcceptedShardToBeacon(shardID, block.Header.Height, &shardState.Hash)
		}
	}
	// if committee of this epoch isn't store yet then store it
	// @NOTICE: Change to height
	Logger.log.Infof("Store Committee in Height %+v \n", block.Header.Height)
	// res, err := blockchain.config.DataBase.HasCommitteeByEpoch(block.Header.Epoch)
	// if res == false {
	if err := blockchain.config.DataBase.StoreCommitteeByEpoch(block.Header.Height, blockchain.BestState.Beacon.ShardCommittee); err != nil {
		return err
	}
	// }
	shardCommitteeByte, err := blockchain.config.DataBase.FetchCommitteeByEpoch(block.Header.Epoch)
	if err != nil {
		fmt.Println("No committee for this epoch")
	}
	shardCommittee := make(map[byte][]string)
	if err := json.Unmarshal(shardCommitteeByte, &shardCommittee); err != nil {
		fmt.Println("Fail to unmarshal shard committee")
	}
	// fmt.Println("Beacon Process/Shard Committee in Epoch ", block.Header.Epoch, shardCommittee)
	//=========Store cross shard state ==================================
	lastCrossShardState := GetBestStateBeacon().LastCrossShardState
	GetBestStateBeacon().lockMu.Lock()
	if block.Body.ShardState != nil {
		for fromShard, shardBlocks := range block.Body.ShardState {

			go func(fromShard byte, shardBlocks []ShardState) {
				for _, shardBlock := range shardBlocks {
					for _, toShard := range shardBlock.CrossShard {
						if fromShard == toShard {
							continue
						}
						if lastCrossShardState[fromShard] == nil {
							lastCrossShardState[fromShard] = make(map[byte]uint64)
						}
						lastHeight := lastCrossShardState[fromShard][toShard] // get last cross shard height from shardID  to crossShardShardID
						waitHeight := shardBlock.Height
						// fmt.Println("StoreCrossShardNextHeight", fromShard, toShard, lastHeight, waitHeight)
						blockchain.config.DataBase.StoreCrossShardNextHeight(fromShard, toShard, lastHeight, waitHeight)
						//beacon process shard_to_beacon in order so cross shard next height also will be saved in order
						//dont care overwrite this value
						blockchain.config.DataBase.StoreCrossShardNextHeight(fromShard, toShard, waitHeight, 0)
						if lastCrossShardState[fromShard] == nil {
							lastCrossShardState[fromShard] = make(map[byte]uint64)
						}
						lastCrossShardState[fromShard][toShard] = waitHeight //update lastHeight to waitHeight
					}
				}
				blockchain.config.CrossShardPool[fromShard].UpdatePool()
			}(fromShard, shardBlocks)
		}
	}
	GetBestStateBeacon().lockMu.Unlock()
	// Process instructions and store stability data
	if err := blockchain.updateStabilityLocalState(block); err != nil {
		return err
	}
	// ************ Store block at last
	Logger.log.Info("Store StabilityInfo ")
	if err := blockchain.config.DataBase.StoreStabilityInfoByHeight(block.Header.Height, bestStateBeacon.StabilityInfo); err != nil {
		return err
	}
	//========Store new Beaconblock and new Beacon bestState in cache
	Logger.log.Infof("Store Beacon BestState  ")
	if err := blockchain.config.DataBase.StoreBeaconBestState(blockchain.BestState.Beacon); err != nil {
		return err
	}

	Logger.log.Info("Store Beacon Block ", block.Header.Height, *block.Hash())
	if err := blockchain.config.DataBase.StoreBeaconBlock(block); err != nil {
		return err
	}
	blockHash := block.Hash()
	if err := blockchain.config.DataBase.StoreBeaconBlockIndex(blockHash, block.Header.Height); err != nil {
		return err
	}

	//=========Remove beacon block in pool
	blockchain.config.BeaconPool.SetBeaconState(blockchain.BestState.Beacon.BeaconHeight)

	//=========Remove shard to beacon block in pool
	Logger.log.Info("Remove block from pool block with hash  ", *block.Hash(), block.Header.Height, blockchain.BestState.Beacon.BestShardHeight)
	blockchain.config.ShardToBeaconPool.SetShardState(blockchain.BestState.Beacon.GetBestShardHeight())

	Logger.log.Info("Finish Insert new block , with hash", block.Header.Height, *block.Hash())
	if block.Header.Height%50 == 0 {
		fmt.Printf("[db] inserted beacon height: %d\n", block.Header.Height)
	}
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
	//verify producer sig
	blkHash := block.Header.Hash()
	producerPk := base58.Base58Check{}.Encode(block.Header.ProducerAddress.Pk, common.ZeroByte)
	err := cashec.ValidateDataB58(producerPk, block.ProducerSig, blkHash.GetBytes())
	if err != nil {
		return NewBlockChainError(ProducerError, errors.New("Producer's sig not match"))
	}
	//verify producer
	producerPosition := (blockchain.BestState.Beacon.BeaconProposerIdx + block.Header.Round) % len(blockchain.BestState.Beacon.BeaconCommittee)
	tempProducer := blockchain.BestState.Beacon.BeaconCommittee[producerPosition]
	if strings.Compare(tempProducer, producerPk) != 0 {
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
	//jackalope: temporary commment for debug purpose
	//if block.Header.Timestamp <= parentBlockInterface.Header.Timestamp {
	//	return NewBlockChainError(TimestampError, errors.New("timestamp of new block can't equal to parent block"))
	//}

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
		// @UNCOMMENT TO TEST
		beaconBestState := BestStateBeacon{}
		tempShardStates := make(map[byte][]ShardState)
		validStakers := [][]string{}
		validSwappers := make(map[byte][][]string)
		stabilityInstructions := [][]string{}
		tempMarshal, err := json.Marshal(*blockchain.BestState.Beacon)
		err = json.Unmarshal(tempMarshal, &beaconBestState)
		if err != nil {
			return NewBlockChainError(UnExpectedError, errors.New("Fail to Unmarshal beacon beststate"))
		}
		beaconBestState.CandidateShardWaitingForCurrentRandom = blockchain.BestState.Beacon.CandidateShardWaitingForCurrentRandom
		beaconBestState.CandidateShardWaitingForNextRandom = blockchain.BestState.Beacon.CandidateShardWaitingForNextRandom
		beaconBestState.CandidateBeaconWaitingForCurrentRandom = blockchain.BestState.Beacon.CandidateBeaconWaitingForCurrentRandom
		beaconBestState.CandidateBeaconWaitingForNextRandom = blockchain.BestState.Beacon.CandidateBeaconWaitingForNextRandom
		if reflect.DeepEqual(beaconBestState, BestStateBeacon{}) {
			panic(NewBlockChainError(BeaconError, errors.New("problem with beststate in producing new block")))
		}
		accumulativeValues := &accumulativeValues{
			saleDataMap: map[string]*component.SaleData{},
		}
		allShardBlocks := blockchain.config.ShardToBeaconPool.GetValidPendingBlock(nil)
		var keys []int
		for k := range allShardBlocks {
			keys = append(keys, int(k))
		}
		sort.Ints(keys)
		for _, value := range keys {
			shardID := byte(value)
			shardBlocks := allShardBlocks[shardID]
			if len(shardBlocks) >= len(block.Body.ShardState[shardID]) {
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
				for _, shardBlock := range shardBlocks {
					tempShardState, validStaker, validSwapper, stabilityInstruction := blockchain.GetShardStateFromBlock(&beaconBestState, shardBlock, accumulativeValues, shardID)
					tempShardStates[shardID] = append(tempShardStates[shardID], tempShardState[shardID])
					validStakers = append(validStakers, validStaker...)
					validSwappers[shardID] = append(validSwappers[shardID], validSwapper[shardID]...)
					stabilityInstructions = append(stabilityInstructions, stabilityInstruction...)
				}
			} else {
				return NewBlockChainError(ShardStateError, errors.New("shardstate fail to verify with ShardToBeacon Block in pool"))
			}
		}
		votingInstructionDCB, err := blockchain.generateVotingInstructionWOIns(DCBConstitutionHelper{})
		if err != nil {
			fmt.Println("[ndh]-Build DCB voting instruction failed: ", err)
		} else {
			if len(votingInstructionDCB) != 0 {
				stabilityInstructions = append(stabilityInstructions, votingInstructionDCB...)
			}
		}
		votingInstructionGOV, err := blockchain.generateVotingInstructionWOIns(GOVConstitutionHelper{})
		if err != nil {
			fmt.Println("[ndh]-Build GOV voting instruction failed: ", err)
		} else {
			if len(votingInstructionGOV) != 0 {
				stabilityInstructions = append(stabilityInstructions, votingInstructionGOV...)
			}
		}
		oracleInsts, err := blockchain.buildOracleRewardInstructions(&beaconBestState)
		if err != nil {
			fmt.Println("Build oracle reward instructions failed: ", err)
		} else if len(oracleInsts) > 0 {
			stabilityInstructions = append(stabilityInstructions, oracleInsts...)
		}
		tempInstruction := beaconBestState.GenerateInstruction(block, validStakers, validSwappers, beaconBestState.CandidateShardWaitingForCurrentRandom, stabilityInstructions)
		fmt.Println("BeaconProcess/tempInstruction: ", tempInstruction)
		tempInstructionArr := []string{}
		for _, strs := range tempInstruction {
			tempInstructionArr = append(tempInstructionArr, strs...)
		}
		tempInstructionHash, err := GenerateHashFromStringArray(tempInstructionArr)
		if err != nil {
			return NewBlockChainError(HashError, errors.New("Fail to generate hash for instruction"))
		}
		fmt.Println("BeaconProcess/tempInstructionHash: ", tempInstructionHash)
		fmt.Println("BeaconProcess/block.Header.InstructionHash: ", block.Header.InstructionHash)
		if strings.Compare(tempInstructionHash.String(), block.Header.InstructionHash.String()) != 0 {
			return NewBlockChainError(InstructionHashError, errors.New("instruction hash is not correct"))
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
		if len(bestStateBeacon.BeaconCommittee) > 3 && len(block.ValidatorsIdx[1]) < (len(bestStateBeacon.BeaconCommittee)>>1) {
			return NewBlockChainError(SignatureError, errors.New("block validators and Beacon committee is not compatible "+fmt.Sprint(len(block.ValidatorsIdx))))
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
	// for shardID, shardStates := range block.Body.ShardState {
	// 	// Do not check this condition with first minted block (genesis block height = 1)
	// 	if bestStateBeacon.BeaconHeight != 2 {
	// fmt.Printf("Beacon Process/Check ShardStates with BestState Current Shard Height %+v \n", bestStateBeacon.AllShardState[shardID][len(bestStateBeacon.AllShardState[shardID])-1].Height)
	// fmt.Printf("Beacon Process/Check ShardStates with BestState FirstShardHeight %+v \n", shardStates[0].Height)
	// if shardStates[0].Height-bestStateBeacon.AllShardState[shardID][len(bestStateBeacon.AllShardState[shardID])-1].Height != 1 {
	// 	return NewBlockChainError(ShardStateError, errors.New("Shardstates are not compatible with beacon best state"))
	// }
	// }
	// }
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
func (bestStateBeacon *BestStateBeacon) Update(newBlock *BeaconBlock, chain *BlockChain) error {
	bestStateBeacon.lockMu.Lock()
	defer bestStateBeacon.lockMu.Unlock()

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
	if newBlock.Header.Height == 1 {
		bestStateBeacon.BeaconProposerIdx = 0
	} else {
		bestStateBeacon.BeaconProposerIdx = common.IndexOfStr(base58.Base58Check{}.Encode(newBlock.Header.ProducerAddress.Pk, common.ZeroByte), bestStateBeacon.BeaconCommittee)
	}

	allShardState := newBlock.Body.ShardState
	// if bestStateBeacon.AllShardState == nil {
	// 	bestStateBeacon.AllShardState = make(map[byte][]ShardState)
	// 	for index := 0; index < common.MAX_SHARD_NUMBER; index++ {
	// 		bestStateBeacon.AllShardState[byte(index)] = []ShardState{
	// 			ShardState{
	// 				Height: 1,
	// 			},
	// 		}
	// 	}
	// }
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
		//if _, ok := bestStateBeacon.AllShardState[shardID]; !ok {
		//	bestStateBeacon.AllShardState[shardID] = []ShardState{}
		//}
		//bestStateBeacon.AllShardState[shardID] = append(bestStateBeacon.AllShardState[shardID], shardStates...)
	}

	//cross shard state

	// update param
	err := bestStateBeacon.updateOracleParams(chain)
	if err != nil {
		Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
		return NewBlockChainError(UnExpectedError, err)
	}
	instructions := newBlock.Body.Instructions
	for _, l := range instructions {
		if len(l) < 1 {
			continue
		}
		// For stability instructions
		err := bestStateBeacon.processStabilityInstruction(l, chain.config.DataBase)
		if err != nil {
			Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
			return NewBlockChainError(UnExpectedError, err)
		}

		if l[0] == SetAction {
			bestStateBeacon.Params[l[1]] = l[2]
		}
		if l[0] == DeleteAction {
			delete(bestStateBeacon.Params, l[1])
		}
		if l[0] == SwapAction {
			fmt.Println("SWAP", l)
			// format
			// ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "shard" "shardID"]
			// ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "beacon"]
			inPubkeys := strings.Split(l[1], ",")
			outPubkeys := strings.Split(l[2], ",")
			fmt.Println("SWAP l1", l[1])
			fmt.Println("SWAP l2", l[2])
			fmt.Println("SWAP inPubkeys", inPubkeys)
			fmt.Println("SWAP outPubkeys", outPubkeys)
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
		if l[0] == RandomAction {
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
		if l[0] == StakeAction && l[2] == "beacon" {
			beacon := strings.Split(l[1], ",")
			newBeaconCandidate = append(newBeaconCandidate, beacon...)
		}
		if l[0] == StakeAction && l[2] == "shard" {
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
			fmt.Println(bestStateBeacon.ShardCommittee[byte(shardID)])
		}
		bestStateBeacon.Epoch = 1
	} else {
		bestStateBeacon.CandidateBeaconWaitingForNextRandom = append(bestStateBeacon.CandidateBeaconWaitingForNextRandom, newBeaconCandidate...)
		bestStateBeacon.CandidateShardWaitingForNextRandom = append(bestStateBeacon.CandidateShardWaitingForNextRandom, newShardCandidate...)
		// fmt.Println("Beacon Process/Before: CandidateShardWaitingForNextRandom: ", bestStateBeacon.CandidateShardWaitingForNextRandom)
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
			Logger.log.Critical("==================Beacon Process: Snapshot candidate====================")
			Logger.log.Critical("Beacon Process: CandidateShardWaitingForCurrentRandom: ", bestStateBeacon.CandidateShardWaitingForCurrentRandom)
			Logger.log.Critical("Beacon Process: CandidateBeaconWaitingForCurrentRandom: ", bestStateBeacon.CandidateBeaconWaitingForCurrentRandom)
			// reset candidate list
			bestStateBeacon.CandidateShardWaitingForNextRandom = []string{}
			bestStateBeacon.CandidateBeaconWaitingForNextRandom = []string{}
			Logger.log.Critical("Beacon Process/After: CandidateShardWaitingForNextRandom: ", bestStateBeacon.CandidateShardWaitingForNextRandom)
			Logger.log.Critical("Beacon Process/After: CandidateBeaconWaitingForCurrentRandom: ", bestStateBeacon.CandidateBeaconWaitingForCurrentRandom)
			// assign random timestamp
			bestStateBeacon.CurrentRandomTimeStamp = newBlock.Header.Timestamp
		}
		// if get new random number
		// Assign candidate to shard
		// assign CandidateShardWaitingForCurrentRandom to ShardPendingValidator with CurrentRandom
		if randomFlag {
			bestStateBeacon.IsGetRandomNumber = true
			//fmt.Println("Beacon Process/Update/RandomFlag: Shard Candidate Waiting for Current Random Number", bestStateBeacon.CandidateShardWaitingForCurrentRandom)
			//Logger.log.Critical("bestStateBeacon.ShardPendingValidator", bestStateBeacon.ShardPendingValidator)
			//Logger.log.Critical("bestStateBeacon.CandidateShardWaitingForCurrentRandom", bestStateBeacon.CandidateShardWaitingForCurrentRandom)
			//Logger.log.Critical("bestStateBeacon.CurrentRandomNumber", bestStateBeacon.CurrentRandomNumber)
			//Logger.log.Critical("bestStateBeacon.ActiveShards", bestStateBeacon.ActiveShards)
			err := AssignValidatorShard(bestStateBeacon.ShardPendingValidator, bestStateBeacon.CandidateShardWaitingForCurrentRandom, bestStateBeacon.CurrentRandomNumber, bestStateBeacon.ActiveShards)
			if err != nil {
				Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
				return NewBlockChainError(UnExpectedError, err)
			}
			// delete CandidateShardWaitingForCurrentRandom list
			bestStateBeacon.CandidateShardWaitingForCurrentRandom = []string{}
			//fmt.Println("Beacon Process/Update/RandomFalg: Shard Pending Validator", bestStateBeacon.ShardPendingValidator)
			// Shuffle candidate
			// shuffle CandidateBeaconWaitingForCurrentRandom with current random number
			//fmt.Println("Beacon Process/Update/RandomFlag: Beacon Candidate Waiting for Current Random Number", bestStateBeacon.CandidateBeaconWaitingForCurrentRandom)
			newBeaconPendingValidator, err := ShuffleCandidate(bestStateBeacon.CandidateBeaconWaitingForCurrentRandom, bestStateBeacon.CurrentRandomNumber)
			//fmt.Println("Beacon Process/Update/RandomFalg: NewBeaconPendingValidator", newBeaconPendingValidator)
			if err != nil {
				Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
				return NewBlockChainError(UnExpectedError, err)
			}
			bestStateBeacon.CandidateBeaconWaitingForCurrentRandom = []string{}
			bestStateBeacon.BeaconPendingValidator = append(bestStateBeacon.BeaconPendingValidator, newBeaconPendingValidator...)
			//fmt.Println("Beacon Process/Update/RandomFalg: Beacon Pending Validator", bestStateBeacon.BeaconPendingValidator)
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
		if len(v) < 1 {
			continue
		}
		if v[0] == StakeAction && v[2] == "beacon" {
			beacon = strings.Split(v[1], ",")
		}
		if v[0] == StakeAction && v[2] == "shard" {
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
	hash := common.HashB([]byte(seed))
	// fmt.Println("Candidate public key", candidate)
	// fmt.Println("Hash of candidate serialized pubkey and random number", hash)
	// fmt.Printf("\"%d\",\n", hash[len(hash)-1])
	// fmt.Println("Shard to be assign", hash[len(hash)-1])
	shardID = byte(int(hash[len(hash)-1]) % activeShards)
	Logger.log.Critical("calculateCandidateShardID/shardID", shardID)
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
		hash := common.HashB([]byte(seed))
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
