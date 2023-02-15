package blockchain

import (
	"fmt"
	"log"
	"math/big"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/privacy/key"

	"github.com/incognitochain/incognito-chain/transaction"

	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

type mintDelegationRewardInfo struct {
	TxRequest       metadata.Transaction
	ReceiverPayment privacy.PaymentAddress
	RewardAmount    uint64
}

func (blockchain *BlockChain) addShardRewardRequestToBeacon(beaconBlock *types.BeaconBlock, rewardStateDB *statedb.StateDB, bestState *BeaconBestState) error {
	//get shard block version that confirmed by this beacon block
	shardsBlockVersions := map[int]map[uint64]int{}
	for _, sID := range blockchain.GetShardIDs() {
		shardsBlockVersions[sID] = map[uint64]int{}
		sStates := beaconBlock.Body.ShardState[byte(sID)]
		for _, sState := range sStates {
			shardsBlockVersions[sID][sState.Height] = sState.Version
		}
	}

	for _, inst := range beaconBlock.Body.Instructions {
		if len(inst) <= 2 {
			continue
		}
		version := beaconBlock.Header.Version
		if inst[0] == instruction.ACCEPT_BLOCK_REWARD_V3_ACTION {
			acceptBlockRewardIns, err := instruction.ValidateAndImportAcceptBlockRewardV3InstructionFromString(inst)
			if err != nil {
				return err
			}
			if shardBlockVersions, ok := shardsBlockVersions[int(acceptBlockRewardIns.ShardID())]; ok {
				if blkVersion, ok := shardBlockVersions[acceptBlockRewardIns.ShardBlockHeight()]; ok {
					version = blkVersion
				}
			}
			rewardAmount, err := blockchain.GetRewardAmount(acceptBlockRewardIns.ShardID(), version, acceptBlockRewardIns.ShardBlockHeight())
			if err != nil {
				return err
			} else {
				if beaconBlock.Header.Version >= types.INSTANT_FINALITY_VERSION_V2 {
					if bestState.RewardMinted+rewardAmount > config.Param().MaxReward {
						if config.Param().MaxReward > bestState.RewardMinted {
							rewardAmount = config.Param().MaxReward - bestState.RewardMinted
						} else {
							rewardAmount = 0
						}
					}
					bestState.RewardMinted += rewardAmount
				}
			}

			acceptBlockRewardIns.TxsFee()[common.PRVCoinID] += rewardAmount

			for key, value := range acceptBlockRewardIns.TxsFee() {
				if value != 0 {
					err = statedb.AddShardRewardRequestMultiset(
						rewardStateDB,
						beaconBlock.Header.Epoch,
						acceptBlockRewardIns.ShardID(),
						acceptBlockRewardIns.SubsetID(),
						key, value)
					if err != nil {
						return err
					}
				}
			}
			continue
		}
		if instruction.IsConsensusInstruction(inst[0]) {
			continue
		}
		metaType, err := strconv.Atoi(inst[0])
		if err != nil {
			continue
		}
		if metaType == instruction.ACCEPT_BLOCK_REWARD_V1_ACTION {
			acceptedBlkRewardInfo, err := instruction.NewAcceptedBlockRewardV1FromString(inst[2])
			if err != nil {
				return err
			}

			if acceptedBlkRewardInfo.TxsFee == nil {
				acceptedBlkRewardInfo.TxsFee = map[common.Hash]uint64{}
			}
			if shardBlockVersions, ok := shardsBlockVersions[int(acceptedBlkRewardInfo.ShardID)]; ok {
				if blkVersion, ok := shardBlockVersions[acceptedBlkRewardInfo.ShardBlockHeight]; ok {
					version = blkVersion
				}
			}
			rewardAmount, err := blockchain.GetRewardAmount(acceptedBlkRewardInfo.ShardID, version, acceptedBlkRewardInfo.ShardBlockHeight)
			if err != nil {
				return err
			} else {
				if beaconBlock.Header.Version >= types.INSTANT_FINALITY_VERSION_V2 {
					if bestState.RewardMinted+rewardAmount > config.Param().MaxReward {
						if config.Param().MaxReward > bestState.RewardMinted {
							rewardAmount = config.Param().MaxReward - bestState.RewardMinted
						} else {
							rewardAmount = 0
						}
					}
					bestState.RewardMinted += rewardAmount
				}
			}
			acceptedBlkRewardInfo.TxsFee[common.PRVCoinID] += rewardAmount

			for key, value := range acceptedBlkRewardInfo.TxsFee {
				if value != 0 {
					err = statedb.AddShardRewardRequest(
						rewardStateDB,
						beaconBlock.Header.Epoch,
						acceptedBlkRewardInfo.ShardID,
						key, value)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (blockchain *BlockChain) processSalaryInstructions(rewardStateDB *statedb.StateDB, beaconBlocks []*types.BeaconBlock, confirmBeaconHeight uint64, shardID byte) error {
	for _, beaconBlock := range beaconBlocks {
		for _, l := range beaconBlock.Body.Instructions {

			if len(l) <= 2 {
				continue
			}

			if l[0] == instruction.SHARD_RECEIVE_REWARD_V3_ACTION {
				shardReceiveRewardV3, err := instruction.ValidateAndImportShardReceiveRewardV3InstructionFromString(l)
				if err != nil {
					Logger.log.Debug(err)
					continue
				}
				if shardReceiveRewardV3.Epoch() != 0 {
					if confirmBeaconHeight < config.Param().ConsensusParam.EnableSlashingHeightV2 {
						cInfos, err := blockchain.GetAllCommitteeStakeInfo(shardReceiveRewardV3.Epoch())
						if err != nil {
							return NewBlockChainError(ProcessSalaryInstructionsError, err)
						}
						shardSubsetStakerInfo := getCommitteeToPayRewardMultiset(cInfos[int(shardReceiveRewardV3.ShardID())], shardReceiveRewardV3)
						err = blockchain.addShardCommitteeReward(rewardStateDB, shardID, shardReceiveRewardV3.Reward(), shardSubsetStakerInfo)
						if err != nil {
							return err
						}
					} else {
						cInfosV2, err := blockchain.GetAllCommitteeStakeInfoSlashingVersion(shardReceiveRewardV3.Epoch())
						if err != nil {
							return NewBlockChainError(ProcessSalaryInstructionsError, err)
						}
						shardSubsetStakerInfo := getCommitteeToPayRewardMultisetSlashingVersion(cInfosV2[int(shardReceiveRewardV3.ShardID())], shardReceiveRewardV3)
						beaconBestState := blockchain.BeaconChain.GetBestView().(*BeaconBestState)
						nonSlashingCInfosV2, err := beaconBestState.GetNonSlashingCommittee(shardSubsetStakerInfo, shardReceiveRewardV3.Epoch(), shardReceiveRewardV3.ShardID())
						if err != nil {
							return NewBlockChainError(ProcessSalaryInstructionsError, err)
						}
						err = blockchain.addShardCommitteeRewardSlashingVersion(rewardStateDB, shardID, shardReceiveRewardV3.Reward(), nonSlashingCInfosV2)
						if err != nil {
							return err
						}
					}
				}
				continue
			}

			if instruction.IsConsensusInstruction(l[0]) {
				continue
			}

			shardToProcess, err := strconv.Atoi(l[1])
			if err != nil {
				continue
			}
			instType, err := strconv.Atoi(l[0])
			if err != nil {
				return NewBlockChainError(ProcessSalaryInstructionsError, err)
			}
			if shardToProcess == int(shardID) {
				switch instType {
				case metadata.BeaconRewardRequestMeta:
					beaconBlkRewardInfo, err := metadata.NewBeaconBlockRewardInfoFromStr(l[3])
					if err != nil {
						return NewBlockChainError(ProcessSalaryInstructionsError, err)
					}
					for key := range beaconBlkRewardInfo.BeaconReward {
						Logger.log.Criticalf("Add Committee Reward BeaconReward, Public Key %+v, reward %+v, token %+v", beaconBlkRewardInfo.PayToPublicKey, beaconBlkRewardInfo.BeaconReward[key], key)
						err = statedb.AddCommitteeReward(rewardStateDB, beaconBlkRewardInfo.PayToPublicKey, beaconBlkRewardInfo.BeaconReward[key], key)
						if err != nil {
							return NewBlockChainError(ProcessSalaryInstructionsError, err)
						}
					}
					continue

				case metadata.IncDAORewardRequestMeta:
					incDAORewardInfo, err := metadata.NewIncDAORewardInfoFromStr(l[3])
					if err != nil {
						return NewBlockChainError(ProcessSalaryInstructionsError, err)
					}
					keyWalletDevAccount, err := wallet.Base58CheckDeserialize(config.Param().IncognitoDAOAddress)
					if err != nil {
						return NewBlockChainError(ProcessSalaryInstructionsError, err)
					}
					for key := range incDAORewardInfo.IncDAOReward {
						tempPublicKey := base58.Base58Check{}.Encode(keyWalletDevAccount.KeySet.PaymentAddress.Pk, common.Base58Version)
						Logger.log.Criticalf("Add Committee Reward IncDAOReward, Public Key %+v, reward %+v, token %+v", tempPublicKey, incDAORewardInfo.IncDAOReward[key], key)
						err = statedb.AddCommitteeReward(rewardStateDB, tempPublicKey, incDAORewardInfo.IncDAOReward[key], key)
						if err != nil {
							return NewBlockChainError(ProcessSalaryInstructionsError, err)
						}
					}
					continue
				}
			}
			switch instType {
			case instruction.SHARD_RECEIVE_REWARD_V1_ACTION:
				shardRewardInfo, err := instruction.NewShardReceiveRewardV1FromString(l[3])
				if err != nil {
					return NewBlockChainError(ProcessSalaryInstructionsError, err)
				}
				if confirmBeaconHeight < config.Param().ConsensusParam.EnableSlashingHeightV2 {
					cInfos, err := blockchain.GetAllCommitteeStakeInfo(shardRewardInfo.Epoch)
					if err != nil {
						return NewBlockChainError(ProcessSalaryInstructionsError, err)
					}
					err = blockchain.addShardCommitteeReward(rewardStateDB, shardID, shardRewardInfo.ShardReward, cInfos[int(shardToProcess)])
					if err != nil {
						return err
					}
				} else {
					cInfosV2, err := blockchain.GetAllCommitteeStakeInfoSlashingVersion(shardRewardInfo.Epoch)
					if err != nil {
						return NewBlockChainError(ProcessSalaryInstructionsError, err)
					}
					beaconBestState := blockchain.BeaconChain.GetBestView().(*BeaconBestState)
					nonSlashingCInfosV2, err := beaconBestState.GetNonSlashingCommittee(cInfosV2[int(shardToProcess)], shardRewardInfo.Epoch, byte(shardToProcess))
					if err != nil {
						return NewBlockChainError(ProcessSalaryInstructionsError, err)
					}
					err = blockchain.addShardCommitteeRewardSlashingVersion(rewardStateDB, shardID, shardRewardInfo.ShardReward, nonSlashingCInfosV2)
					if err != nil {
						return err
					}
				}
				continue
			}

		}
	}
	return nil
}

func getCommitteeToPayRewardMultiset(
	committees []*statedb.StakerInfo,
	shardReceiveRewardV3 *instruction.ShardReceiveRewardV3,
) []*statedb.StakerInfo {
	res := []*statedb.StakerInfo{}
	for i, v := range committees {
		if i%MaxSubsetCommittees == int(shardReceiveRewardV3.SubsetID()) {
			res = append(res, v)
		}
	}
	return res
}

func getCommitteeToPayRewardMultisetSlashingVersion(
	committees []*statedb.StakerInfoSlashingVersion,
	shardReceiveRewardV3 *instruction.ShardReceiveRewardV3,
) []*statedb.StakerInfoSlashingVersion {
	res := []*statedb.StakerInfoSlashingVersion{}
	for i, v := range committees {
		if i%MaxSubsetCommittees == int(shardReceiveRewardV3.SubsetID()) {
			res = append(res, v)
		}
	}
	return res
}

func (blockchain *BlockChain) addShardCommitteeReward(
	rewardStateDB *statedb.StateDB,
	shardID byte,
	reward map[common.Hash]uint64,
	cStakeInfos []*statedb.StakerInfo,
) (
	err error,
) {
	committeeSize := len(cStakeInfos)
	for _, candidate := range cStakeInfos {
		if common.GetShardIDFromLastByte(candidate.RewardReceiver().Pk[common.PublicKeySize-1]) == shardID {
			for key, value := range reward {
				tempPK := base58.Base58Check{}.Encode(candidate.RewardReceiver().Pk, common.Base58Version)
				Logger.log.Criticalf("Add Committee Reward ShardCommitteeReward, Public Key %+v, reward %+v, token %+v", tempPK, value/uint64(committeeSize), key)
				err = statedb.AddCommitteeReward(rewardStateDB, tempPK, value/uint64(committeeSize), key)
				if err != nil {
					return NewBlockChainError(ProcessSalaryInstructionsError, err)
				}
			}
		}
	}
	return nil
}

func (blockchain *BlockChain) addShardCommitteeRewardSlashingVersion(
	rewardStateDB *statedb.StateDB,
	shardID byte,
	reward map[common.Hash]uint64,
	cStakeInfos []*statedb.StakerInfoSlashingVersion,
) (
	err error,
) {
	committeeSize := len(cStakeInfos)
	for _, candidate := range cStakeInfos {
		if common.GetShardIDFromLastByte(candidate.RewardReceiver().Pk[common.PublicKeySize-1]) == shardID {
			for key, value := range reward {
				tempPK := base58.Base58Check{}.Encode(candidate.RewardReceiver().Pk, common.Base58Version)
				Logger.log.Criticalf("Add Committee Reward ShardCommitteeReward, Public Key %+v, reward %+v, token %+v", tempPK, value/uint64(committeeSize), key)
				err = statedb.AddCommitteeReward(rewardStateDB, tempPK, value/uint64(committeeSize), key)
				if err != nil {
					return NewBlockChainError(ProcessSalaryInstructionsError, err)
				}
			}
		}
	}
	return nil
}

func (blockchain *BlockChain) calculateRewardMultiset(
	splitRewardRuleProcessor committeestate.SplitRewardRuleProcessor,
	curView *BeaconBestState,
	maxBeaconBlockCreation uint64,
	maxSubsetsCommittee int,
	beaconHeight uint64,
	epoch uint64,
	isSplitRewardForCustodian bool,
	percentCustodianRewards uint64,
) (map[common.Hash]uint64,
	[][]map[common.Hash]uint64,
	map[common.Hash]uint64,
	map[common.Hash]uint64, error,
) {
	allCoinID := statedb.GetAllTokenIDForRewardMultiset(curView.rewardStateDB, epoch)
	currentBeaconYear, err := blockchain.GetYearOfBeacon(beaconHeight)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	percentForIncognitoDAO := getPercentForIncognitoDAOV2(currentBeaconYear)
	totalRewardForShardSubset := make([][]map[common.Hash]uint64, curView.ActiveShards)
	totalRewards := make([][]map[common.Hash]uint64, curView.ActiveShards)
	totalRewardForBeacon := map[common.Hash]uint64{}
	totalRewardForIncDAO := map[common.Hash]uint64{}
	totalRewardForCustodian := map[common.Hash]uint64{}

	for shardID := 0; shardID < curView.ActiveShards; shardID++ {
		totalRewardForShardSubset[shardID] = make([]map[common.Hash]uint64, maxSubsetsCommittee)
		totalRewards[shardID] = make([]map[common.Hash]uint64, maxSubsetsCommittee)
		for subsetID := 0; subsetID < maxSubsetsCommittee; subsetID++ {
			if totalRewards[shardID][subsetID] == nil {
				totalRewards[shardID][subsetID] = map[common.Hash]uint64{}
			}
			if totalRewardForShardSubset[shardID][subsetID] == nil {
				totalRewardForShardSubset[shardID][subsetID] = map[common.Hash]uint64{}
			}

			for _, coinID := range allCoinID {
				totalRewards[shardID][subsetID][coinID], err = statedb.GetRewardOfShardByEpochMultiset(
					curView.rewardStateDB, epoch,
					byte(shardID), byte(subsetID), coinID)
				if err != nil {
					return nil, nil, nil, nil, err
				}
				if totalRewards[shardID][subsetID][coinID] == 0 {
					delete(totalRewards[shardID][subsetID], coinID)
				}
			}

			env := committeestate.NewSplitRewardEnvironmentMultiset(
				byte(shardID),
				byte(subsetID),
				byte(maxSubsetsCommittee),
				beaconHeight,
				totalRewards[shardID][subsetID],
				isSplitRewardForCustodian,
				percentCustodianRewards,
				percentForIncognitoDAO,
				curView.GetBeaconCommittee(),
				curView.GetShardCommittee(),
			)

			Logger.log.Info("[dcs] env.MaxSubsetCommittees:", env.MaxSubsetCommittees)
			rewardForBeacon, rewardForShardSubset, rewardForDAO, rewardForCustodian, err := splitRewardRuleProcessor.SplitReward(env)
			if err != nil {
				return nil, nil, nil, nil, err
			}

			plusMap(rewardForBeacon, totalRewardForBeacon)
			plusMap(rewardForShardSubset, totalRewardForShardSubset[shardID][subsetID])
			plusMap(rewardForDAO, totalRewardForIncDAO)
			plusMap(rewardForCustodian, totalRewardForCustodian)
		}
	}

	return totalRewardForBeacon, totalRewardForShardSubset, totalRewardForIncDAO, totalRewardForCustodian, nil
}

func (blockchain *BlockChain) calculateReward(
	splitRewardRuleProcessor committeestate.SplitRewardRuleProcessor,
	curView *BeaconBestState,
	numberOfActiveShards int,
	beaconHeight uint64,
	epoch uint64,
	rewardStateDB *statedb.StateDB,
	isSplitRewardForCustodian bool,
	percentCustodianRewards uint64,
) (map[common.Hash]uint64,
	[]map[common.Hash]uint64,
	map[common.Hash]uint64,
	map[common.Hash]uint64, uint64, error,
) {
	allCoinID := statedb.GetAllTokenIDForReward(rewardStateDB, epoch)
	currentBeaconYear, err := blockchain.GetYearOfBeacon(beaconHeight)
	if err != nil {
		return nil, nil, nil, nil, 0, err
	}
	percentForIncognitoDAO := getPercentForIncognitoDAOV2(currentBeaconYear)
	totalRewardForShard := make([]map[common.Hash]uint64, numberOfActiveShards)
	totalRewards := make([]map[common.Hash]uint64, numberOfActiveShards)
	totalRewardForBeacon := map[common.Hash]uint64{}
	totalRewardForIncDAO := map[common.Hash]uint64{}
	totalRewardForCustodian := map[common.Hash]uint64{}
	delegationReward := uint64(0)
	for id := 0; id < numberOfActiveShards; id++ {
		if totalRewards[id] == nil {
			totalRewards[id] = map[common.Hash]uint64{}
		}
		if totalRewardForShard[id] == nil {
			totalRewardForShard[id] = map[common.Hash]uint64{}
		}

		for _, coinID := range allCoinID {
			totalRewards[id][coinID], err = statedb.GetRewardOfShardByEpoch(rewardStateDB, epoch, byte(id), coinID)
			if err != nil {
				return nil, nil, nil, nil, 0, err
			}
			if totalRewards[id][coinID] == 0 {
				delete(totalRewards[id], coinID)
			}
		}

		env := committeestate.NewSplitRewardEnvironmentV1(
			byte(id),
			beaconHeight,
			totalRewards[id],
			isSplitRewardForCustodian,
			percentCustodianRewards,
			percentForIncognitoDAO,
			numberOfActiveShards,
			curView.GetBeaconCommittee(),
			curView.GetShardCommittee(),
		)
		rewardForBeacon, rewardForShard, rewardForDAO, rewardForCustodian, err := splitRewardRuleProcessor.SplitReward(env)
		if err != nil {
			return nil, nil, nil, nil, 0, err
		}

		if curView.TriggeredFeature[config.DELEGATION_REWARD] != 0 {
			shardDelegationReward := rewardForShard[common.PRVCoinID] * 5 / 10
			rewardForShard[common.PRVCoinID] = rewardForShard[common.PRVCoinID] - shardDelegationReward
			delegationReward += shardDelegationReward
		}

		plusMap(rewardForBeacon, totalRewardForBeacon)
		plusMap(rewardForShard, totalRewardForShard[id])
		plusMap(rewardForDAO, totalRewardForIncDAO)
		plusMap(rewardForCustodian, totalRewardForCustodian)
	}

	return totalRewardForBeacon, totalRewardForShard, totalRewardForIncDAO, totalRewardForCustodian, delegationReward, nil
}

func (blockchain *BlockChain) buildRewardInstructionByEpoch(
	curView *BeaconBestState,
	blkHeight, epoch uint64,
	isSplitRewardForCustodian bool,
	percentCustodianRewards uint64,
	isSplitRewardForPdex bool,
	pdexRewardPercent uint,
	blockVersion int,
) ([][]string, map[common.Hash]uint64, uint64, error) {

	//Declare variables
	var resInst [][]string
	var err error
	var instRewardForBeacons [][]string
	var instRewardForIncDAO [][]string
	var instRewardForShards [][]string
	var delegationRewardInst [][]string
	beaconBestView := blockchain.BeaconChain.GetBestView().(*BeaconBestState)

	totalRewardForBeacon := make(map[common.Hash]uint64)
	totalRewardForShard := make([]map[common.Hash]uint64, beaconBestView.ActiveShards)
	totalRewardForShardSubset := make([][]map[common.Hash]uint64, beaconBestView.ActiveShards)
	totalRewardForCustodian := make(map[common.Hash]uint64)
	totalRewardForIncDAO := make(map[common.Hash]uint64)
	rewardForPdex := uint64(0)
	delegationReward := uint64(0)
	if blockVersion >= types.BLOCK_PRODUCINGV3_VERSION && blockVersion < types.INSTANT_FINALITY_VERSION_V2 {
		splitRewardRuleProcessor := committeestate.GetRewardSplitRule(blockVersion)
		totalRewardForBeacon,
			totalRewardForShardSubset,
			totalRewardForIncDAO,
			totalRewardForCustodian,
			err = blockchain.calculateRewardMultiset(
			splitRewardRuleProcessor,
			curView,
			uint64(config.Param().BlockTime.MaxBeaconBlockCreation.Seconds()),
			MaxSubsetCommittees,
			blkHeight,
			epoch,
			isSplitRewardForCustodian,
			percentCustodianRewards,
		)

		instRewardForShards, err = blockchain.buildInstructionRewardForShardsV3(epoch, totalRewardForShardSubset)
		if err != nil {
			return nil, nil, rewardForPdex, err
		}

	} else {
		splitRewardRuleProcessor := committeestate.GetRewardSplitRule(blockVersion)
		totalRewardForBeacon,
			totalRewardForShard,
			totalRewardForIncDAO,
			totalRewardForCustodian,
			delegationReward,
			err = blockchain.calculateReward(
			splitRewardRuleProcessor,
			curView,
			curView.ActiveShards, blkHeight, epoch,
			curView.GetBeaconRewardStateDB(),
			isSplitRewardForCustodian, percentCustodianRewards,
		)

		instRewardForShards, err = blockchain.buildInstructionRewardForShards(epoch, totalRewardForShard)
		if err != nil {
			return nil, nil, rewardForPdex, err
		}
	}

	if len(totalRewardForBeacon) > 0 {
		committeeOfEpoch, err := blockchain.GetBeaconCommitteeOfEpoch(epoch)
		if err != nil {
			return nil, nil, rewardForPdex, err
		}
		committeeGotReward, err := curView.beaconCommitteeState.GetNonSlashingRewardReceiver(committeeOfEpoch)
		if err != nil {
			return nil, nil, rewardForPdex, err
		}
		instRewardForBeacons, err = curView.buildInstRewardForBeacons(epoch, totalRewardForBeacon, committeeGotReward)
		if err != nil {
			return nil, nil, rewardForPdex, err
		}
	}

	if len(totalRewardForIncDAO) > 0 {
		prvReward := totalRewardForIncDAO[common.PRVCoinID]
		if prvReward > 0 && isSplitRewardForPdex {
			temp := new(big.Int).Mul(new(big.Int).SetUint64(prvReward), new(big.Int).SetUint64(uint64(pdexRewardPercent)))
			temp = temp.Div(temp, big.NewInt(100))
			if !temp.IsUint64() {
				return nil, nil, rewardForPdex, errors.New("Reward for Pdex is not uint64")
			}
			rewardForPdex = temp.Uint64()
			totalRewardForIncDAO[common.PRVCoinID] = prvReward - rewardForPdex
		}
		instRewardForIncDAO, err = blockchain.buildInstRewardForIncDAO(epoch, totalRewardForIncDAO)
		if err != nil {
			return nil, nil, rewardForPdex, err
		}
	}

	if delegationReward > 0 {
		Logger.log.Info("delegationReward", delegationReward)
		delegationRewardInst, err = curView.CalculateDelegationSharePrice(blockchain, delegationReward)
		if err != nil {
			panic(err)
		}
	}

	resInst = common.AppendSliceString(instRewardForBeacons, instRewardForIncDAO, instRewardForShards, delegationRewardInst)
	return resInst, totalRewardForCustodian, rewardForPdex, nil
}

// buildInstRewardForBeacons create reward instruction for beacons
func (beaconBestState *BeaconBestState) buildInstRewardForBeacons(epoch uint64, totalReward map[common.Hash]uint64, rewardReceiver []key.PaymentAddress) ([][]string, error) {
	resInst := [][]string{}
	baseRewards := map[common.Hash]uint64{}
	for key, value := range totalReward {
		baseRewards[key] = value / uint64(len(rewardReceiver))
	}
	for _, receiver := range rewardReceiver {
		singleInst, err := metadata.BuildInstForBeaconReward(baseRewards, receiver.Pk)
		if err != nil {
			Logger.log.Errorf("BuildInstForBeaconReward error %+v\n Totalreward: %+v, epoch: %+v, reward: %+v\n", err, totalReward, epoch, baseRewards)
			return nil, err
		}
		resInst = append(resInst, singleInst)
	}
	return resInst, nil
}

func (blockchain *BlockChain) buildInstRewardForIncDAO(epoch uint64, totalReward map[common.Hash]uint64) ([][]string, error) {
	resInst := [][]string{}
	devRewardInst, err := metadata.BuildInstForIncDAOReward(totalReward, config.Param().IncognitoDAOAddress)
	if err != nil {
		Logger.log.Errorf("buildInstRewardForIncDAO error %+v\n Totalreward: %+v, epoch: %+v\n", err, totalReward, epoch)
		return nil, err
	}
	resInst = append(resInst, devRewardInst)
	return resInst, nil
}

func (blockchain *BlockChain) buildInstructionRewardForShardsV3(epoch uint64, totalRewards [][]map[common.Hash]uint64) ([][]string, error) {
	resInst := [][]string{}

	for shardID, v := range totalRewards {
		for subsetID, reward := range v {
			if len(reward) > 0 {
				shardSubsetReward := instruction.NewShardReceiveRewardV3WithValue(reward, epoch, byte(shardID), byte(subsetID))
				shardSubsetRewardInst := shardSubsetReward.String()
				resInst = append(resInst, shardSubsetRewardInst)
			}
		}
	}

	return resInst, nil
}

func (blockchain *BlockChain) buildInstructionRewardForShards(epoch uint64, totalRewards []map[common.Hash]uint64) ([][]string, error) {
	resInst := [][]string{}
	for i, reward := range totalRewards {
		if len(reward) > 0 {
			shardRewardInst, err := instruction.NewShardReceiveRewardV1WithValue(reward, epoch, byte(i))
			if err != nil {
				Logger.log.Errorf("BuildInstForShardReward error %+v\n Totalreward: %+v, epoch: %+v\n; shard:%+v", err, reward, epoch, byte(i))
				return nil, err
			}
			resInst = append(resInst, shardRewardInst...)
		}
	}
	return resInst, nil
}

func (blockchain *BlockChain) buildWithDrawTransactionResponse(view *ShardBestState, txRequest *metadata.Transaction, blkProducerPrivateKey *privacy.PrivateKey, shardID byte) (metadata.Transaction, error) {
	if (*txRequest).GetMetadataType() != metadata.WithDrawRewardRequestMeta {
		return nil, errors.New("Can not understand this request!")
	}
	requestDetail := (*txRequest).GetMetadata().(*metadata.WithDrawRewardRequest)
	tempPublicKey := base58.Base58Check{}.Encode(requestDetail.PaymentAddress.Pk, common.Base58Version)
	amount, err := statedb.GetCommitteeReward(view.GetShardRewardStateDB(), tempPublicKey, requestDetail.TokenID)
	if (amount == 0) || (err != nil) {
		return nil, errors.New("Not enough reward")
	}
	responseMeta, err := metadata.NewWithDrawRewardResponse(requestDetail, (*txRequest).Hash())
	if err != nil {
		return nil, err
	}
	txParam := transaction.TxSalaryOutputParams{Amount: amount, ReceiverAddress: &requestDetail.PaymentAddress, TokenID: &requestDetail.TokenID}
	makeMD := func(c privacy.Coin) metadata.Metadata {
		if c != nil && c.GetSharedRandom() != nil {
			responseMeta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return responseMeta
	}

	salaryTx, err := txParam.BuildTxSalary(blkProducerPrivateKey, view.GetCopiedTransactionStateDB(), makeMD)
	if err != nil {
		return nil, errors.Errorf("cannot init salary tx. Error %v", err)
	}
	salaryTx.SetType(common.TxRewardType)
	return salaryTx, nil
}

func (blockchain *BlockChain) GetBlockTimeByBlockVersion(blkVersion int) (int64, error) {
	blockTimeMap := config.Param().BlockTimeParam
	defaultBlockTime := blockTimeMap[BLOCKTIME_DEFAULT]

	blockTime := defaultBlockTime
	for _, anchor := range blockchain.GetBeaconBestState().TSManager.Anchors {
		if config.Param().FeatureVersion[anchor.Feature] > int64(blkVersion) {
			break
		}
		blockTime = int64(anchor.Timeslot)
	}

	return blockTime, nil
}

func (blockchain *BlockChain) GetBasicRewardByVersion(version int) (uint64, error) {
	blockTimeMap := config.Param().BlockTimeParam
	defaultBlockTime := blockTimeMap[BLOCKTIME_DEFAULT]
	curBlockTime, err := blockchain.GetBlockTimeByBlockVersion(version)
	if err != nil {
		return 0, err
	}
	return config.Param().BasicReward * uint64(curBlockTime) / uint64(defaultBlockTime), nil
}

func (blockchain *BlockChain) getRewardAmount(blkHeight uint64) uint64 {
	blockBeaconInterval := config.Param().BlockTime.MinBeaconBlockInterval.Seconds()
	blockInYear := getNoBlkPerYear(uint64(blockBeaconInterval))
	n := (blkHeight - 1) / blockInYear
	reward := uint64(config.Param().BasicReward)
	for ; n > 0; n-- {
		reward *= 91
		reward /= 100
	}
	return reward
}

func (blockchain *BlockChain) GetRewardAmount(shardID byte, shardVersion int, shardHeight uint64) (uint64, error) {
	if shardVersion < types.ADJUST_BLOCKTIME_VERSION {
		return blockchain.getRewardAmount(shardHeight), nil
	}
	basicReward, err := blockchain.GetBasicRewardByVersion(shardVersion)
	if err != nil {
		return 0, err
	}
	yearOfShardHeight, err := blockchain.GetYearOfShard(shardID, shardHeight)
	if err != nil {
		return 0, err
	}
	return blockchain.getRewardAmountV2(basicReward, yearOfShardHeight), nil

}

func (blockchain *BlockChain) getRewardAmountV2(basicReward, year uint64) uint64 {
	n := year
	reward := basicReward
	for ; n > 0; n-- {
		reward *= 91
		reward /= 100
	}
	return reward
}

func (blockchain *BlockChain) GetYearOfBeacon(blockHeight uint64) (uint64, error) {

	bView := blockchain.GetBeaconBestState()
	if bView == nil {
		return 0, errors.Errorf("Can not get beacon view for get reward amount at block beacon %v", blockHeight)
	}
	featureManager := bView.TSManager
	return getYearOfBlockChain(&featureManager, blockHeight), nil
}

func (blockchain *BlockChain) GetYearOfShard(sID byte, blockHeight uint64) (uint64, error) {

	bView := blockchain.GetBeaconBestState()
	if bView == nil {
		return 0, errors.Errorf("Can not get beacon view for get reward amount at block beacon %v", blockHeight)
	}
	featureManager := bView.ShardTSManager[sID]
	return getYearOfBlockChain(featureManager, blockHeight), nil
}

func getYearOfBlockChain(featuresManager *TSManager, blkHeight uint64) uint64 {
	anchors := featuresManager.Anchors
	startBlkHeight := uint64(0)
	endBlkHeight := uint64(0)
	prevBlockTime := config.Param().BlockTimeParam[BLOCKTIME_DEFAULT]
	defBlockTime := config.Param().BlockTimeParam[BLOCKTIME_DEFAULT]
	totalBlock := uint64(0)
	if len(anchors) > 0 {
		for _, anchor := range anchors {
			endBlkHeight = anchor.BlockHeight
			if endBlkHeight > blkHeight {
				endBlkHeight = blkHeight
			}
			// fmt.Printf("%v - %v - %v - %v\n", prevBlockTime, startBlkHeight, endBlkHeight, (endBlkHeight-startBlkHeight)*uint64(prevBlockTime)/uint64(defBlockTime))
			totalBlock += (endBlkHeight - startBlkHeight) * uint64(prevBlockTime) / uint64(defBlockTime)
			startBlkHeight = endBlkHeight
			prevBlockTime = config.Param().BlockTimeParam[anchor.Feature]
		}
	}
	if endBlkHeight < blkHeight {
		endBlkHeight = blkHeight
		// fmt.Printf("%v - %v - %v - %v\n", prevBlockTime, startBlkHeight, endBlkHeight, (endBlkHeight-startBlkHeight)*uint64(prevBlockTime)/uint64(defBlockTime))
		totalBlock += (endBlkHeight - startBlkHeight) * uint64(prevBlockTime) / uint64(defBlockTime)
	}
	// fmt.Printf("%v - %v\n", totalBlock, getNoBlkPerYear(uint64(defBlockTime)))
	return totalBlock / getNoBlkPerYear(uint64(defBlockTime))
}

func getNoBlkPerYear(blockCreationTimeSeconds uint64) uint64 {
	return (365.25 * 24 * 60 * 60) / blockCreationTimeSeconds
}

func GetBlockTimeInterval(blkTimeFeature string) int64 {
	blockTimeMap := config.Param().BlockTimeParam
	if blkTime, ok := blockTimeMap[blkTimeFeature]; ok {
		return int64(blkTime)
	}
	return int64(config.Param().BlockTime.MinBeaconBlockInterval.Seconds())

}

func GetNumberBlkPerYear(blkTimeFeature string) uint64 {
	return getNoBlkPerYear(uint64(GetBlockTimeInterval(blkTimeFeature)))
}

func getPercentForIncognitoDAO(blockHeight, blkPerYear uint64) int {
	year := (blockHeight - 1) / blkPerYear
	if year > (UpperBoundPercentForIncDAO - LowerBoundPercentForIncDAO) {
		return LowerBoundPercentForIncDAO
	} else {
		return UpperBoundPercentForIncDAO - int(year)
	}
}

func getPercentForIncognitoDAOV2(year uint64) int {
	if year > (UpperBoundPercentForIncDAO - LowerBoundPercentForIncDAO) {
		return LowerBoundPercentForIncDAO
	} else {
		return UpperBoundPercentForIncDAO - int(year)
	}
}

// plusMap(src, dst): dst = dst + src
func plusMap(src, dst map[common.Hash]uint64) {
	if src != nil {
		for key, value := range src {
			dst[key] += value
		}
	}
}

func (blockchain *BlockChain) buildMintDRewardTxsFromBeaconInstructions(
	curView *ShardBestState,
	beaconBlocks []*types.BeaconBlock,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (
	[]metadata.Transaction,
	error,
) {
	mintTxs := []metadata.Transaction{}
	mintInfoByRequestPK, err := blockchain.getMintDRewardInfoFromBeaconInstructions(curView, beaconBlocks, shardID)
	if err != nil {
		return nil, err
	}
	pks := []string{}
	for pk := range mintInfoByRequestPK {
		pks = append(pks, pk)
	}
	sort.Slice(pks, func(i, j int) bool {
		return pks[i] < pks[j]
	})
	for _, pk := range pks {
		mintInfo := mintInfoByRequestPK[pk]
		mintTx, amount, err := blockchain.buildMintDelegationRewardTransaction(
			curView,
			&mintInfo,
			producerPrivateKey,
			shardID,
		)
		if err != nil {
			return nil, err
		}
		Logger.log.Infof("Build tx mint delegation reward successful, txID %v, amount %v, for pk %v", mintTx.Hash().String(), amount, pk)
		mintTxs = append(mintTxs, mintTx)
	}
	return mintTxs, nil
}

func (blockchain *BlockChain) getMintDRewardInfoFromBeaconInstructions(
	curView *ShardBestState,
	bBlocks []*types.BeaconBlock,
	shardID byte,
) (
	map[string]mintDelegationRewardInfo, //pk request -> mint info
	error,
) {
	mintInsts := getMintDRewardInstructionsFromBeaconBlocks(bBlocks)
	if len(mintInsts) == 0 {
		return nil, nil
	}
	mintInfoByRequestID := map[common.Hash]interface{}{}
	mintInfoByRequestPK := map[string]mintDelegationRewardInfo{}
	for _, inst := range mintInsts {
		for idx, requestPayment := range inst.GetPaymentAddressesStruct() {
			pkRequest := base58.Base58Check{}.Encode(requestPayment.Pk, common.Base58Version)
			txRequestHash := inst.TxRequestIdHashes[idx]
			if _, ok := mintInfoByRequestID[txRequestHash]; ok {
				Logger.log.Infof("Duplicate request for the same tx request hash %v. Detail info: Pk %v, payment %v, reward amount %v", txRequestHash.String(), pkRequest, inst.PaymentAddresses[idx], inst.RewardAmount[idx])
				continue
			}
			if _, ok := mintInfoByRequestPK[pkRequest]; ok {
				Logger.log.Infof("Duplicate request for the same PK %v. Detail info: Tx request hash %v, payment %v, reward amount %v", pkRequest, txRequestHash.String(), inst.PaymentAddresses[idx], inst.RewardAmount[idx])
				continue
			}
			if sID := privacy.GetShardIDFromPublicKey(requestPayment.Pk); sID != shardID {
				Logger.log.Infof("SKIP This payment %v (shard %v) is not belong to current shard %v, tx request %v, amount %v", inst.PaymentAddresses[idx], sID, shardID, inst.TxRequestIDs[idx], inst.RewardAmount[idx])
				continue
			}
			rawTxRequest, err := blockchain.ShardChain[shardID].BlockStorage.GetStakingTx(txRequestHash)
			if err != nil {
				Logger.log.Infof("SKIP This request (payment %v, dreward amount %v) belong to current shard %v, but can not find tx request %v in database,", inst.PaymentAddresses[idx], inst.RewardAmount[idx], shardID, inst.TxRequestIDs[idx])
				continue
			}
			mInfo := mintDelegationRewardInfo{
				TxRequest:       rawTxRequest,
				ReceiverPayment: requestPayment,
				RewardAmount:    inst.RewardAmount[idx],
			}
			mintInfoByRequestID[txRequestHash] = nil
			mintInfoByRequestPK[pkRequest] = mInfo
		}
	}
	if len(mintInfoByRequestID) != len(mintInfoByRequestPK) {
		return nil, errors.Errorf("Get mint delegation reward info error, total txs request %v is different from total pk request %v", len(mintInfoByRequestID), len(mintInfoByRequestPK))
	}
	return mintInfoByRequestPK, nil
}

func (blockchain *BlockChain) ValidateMintDRewardTxsFromBeaconInstructions(
	curView *ShardBestState,
	beaconBlocks []*types.BeaconBlock,
	shardBlock *types.ShardBlock,
	shardID byte,
) error {
	mintInfoByRequestPKGot := map[string]mintDelegationRewardInfo{}
	mintInfoByRequestPKExpected, err := blockchain.getMintDRewardInfoFromBeaconInstructions(curView, beaconBlocks, shardID)
	if err != nil {
		return err
	}
	mintDRewardTxs := map[common.Hash]struct{}{}

	for _, tx := range shardBlock.Body.Transactions {
		switch tx.GetMetadata().(type) {
		case *metadata.MintDelegationRewardMetadata:
			if tx.GetType() != common.TxRewardType {
				return errors.Errorf("Wrong transaction type for metadata %v, expected type %v, got %v", tx.GetMetadata().GetType(), common.TxRewardType, tx.GetType())
			}
		default:
			continue
		}

		md := tx.GetMetadata().(*metadata.MintDelegationRewardMetadata)
		h, _ := common.Hash{}.NewHashFromStr(md.TxID)
		_, _, txRequest, err := blockchain.GetTransactionByHashWithShardID(*h, shardID)
		if err != nil {
			err = NewBlockChainError(WrongShardIDError, fmt.Errorf("This request delegation reward tx %v not found in this shard %+v", md.TxID, shardID))
			Logger.log.Error(err)
			return err
		}
		receiverPK := base58.Base58Check{}.Encode(md.ReceiverAddress.Pk, common.Base58Version)
		mintInfo := mintDelegationRewardInfo{
			ReceiverPayment: md.ReceiverAddress,
			TxRequest:       txRequest,
			RewardAmount:    md.RewardAmount,
		}
		mintInfoByRequestPKGot[receiverPK] = mintInfo
		txHash := txRequest.Hash()
		if _, ok := mintDRewardTxs[*txHash]; ok {
			return errors.Errorf("Double tx mint delegation reward from instruction for tx request %+v", tx)
		}
		mintDRewardTxs[*txHash] = struct{}{}

		isMinted, mintCoin, coinID, err := tx.GetTxMintData()
		if err != nil || !isMinted {
			return errors.Errorf("this is not tx mint for mint delegation reward. Error %v", err)
		}
		if coinID.String() != common.PRVIDStr {
			return errors.Errorf("Delegation reward tx only mints prv. Error token %v", coinID.String())
		}
		if ok := mintCoin.CheckCoinValid(mintInfo.ReceiverPayment, md.SharedRandom, mintInfo.RewardAmount); !ok {
			return errors.Errorf("mint data is invalid: Address %v; Amount %v, StakeAmount: %v", mintInfo.ReceiverPayment, mintCoin.GetValue(), mintInfo.RewardAmount)
		}

	}

	if len(mintInfoByRequestPKExpected) != len(mintInfoByRequestPKGot) {
		return errors.Errorf("List mint delegation tx of producer (len %v) and validator (len %v) not match", len(mintInfoByRequestPKGot), len(mintInfoByRequestPKExpected))
	}

	for k, gotInfo := range mintInfoByRequestPKGot {
		if expectedInfo, ok := mintInfoByRequestPKExpected[k]; !ok {
			return errors.Errorf("Mint delegation reward for pk %v of producer (tx request %v, amount %v) and validator (not found) not match", k, gotInfo.TxRequest.Hash().String(), gotInfo.RewardAmount)
		} else {
			if (expectedInfo.RewardAmount != gotInfo.RewardAmount) ||
				(expectedInfo.ReceiverPayment.String() != gotInfo.ReceiverPayment.String()) ||
				(expectedInfo.TxRequest.Hash().String() != gotInfo.TxRequest.String()) {
				return errors.Errorf("Mint delegation reward for pk %v of producer (tx request %v, amount %v, payment %v) and validator (tx request %v, amount %v, payment %v) not match", k, gotInfo.TxRequest.Hash().String(), gotInfo.RewardAmount, gotInfo.ReceiverPayment.String(), expectedInfo.TxRequest.Hash().String(), expectedInfo.RewardAmount, expectedInfo.TxRequest.Hash().String())
			}
		}
	}
	return nil
}

func getMintDRewardInstructionsFromBeaconBlocks(
	bBlocks []*types.BeaconBlock,
) []instruction.MintDelegationRewardInstruction {
	res := []instruction.MintDelegationRewardInstruction{}
	for _, bBlock := range bBlocks {
		for _, inst := range bBlock.Body.Instructions {
			if mI, err := instruction.ValidateAndImportMintDelegationRewardInstructionFromString(inst); err == nil {
				res = append(res, *mI)
			} else {
				if inst[0] == instruction.MINT_DREWARD_ACTION {
					Logger.log.Errorf("Can not parse instruction mint delegation reward from block %v-%v, inst %+v", bBlock.Header.Height, bBlock.Hash().String(), inst)
				}
			}
		}
	}
	return res
}

func (blockchain *BlockChain) buildMintDelegationRewardTransaction(
	curView *ShardBestState,
	mintInfo *mintDelegationRewardInfo,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (
	metadata.Transaction,
	uint64,
	error,
) {

	txRequestHash := mintInfo.TxRequest.Hash().String()
	mintDRewardMeta := metadata.NewMintDelegationReward(
		txRequestHash,
		mintInfo.ReceiverPayment,
		mintInfo.RewardAmount,
		metadata.MintDelegationRewardMeta,
	)
	txParam := transaction.TxSalaryOutputParams{
		Amount:          mintInfo.RewardAmount,
		ReceiverAddress: &mintInfo.ReceiverPayment,
		TokenID:         &common.PRVCoinID,
		Type:            common.TxRewardType,
	}

	makeMD := func(c privacy.Coin) metadata.Metadata {
		if c != nil && c.GetSharedRandom() != nil {
			mintDRewardMeta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return mintDRewardMeta
	}
	mintDRewardTx, err := txParam.BuildTxSalary(producerPrivateKey, curView.GetCopiedTransactionStateDB(), makeMD)
	if err != nil {
		return nil, 0, errors.Errorf("cannot init mint delegation reward tx. Error %v", err)
	}
	return mintDRewardTx, mintInfo.RewardAmount, nil
}
func (blockchain *BlockChain) GetBeaconSharePriceByEpoch(epoch uint64, uid string) (uint64, error) {
	blockHeight := (epoch - 1) * config.Param().EpochParam.NumberOfBlockInEpoch
	beaconConsensusStateRootHash, err := blockchain.GetBeaconRootsHashFromBlockHeight(
		blockHeight,
	)
	if err != nil {
		return 0, err
	}
	stateDB, err := statedb.NewWithPrefixTrie(beaconConsensusStateRootHash.ConsensusStateDBRootHash,
		statedb.NewDatabaseAccessWarper(blockchain.GetBeaconChainDatabase()))
	if err != nil {
		return 0, err
	}

	sharePrice, _, err := statedb.GetBeaconSharePrice(stateDB, uid)
	if err != nil {
		return 0, err
	}
	return sharePrice.GetPrice(), nil

}

func (blockchain *BlockChain) GetDelegationRewardAmount(stateDB *statedb.StateDB, pk key.PublicKey) (uint64, error) {
	rewardState, has, err := statedb.GetDelegationReward(stateDB, pk)
	if err != nil {
		return 0, err
	}
	if !has {
		return 0, nil
	}
	reward := uint64(0)
	for _, epochDelegate := range rewardState.Reward {
		//sort epoch
		epochSortList := []int{}
		for epoch, _ := range epochDelegate {
			epochSortList = append(epochSortList, epoch)
		}
		sort.Slice(epochSortList, func(i, j int) bool {
			return epochSortList[i] < epochSortList[j]
		})
		for i, epoch := range epochSortList {
			beaconID := epochDelegate[epoch].BeaconUID
			amount := epochDelegate[epoch].Amount
			startBeaconSharePrice, err := blockchain.GetBeaconSharePriceByEpoch(uint64(epoch), beaconID)
			if err != nil {
				return 0, errors.New("Can not get beacon share price")
			}
			unit := float64(amount) / float64(startBeaconSharePrice)
			if i == len(epochSortList)-1 {
				//last epoch
				endBeaconSharePrice, _, err := statedb.GetBeaconSharePrice(stateDB, beaconID)
				if err != nil {
					panic(fmt.Sprint("GetBeaconSharePriceByEpoch error 2", err))
				}
				endBeaconSharePrice.GetPrice()
				reward += uint64(unit * float64(endBeaconSharePrice.GetPrice()-startBeaconSharePrice))
				log.Println("lastBeaconID", beaconID, epoch, amount, startBeaconSharePrice, unit, endBeaconSharePrice.GetPrice(), reward)
			} else {
				endBeaconSharePrice, err := blockchain.GetBeaconSharePriceByEpoch(uint64(epochSortList[i+1]), beaconID)
				if err != nil {
					panic(fmt.Sprint("GetBeaconSharePriceByEpoch error 3", err))
				}
				reward += uint64(unit * float64(endBeaconSharePrice-startBeaconSharePrice))
				log.Println("BeaconID", beaconID, epoch, amount, startBeaconSharePrice, unit, endBeaconSharePrice, reward)
			}
		}
	}
	return reward, nil
}
