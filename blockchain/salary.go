package blockchain

import (
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/blockchain/committeestate"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

func (blockchain *BlockChain) addShardRewardRequestToBeacon(beaconBlock *types.BeaconBlock, rewardStateDB *statedb.StateDB) error {
	for _, inst := range beaconBlock.Body.Instructions {
		if len(inst) <= 2 {
			continue
		}
		if inst[0] == instruction.SET_ACTION || inst[0] == instruction.STAKE_ACTION || inst[0] == instruction.RANDOM_ACTION || inst[0] == instruction.SWAP_ACTION || inst[0] == instruction.ASSIGN_ACTION {
			continue
		}
		if inst[0] == instruction.ACCEPT_BLOCK_REWARD_ACTION {
			acceptBlockRewardIns, err := instruction.ValidateAndImportAcceptBlockRewardInstructionFromString(inst)
			if err != nil {
				return err
			}
			acceptBlockRewardIns.TxsFee()[common.PRVCoinID] += blockchain.getRewardAmount(acceptBlockRewardIns.ShardBlockHeight())

			for key, value := range acceptBlockRewardIns.TxsFee() {
				if value != 0 {
					err = statedb.AddShardRewardRequestV2(
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
		metaType, err := strconv.Atoi(inst[0])
		if err != nil {
			continue
		}
		switch metaType {
		case metadata.AcceptedBlockRewardInfoMeta:
			acceptedBlkRewardInfo, err := metadata.NewAcceptedBlockRewardInfoFromStr(inst[2])
			if err != nil {
				return err
			}

			if acceptedBlkRewardInfo.TxsFee == nil {
				acceptedBlkRewardInfo.TxsFee = map[common.Hash]uint64{}
			}
			acceptedBlkRewardInfo.TxsFee[common.PRVCoinID] += blockchain.getRewardAmount(acceptedBlkRewardInfo.ShardBlockHeight)

			for key, value := range acceptedBlkRewardInfo.TxsFee {
				if value != 0 {
					err = statedb.AddShardRewardRequest(rewardStateDB, beaconBlock.Header.Epoch, acceptedBlkRewardInfo.ShardID, key, value)
					if err != nil {
						return err
					}
				}
			}
			continue
		}
	}
	return nil
}

func (blockchain *BlockChain) processSalaryInstructions(rewardStateDB *statedb.StateDB, beaconBlocks []*types.BeaconBlock, shardID byte) error {
	cInfos := make(map[int][]*statedb.StakerInfo)
	isInit := false
	epoch := uint64(0)
	for _, beaconBlock := range beaconBlocks {
		for _, l := range beaconBlock.Body.Instructions {
			if l[0] == instruction.STAKE_ACTION || l[0] == instruction.RANDOM_ACTION || l[0] == instruction.FINISH_SYNC_ACTION {
				continue
			}
			if len(l) <= 2 {
				continue
			}
			shardToProcess, err := strconv.Atoi(l[1])
			if err != nil {
				continue
			}
			metaType, err := strconv.Atoi(l[0])
			if err != nil {
				return NewBlockChainError(ProcessSalaryInstructionsError, err)
			}
			if shardToProcess == int(shardID) {
				switch metaType {
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
					keyWalletDevAccount, err := wallet.Base58CheckDeserialize(blockchain.config.ChainParams.IncognitoDAOAddress)
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
			switch metaType {
			case metadata.ShardBlockRewardRequestMeta:
				shardRewardInfo, err := metadata.NewShardBlockRewardInfoFromString(l[3])
				if err != nil {
					return NewBlockChainError(ProcessSalaryInstructionsError, err)
				}
				if (!isInit) || (epoch != shardRewardInfo.Epoch) {
					isInit = true
					height := blockchain.GetLastBeaconHeightInEpoch(shardRewardInfo.Epoch)
					var beaconConsensusRootHash common.Hash
					beaconConsensusRootHash, err = blockchain.GetBeaconConsensusRootHash(blockchain.GetBeaconBestState(), height)
					if err != nil {
						return NewBlockChainError(ProcessSalaryInstructionsError, fmt.Errorf("Beacon Consensus Root Hash of Height %+v not found ,error %+v", height, err))
					}
					beaconConsensusStateDB, err := statedb.NewWithPrefixTrie(beaconConsensusRootHash, statedb.NewDatabaseAccessWarper(blockchain.GetBeaconChainDatabase()))
					if err != nil {
						return NewBlockChainError(ProcessSalaryInstructionsError, err)
					}
					cInfos = statedb.GetAllCommitteeStakeInfo(beaconConsensusStateDB, blockchain.GetShardIDs())
				}
				err = blockchain.addShardCommitteeRewardV2(rewardStateDB, shardID, shardRewardInfo, cInfos[int(shardToProcess)])
				if err != nil {
					return err
				}
				continue
			}

		}
	}
	return nil
}

func (blockchain *BlockChain) addShardCommitteeRewardV2(
	rewardStateDB *statedb.StateDB,
	shardID byte,
	rewardInfoShardToProcess *metadata.ShardBlockRewardInfo,
	cStakeInfos []*statedb.StakerInfo,
) (
	err error,
) {
	committeeSize := len(cStakeInfos)
	for _, candidate := range cStakeInfos {
		if common.GetShardIDFromLastByte(candidate.RewardReceiver().Pk[common.PublicKeySize-1]) == shardID {
			for key, value := range rewardInfoShardToProcess.ShardReward {
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

func calculateReward(
	maxBeaconBlockCreation uint64,
	splitRewardRuleProcessor committeestate.SplitRewardRuleProcessor,
	numberOfActiveShards int,
	beaconHeight uint64,
	epoch uint64,
	rewardStateDB *statedb.StateDB,
	isSplitRewardForCustodian bool,
	percentCustodianRewards uint64,
	version int,
) (map[common.Hash]uint64,
	[]map[common.Hash]uint64,
	map[common.Hash]uint64,
	map[common.Hash]uint64, error,
) {
	allCoinID := statedb.GetAllTokenIDForReward(rewardStateDB, epoch)
	blocksPerYear := getNoBlkPerYear(maxBeaconBlockCreation)
	percentForIncognitoDAO := getPercentForIncognitoDAO(beaconHeight, blocksPerYear)
	totalRewardForShard := make([]map[common.Hash]uint64, numberOfActiveShards)
	totalRewards := make([]map[common.Hash]uint64, numberOfActiveShards)
	totalRewardForBeacon := map[common.Hash]uint64{}
	totalRewardForIncDAO := map[common.Hash]uint64{}
	totalRewardForCustodian := map[common.Hash]uint64{}
	var err error

	for id := 0; id < numberOfActiveShards; id++ {
		if totalRewards[id] == nil {
			totalRewards[id] = map[common.Hash]uint64{}
		}
		if totalRewardForShard[id] == nil {
			totalRewardForShard[id] = map[common.Hash]uint64{}
		}

		for _, coinID := range allCoinID {
			switch version {
			case 1:
				totalRewards[id][coinID], err = statedb.GetRewardOfShardByEpoch(rewardStateDB, epoch, byte(id), coinID)
			case 2:
				for subsetID := 0; subsetID < MaxSubsetCommittees; subsetID++ {
					totalRewards[id][coinID], err = statedb.GetRewardOfShardByEpochV2(
						rewardStateDB, epoch,
						byte(id), byte(subsetID), coinID,
					)
				}
			}
			if err != nil {
				return nil, nil, nil, nil, err
			}
			if totalRewards[id][coinID] == 0 {
				delete(totalRewards[id], coinID)
			}
		}

		env := committeestate.NewSplitRewardEnvironment(
			byte(id),
			beaconHeight,
			totalRewards[id],
			isSplitRewardForCustodian,
			percentCustodianRewards,
			percentForIncognitoDAO,
			numberOfActiveShards,
		)
		rewardForBeacon, rewardForShard, rewardForDAO, rewardForCustodian, err := splitRewardRuleProcessor.SplitReward(env)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		plusMap(rewardForBeacon, totalRewardForBeacon)
		plusMap(rewardForShard, totalRewardForShard[id])
		plusMap(rewardForDAO, totalRewardForIncDAO)
		plusMap(rewardForCustodian, totalRewardForCustodian)
	}

	return totalRewardForBeacon, totalRewardForShard, totalRewardForIncDAO, totalRewardForCustodian, nil
}

func (blockchain *BlockChain) buildRewardInstructionByEpoch(
	curView *BeaconBestState,
	blkHeight, epoch uint64,
	isSplitRewardForCustodian bool,
	percentCustodianRewards uint64,
) ([][]string, map[common.Hash]uint64, error) {

	//Decalre variables
	var resInst [][]string
	var err error
	var instRewardForBeacons [][]string
	var instRewardForIncDAO [][]string
	var instRewardForShards [][]string

	version := 1

	totalRewardForBeacon,
		totalRewardForShard,
		totalRewardForIncDAO,
		totalRewardForCustodian,
		err := calculateReward(uint64(blockchain.config.ChainParams.MaxBeaconBlockCreation.Seconds()),
		curView.beaconCommitteeState.(committeestate.SplitRewardRuleProcessor),
		curView.ActiveShards, blkHeight, epoch,
		curView.GetBeaconRewardStateDB(),
		isSplitRewardForCustodian, percentCustodianRewards, version,
	)
	if err != nil {
		return nil, nil, err
	}

	if len(totalRewardForBeacon) > 0 {
		instRewardForBeacons, err = curView.buildInstRewardForBeacons(epoch, totalRewardForBeacon)
		if err != nil {
			return nil, nil, err
		}
	}

	instRewardForShards, err = blockchain.buildInstRewardForShards(epoch, totalRewardForShard)
	if err != nil {
		return nil, nil, err
	}

	if len(totalRewardForIncDAO) > 0 {
		instRewardForIncDAO, err = blockchain.buildInstRewardForIncDAO(epoch, totalRewardForIncDAO)
		if err != nil {
			return nil, nil, err
		}
	}

	resInst = common.AppendSliceString(instRewardForBeacons, instRewardForIncDAO, instRewardForShards)
	return resInst, totalRewardForCustodian, nil
}

//buildInstRewardForBeacons create reward instruction for beacons
func (beaconBestState *BeaconBestState) buildInstRewardForBeacons(epoch uint64, totalReward map[common.Hash]uint64) ([][]string, error) {
	resInst := [][]string{}
	baseRewards := map[common.Hash]uint64{}
	for key, value := range totalReward {
		baseRewards[key] = value / uint64(len(beaconBestState.GetBeaconCommittee()))
	}
	for _, beaconpublickey := range beaconBestState.GetBeaconCommittee() {
		// indicate reward pubkey
		singleInst, err := metadata.BuildInstForBeaconReward(baseRewards, beaconpublickey.GetNormalKey())
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
	devRewardInst, err := metadata.BuildInstForIncDAOReward(totalReward, blockchain.config.ChainParams.IncognitoDAOAddress)
	if err != nil {
		Logger.log.Errorf("buildInstRewardForIncDAO error %+v\n Totalreward: %+v, epoch: %+v\n", err, totalReward, epoch)
		return nil, err
	}
	resInst = append(resInst, devRewardInst)
	return resInst, nil
}

func (blockchain *BlockChain) buildInstRewardForShards(epoch uint64, totalRewards []map[common.Hash]uint64) ([][]string, error) {
	resInst := [][]string{}
	for i, reward := range totalRewards {
		if len(reward) > 0 {
			shardRewardInst, err := metadata.BuildInstForShardReward(reward, epoch, byte(i))
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
	amount, err := statedb.GetCommitteeReward(blockchain.GetBestStateShardRewardStateDB(shardID), tempPublicKey, requestDetail.TokenID)
	if (amount == 0) || (err != nil) {
		return nil, errors.New("Not enough reward")
	}
	responseMeta, err := metadata.NewWithDrawRewardResponse(requestDetail, (*txRequest).Hash())
	if err != nil {
		return nil, err
	}
	return blockchain.InitTxSalaryByCoinID(
		&requestDetail.PaymentAddress,
		amount,
		blkProducerPrivateKey,
		view.GetCopiedTransactionStateDB(),
		blockchain.GetBeaconBestState().GetBeaconFeatureStateDB(),
		responseMeta,
		requestDetail.TokenID,
		common.GetShardIDFromLastByte(requestDetail.PaymentAddress.Pk[common.PublicKeySize-1]))
}

func (blockchain *BlockChain) getRewardAmount(blkHeight uint64) uint64 {
	blockBeaconInterval := blockchain.config.ChainParams.MinBeaconBlockInterval.Seconds()
	blockInYear := getNoBlkPerYear(uint64(blockBeaconInterval))
	n := (blkHeight - 1) / blockInYear
	reward := uint64(blockchain.config.ChainParams.BasicReward)
	for ; n > 0; n-- {
		reward *= 91
		reward /= 100
	}
	return reward
}

func getNoBlkPerYear(blockCreationTimeSeconds uint64) uint64 {
	return (365.25 * 24 * 60 * 60) / blockCreationTimeSeconds
}

func getPercentForIncognitoDAO(blockHeight, blkPerYear uint64) int {
	year := (blockHeight - 1) / blkPerYear
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
