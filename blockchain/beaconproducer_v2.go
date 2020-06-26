package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/metadata"
	"sort"
	"strings"
	"time"
)

// GetShardState get Shard To Beacon Block
// Rule:
// 1. Shard To Beacon Blocks will be get from Shard To Beacon Pool (only valid block)
// 2. Process shards independently, for each shard:
//	a. Shard To Beacon Block List must be compatible with current shard state in beacon best state:
//  + Increased continuosly in height (10, 11, 12,...)
//	  Ex: Shard state in beacon best state has height 11 then shard to beacon block list must have first block in list with height 12
//  + Shard To Beacon Block List must have incremental height in list (10, 11, 12,... NOT 10, 12,...)
//  + Shard To Beacon Block List can be verify with and only with current shard committee in beacon best state
//  + DO NOT accept Shard To Beacon Block List that can have two arbitrary blocks that can be verify with two different committee set
//  + If in Shard To Beacon Block List have one block with Swap Instruction, then this block must be the last block in this list (or only block in this list)
// return param:
// 1. shard state
// 2. valid stake instruction
// 3. valid swap instruction
// 4. bridge instructions
// 5. accepted reward instructions
// 6. stop auto staking instructions
func (blockchain *BlockChain) GetShardStateV2(
	beaconBestState *BeaconBestState,
	rewardForCustodianByEpoch map[common.Hash]uint64,
	portalParams PortalParams,
) (map[byte][]ShardState, [][]string, [][]string, [][]string, [][]string, [][]string) {
	shardStates := make(map[byte][]ShardState)
	stakeInstructions := [][]string{}
	validStakePublicKeys := []string{}
	stopAutoStakeInstructions := [][]string{}
	swapInstructions := [][]string{}
	//Get shard to beacon block from pool
	allShardBlocks := blockchain.GetShardBlockForBeaconProducer(beaconBestState.BestShardHeight)
	keys := []int{}
	for shardID, shardBlocks := range allShardBlocks {
		strs := fmt.Sprintf("GetShardState shardID: %+v, Height", shardID)
		for _, shardBlock := range shardBlocks {
			strs += fmt.Sprintf(" %d", shardBlock.Header.Height)
		}
		Logger.log.Info(strs)
		keys = append(keys, int(shardID))
	}
	sort.Ints(keys)
	//Shard block is a map ShardId -> array of shard block
	bridgeInstructions := [][]string{}
	acceptedRewardInstructions := [][]string{}
	statefulActionsByShardID := map[byte][][]string{}
	for _, v := range keys {
		shardID := byte(v)
		shardBlocks := allShardBlocks[shardID]
		for _, shardBlock := range shardBlocks {
			shardState, tempStakeInstructions, tempSwapInstructions, tempBridgeInstructions, tempAcceptedRewardInstructions, tempStopAutoStakeInstruction, statefulActions := blockchain.GetShardStateFromBlockV2(beaconBestState, beaconBestState.BeaconHeight+1, shardBlock, shardID, true, validStakePublicKeys)
			shardStates[shardID] = append(shardStates[shardID], shardState[shardID])
			stakeInstructions = append(stakeInstructions, tempStakeInstructions...)
			swapInstructions = append(swapInstructions, tempSwapInstructions...)
			bridgeInstructions = append(bridgeInstructions, tempBridgeInstructions...)
			acceptedRewardInstructions = append(acceptedRewardInstructions, tempAcceptedRewardInstructions)
			stopAutoStakeInstructions = append(stopAutoStakeInstructions, tempStopAutoStakeInstruction...)
			// group stateful actions by shardID
			_, found := statefulActionsByShardID[shardID]
			if !found {
				statefulActionsByShardID[shardID] = statefulActions
			} else {
				statefulActionsByShardID[shardID] = append(statefulActionsByShardID[shardID], statefulActions...)
			}
		}
	}
	// build stateful instructions
	statefulInsts := blockchain.buildStatefulInstructions(beaconBestState.featureStateDB, statefulActionsByShardID, beaconBestState.BeaconHeight+1, rewardForCustodianByEpoch, portalParams)
	bridgeInstructions = append(bridgeInstructions, statefulInsts...)
	return shardStates, stakeInstructions, swapInstructions, bridgeInstructions, acceptedRewardInstructions, stopAutoStakeInstructions
}

func (blockchain *BlockChain) GetShardStateFromBlockV2(
	curView *BeaconBestState,
	newBeaconHeight uint64,
	shardBlock *ShardBlock,
	shardID byte, isProducer bool,
	validStakePublicKeys []string,
) (map[byte]ShardState, [][]string, [][]string, [][]string, []string, [][]string, [][]string) {
	//Variable Declaration
	shardStates := make(map[byte]ShardState)
	stakeInstructions := [][]string{}
	swapInstructions := [][]string{}
	stopAutoStakeInstructions := [][]string{}
	bridgeInstructions := [][]string{}
	acceptedBlockRewardInfo := metadata.NewAcceptedBlockRewardInfo(shardID, shardBlock.Header.TotalTxsFee, shardBlock.Header.Height)
	acceptedRewardInstructions, err := acceptedBlockRewardInfo.GetStringFormat()
	if err != nil {
		// if err then ignore accepted reward instruction
		acceptedRewardInstructions = []string{}
	}
	//Get Shard State from Block
	shardState := ShardState{}
	shardState.CrossShard = make([]byte, len(shardBlock.Header.CrossShardBitMap))
	copy(shardState.CrossShard, shardBlock.Header.CrossShardBitMap)
	shardState.Hash = shardBlock.Header.Hash()
	shardState.Height = shardBlock.Header.Height
	shardStates[shardID] = shardState
	instructions, err := CreateShardInstructionsFromTransactionAndInstruction(shardBlock.Body.Transactions, blockchain, shardBlock.Header.ShardID)
	instructions = append(instructions, shardBlock.Body.Instructions...)

	// extract instructions
	for _, inst := range instructions {
		if len(inst) == 0 {
			continue
		}
		switch inst[0] {
		case instruction.SWAP_ACTION:
			swapInstructions = append(swapInstructions, inst)
		case instruction.STAKE_ACTION:
			stakeInstructions = append(stakeInstructions, inst)
		case instruction.STOP_AUTO_STAKE_ACTION:
			stopAutoStakeInstructions = append(stopAutoStakeInstructions, inst)
		}
	}
	bridgeInstructionForBlock, err := blockchain.buildBridgeInstructions(
		curView.GetBeaconFeatureStateDB(),
		shardID,
		instructions,
		newBeaconHeight,
	)
	if err != nil {
		BLogger.log.Errorf("Build bridge instructions failed: %s", err.Error())
	}
	// Pick instruction with shard committee's pubkeys to save to beacon block
	confirmInsts := pickBridgeSwapConfirmInst(instructions)
	if len(confirmInsts) > 0 {
		bridgeInstructionForBlock = append(bridgeInstructionForBlock, confirmInsts...)
		BLogger.log.Infof("Beacon block %d found bridge swap confirm inst in shard block %d: %s", newBeaconHeight, shardBlock.Header.Height, confirmInsts)
	}
	bridgeInstructions = append(bridgeInstructions, bridgeInstructionForBlock...)
	// Collect stateful actions
	statefulActions := blockchain.collectStatefulActions(instructions)
	Logger.log.Infof("Becon Produce: Got Shard Block %+v Shard %+v \n", shardBlock.Header.Height, shardID)
	return shardStates, stakeInstructions, swapInstructions, bridgeInstructions, acceptedRewardInstructions, stopAutoStakeInstructions, statefulActions
}

//  GenerateInstruction generate instruction for new beacon block
func (beaconBestState *BeaconBestState) GenerateInstructionV2(
	state BeaconCommitteeState,
	newBeaconHeight uint64,
	stakeInstructions [][]string,
	swapInstructions map[byte][][]string,
	stopAutoStakingInstructions [][]string,
	shardCandidates []incognitokey.CommitteePublicKey,
	bridgeInstructions [][]string,
	acceptedRewardInstructions [][]string,
	chainParamEpoch uint64,
	randomTime uint64,
	blockchain *BlockChain,
) ([][]string, error) {
	instructions := [][]string{}
	instructions = append(instructions, bridgeInstructions...)
	instructions = append(instructions, acceptedRewardInstructions...)
	//=======Swap
	// Shard Swap: both abnormal or normal swap
	var keys []int
	for k := range swapInstructions {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		instructions = append(instructions, swapInstructions[byte(shardID)]...)
	}
	// Beacon normal swap
	if newBeaconHeight%chainParamEpoch == 0 {
		swapBeaconInstructions := []string{}
		beaconPendingValidatorStr, err := incognitokey.CommitteeKeyListToString(beaconBestState.BeaconPendingValidator)
		if err != nil {
			return [][]string{}, err
		}
		beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(beaconBestState.BeaconCommittee)
		if err != nil {
			return [][]string{}, err
		}
		beaconSlashRootHash, err := blockchain.GetBeaconSlashRootHash(blockchain.GetBeaconChainDatabase(), newBeaconHeight-1)
		if err != nil {
			return [][]string{}, err
		}
		beaconSlashStateDB, err := statedb.NewWithPrefixTrie(beaconSlashRootHash, statedb.NewDatabaseAccessWarper(blockchain.GetBeaconChainDatabase()))
		producersBlackList, err := blockchain.getUpdatedProducersBlackList(beaconSlashStateDB, true, -1, beaconCommitteeStr, newBeaconHeight-1)
		if err != nil {
			Logger.log.Error(err)
		}
		badProducersWithPunishment := blockchain.buildBadProducersWithPunishment(true, -1, beaconCommitteeStr)
		badProducersWithPunishmentBytes, err := json.Marshal(badProducersWithPunishment)
		if err != nil {
			Logger.log.Error(err)
		}
		if common.IndexOfUint64(newBeaconHeight/chainParamEpoch, blockchain.config.ChainParams.EpochBreakPointSwapNewKey) > -1 {
			epoch := newBeaconHeight / chainParamEpoch
			swapBeaconInstructions, _, beaconCommittee := CreateBeaconSwapActionForKeyListV2(blockchain.config.GenesisParams, beaconPendingValidatorStr, beaconCommitteeStr, beaconBestState.MinBeaconCommitteeSize, epoch)
			instructions = append(instructions, swapBeaconInstructions)
			beaconRootInst, _ := buildBeaconSwapConfirmInstruction(beaconCommittee, newBeaconHeight)
			instructions = append(instructions, beaconRootInst)
		} else {
			_, currentValidators, swappedValidator, beaconNextCommittee, err := SwapValidator(beaconPendingValidatorStr, beaconCommitteeStr, beaconBestState.MaxBeaconCommitteeSize, beaconBestState.MinBeaconCommitteeSize, blockchain.config.ChainParams.Offset, producersBlackList, blockchain.config.ChainParams.SwapOffset)
			if len(swappedValidator) > 0 || len(beaconNextCommittee) > 0 && err == nil {
				swapBeaconInstructions = append(swapBeaconInstructions, "swap")
				swapBeaconInstructions = append(swapBeaconInstructions, strings.Join(beaconNextCommittee, ","))
				swapBeaconInstructions = append(swapBeaconInstructions, strings.Join(swappedValidator, ","))
				swapBeaconInstructions = append(swapBeaconInstructions, "beacon")
				swapBeaconInstructions = append(swapBeaconInstructions, string(badProducersWithPunishmentBytes))
				instructions = append(instructions, swapBeaconInstructions)
				// Generate instruction storing validators pubkey and send to bridge
				beaconRootInst, _ := buildBeaconSwapConfirmInstruction(currentValidators, newBeaconHeight)
				instructions = append(instructions, beaconRootInst)
			}
		}
	}
	// Stake
	instructions = append(instructions, stakeInstructions...)
	// Stop Auto Staking
	instructions = append(instructions, stopAutoStakingInstructions...)
	// Random number for Assign Instruction
	if newBeaconHeight%chainParamEpoch > randomTime && !beaconBestState.IsGetRandomNumber {
		var err error
		var chainTimeStamp int64
		if !TestRandom {
			if newBeaconHeight%chainParamEpoch == chainParamEpoch-1 {
				startTime := time.Now()
				for {
					Logger.log.Criticalf("Block %+v, Enter final block of epoch but still no random number", newBeaconHeight)
					chainTimeStamp, err = blockchain.config.RandomClient.GetCurrentChainTimeStamp()
					if err != nil {
						Logger.log.Error(err)
					} else {
						if chainTimeStamp < beaconBestState.CurrentRandomTimeStamp {
							Logger.log.Infof("Final Block %+v in Epoch but still haven't found new random number", newBeaconHeight)
						} else {
							break
						}
					}
					if time.Since(startTime).Seconds() > beaconBestState.BlockMaxCreateTime.Seconds() {
						return [][]string{}, NewBlockChainError(GenerateInstructionError, fmt.Errorf("Get Current Chain Timestamp for New Block Height %+v Timeout", newBeaconHeight))
					}
					time.Sleep(100 * time.Millisecond)
				}
			} else {
				Logger.log.Criticalf("Block %+v, finding random number", newBeaconHeight)
				chainTimeStamp, err = blockchain.config.RandomClient.GetCurrentChainTimeStamp()
				if err != nil {
					Logger.log.Error(err)
				}
			}
		} else {
			chainTimeStamp = beaconBestState.CurrentRandomTimeStamp + 1
		}
		//==================================
		if err == nil && chainTimeStamp > beaconBestState.CurrentRandomTimeStamp {
			numberOfPendingValidator := make(map[byte]int)
			for i := 0; i < beaconBestState.ActiveShards; i++ {
				if pendingValidators, ok := beaconBestState.ShardPendingValidator[byte(i)]; ok {
					numberOfPendingValidator[byte(i)] = len(pendingValidators)
				} else {
					numberOfPendingValidator[byte(i)] = 0
				}
			}
			randomInstruction, rand, err := beaconBestState.generateRandomInstruction(beaconBestState.CurrentRandomTimeStamp, blockchain.config.RandomClient)
			if err != nil {
				return [][]string{}, err
			}
			instructions = append(instructions, randomInstruction)
			Logger.log.Infof("Beacon Producer found Random Instruction at Block Height %+v, %+v", randomInstruction, newBeaconHeight)
			shardCandidatesStr, err := incognitokey.CommitteeKeyListToString(shardCandidates)
			if err != nil {
				panic(err)
			}
			_, assignedCandidates := assignShardCandidate(shardCandidatesStr, numberOfPendingValidator, rand, blockchain.config.ChainParams.AssignOffset, beaconBestState.ActiveShards)
			var keys []int
			for k := range assignedCandidates {
				keys = append(keys, int(k))
			}
			sort.Ints(keys)
			for _, key := range keys {
				shardID := byte(key)
				candidates := assignedCandidates[shardID]
				Logger.log.Infof("Assign Candidate at Shard %+v: %+v", shardID, candidates)
				shardAssingInstruction := []string{AssignAction}
				shardAssingInstruction = append(shardAssingInstruction, strings.Join(candidates, ","))
				shardAssingInstruction = append(shardAssingInstruction, "shard")
				shardAssingInstruction = append(shardAssingInstruction, fmt.Sprintf("%v", shardID))
				instructions = append(instructions, shardAssingInstruction)
			}
		}
	}
	return instructions, nil
}
