package blockchain

import (
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

func (blockGenerator *BlockGenerator) buildReturnStakingAmountTx(
	swapPublicKey string,
	blkProducerPrivateKey *privacy.PrivateKey,
) (metadata.Transaction, error) {
	// addressBytes := blockGenerator.chain.config.UserKeySet.PaymentAddress.Pk
	//shardID := common.GetShardIDFromLastByte(addressBytes[len(addressBytes)-1])
	publicKey, _ := blockGenerator.chain.config.ConsensusEngine.GetCurrentMiningPublicKey()
	_, committeeShardID := blockGenerator.chain.BestState.Beacon.GetPubkeyRole(publicKey, 0)

	fmt.Println("SA: get tx for ", swapPublicKey, GetBestStateShard(committeeShardID).StakingTx, committeeShardID)
	tx, ok := GetBestStateShard(committeeShardID).StakingTx[swapPublicKey]
	if !ok {
		return nil, NewBlockChainError(GetStakingTransactionError, errors.New("No staking tx in best state"))
	}
	var txHash = &common.Hash{}
	err := (&common.Hash{}).Decode(txHash, tx)
	if err != nil {
		return nil, NewBlockChainError(DecodeHashError, err)
	}
	blockHash, index, err := blockGenerator.chain.config.DataBase.GetTransactionIndexById(*txHash)
	if err != nil {
		return nil, NewBlockChainError(GetTransactionFromDatabaseError, err)
	}
	shardBlock, _, err := blockGenerator.chain.GetShardBlockByHash(blockHash)
	if err != nil || shardBlock == nil {
		Logger.log.Error("ERROR", err, "NO Transaction in block with hash", blockHash, "and index", index, "contains", shardBlock.Body.Transactions[index])
		return nil, NewBlockChainError(FetchShardBlockError, err)
	}
	txData := shardBlock.Body.Transactions[index]
	keyWallet, err := wallet.Base58CheckDeserialize(txData.GetMetadata().(*metadata.StakingMetadata).RewardReceiverPaymentAddress)
	if err != nil {
		Logger.log.Error("SA: cannot get payment address", txData.GetMetadata().(*metadata.StakingMetadata), committeeShardID)
		return nil, NewBlockChainError(WalletKeySerializedError, err)
	}
	Logger.log.Info("SA: build salary tx", txData.GetMetadata().(*metadata.StakingMetadata).RewardReceiverPaymentAddress, committeeShardID)
	paymentShardID := common.GetShardIDFromLastByte(keyWallet.KeySet.PaymentAddress.Pk[len(keyWallet.KeySet.PaymentAddress.Pk)-1])
	if paymentShardID != committeeShardID {
		return nil, NewBlockChainError(WrongShardIDError, fmt.Errorf("Staking Payment Address ShardID %+v, Not From Current Shard %+v", paymentShardID, committeeShardID))
	}
	returnStakingMeta := metadata.NewReturnStaking(
		tx,
		keyWallet.KeySet.PaymentAddress,
		metadata.ReturnStakingMeta,
	)
	returnStakingTx := new(transaction.Tx)
	err = returnStakingTx.InitTxSalary(
		txData.CalculateTxValue(),
		&keyWallet.KeySet.PaymentAddress,
		blkProducerPrivateKey,
		blockGenerator.chain.config.DataBase,
		returnStakingMeta,
	)
	//modify the type of the salary transaction
	returnStakingTx.Type = common.TxReturnStakingType
	if err != nil {
		return nil, NewBlockChainError(InitSalaryTransactionError, err)
	}
	return returnStakingTx, nil
}

func (blockchain *BlockChain) getRewardAmount(blkHeight uint64) uint64 {
	n := blkHeight / blockchain.config.ChainParams.RewardHalflife
	reward := uint64(blockchain.config.ChainParams.BasicReward)
	for ; n > 0; n-- {
		reward *= 9
		reward /= 10
	}
	return reward
}

func (blockchain *BlockChain) BuildRewardInstructionByEpoch(epoch uint64) ([][]string, error) {
	var resInst [][]string
	var instRewardForBeacons [][]string
	var instRewardForDev [][]string
	var instRewardForShards [][]string
	numberOfActiveShards := blockchain.BestState.Beacon.ActiveShards
	allCoinID, err := blockchain.GetAllCoinID()
	if err != nil {
		return nil, err
	}
	epochEndDevReward := DurationHalfLifeRewardForDev / blockchain.config.ChainParams.Epoch
	forDev := epochEndDevReward >= epoch
	totalRewards := make([]map[common.Hash]uint64, numberOfActiveShards)
	totalRewardForBeacon := map[common.Hash]uint64{}
	totalRewardForDev := map[common.Hash]uint64{}
	for ID := 0; ID < numberOfActiveShards; ID++ {
		if totalRewards[ID] == nil {
			totalRewards[ID] = map[common.Hash]uint64{}
		}
		for _, coinID := range allCoinID {
			totalRewards[ID][coinID], err = blockchain.GetDatabase().GetRewardOfShardByEpoch(epoch, byte(ID), coinID)
			if err != nil {
				return nil, err
			}
			if totalRewards[ID][coinID] == 0 {
				delete(totalRewards[ID], coinID)
			}
		}
		rewardForBeacon, rewardForDev, err := splitReward(&totalRewards[ID], numberOfActiveShards, forDev)
		if err != nil {
			return nil, err
		}
		mapPlusMap(rewardForBeacon, &totalRewardForBeacon)

		if forDev {
			mapPlusMap(rewardForDev, &totalRewardForDev)
		}
	}
	if len(totalRewardForBeacon) > 0 {
		instRewardForBeacons, err = blockchain.BuildInstRewardForBeacons(epoch, totalRewardForBeacon)
		if err != nil {
			return nil, err
		}
	}

	instRewardForShards, err = blockchain.BuildInstRewardForShards(epoch, totalRewards)
	if err != nil {
		return nil, err
	}

	if forDev {
		if len(totalRewardForDev) > 0 {
			instRewardForDev, err = blockchain.BuildInstRewardForDev(epoch, totalRewardForDev)
			if err != nil {
				return nil, err
			}
		}
	}

	resInst = common.AppendSliceString(instRewardForBeacons, instRewardForDev, instRewardForShards)

	return resInst, nil
}

// [committeeString][key][value]
func (blockchain *BlockChain) shareRewardForShardCommittee(epoch uint64, totalReward map[common.Hash]uint64, listCommitee []incognitokey.CommitteePublicKey) error {
	reward := map[common.Hash]uint64{}
	for key, value := range totalReward {
		reward[key] = value / uint64(len(listCommitee))
	}
	for key := range totalReward {
		for i, committee := range listCommitee {
			committeeBytes := committee.GetNormalKey()
			// if err != nil {
			// 	for j := i - 1; j >= 0; j-- {
			// 		committeeBytes, _, _ := DecodeCommitteeKey(listCommitee[j])
			// 		_ = blockchain.config.DataBase.RemoveCommitteeReward(committeeBytes, reward[key], key)
			// 	}
			// 	return err
			// }
			err := blockchain.config.DataBase.AddCommitteeReward(committeeBytes, reward[key], key)
			if err != nil {
				for j := i - 1; j >= 0; j-- {
					committeeBytes := listCommitee[j].GetNormalKey()
					_ = blockchain.config.DataBase.RemoveCommitteeReward(committeeBytes, reward[key], key)
				}
				return err
			}
		}
	}
	return nil
}

func (blockchain *BlockChain) updateDatabaseFromBeaconInstructions(beaconBlocks []*BeaconBlock, shardID byte) error {
	// shardCommittee := make(map[byte][]incognitokey.CommitteePublicKey)
	// isInit := false
	// epoch := uint64(0)
	db := blockchain.config.DataBase
	for _, beaconBlock := range beaconBlocks {
		for _, l := range beaconBlock.Body.Instructions {
			if l[0] == StakeAction || l[0] == RandomAction {
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
				return err
			}
			if shardToProcess == int(shardID) {
				switch metaType {
				case metadata.BeaconRewardRequestMeta:
					beaconBlkRewardInfo, err := metadata.NewBeaconBlockRewardInfoFromStr(l[3])
					if err != nil {
						return err
					}
					publicKeyCommittee, _, err := base58.Base58Check{}.Decode(beaconBlkRewardInfo.PayToPublicKey)
					if err != nil {
						return err
					}
					for key := range beaconBlkRewardInfo.BeaconReward {
						err = db.AddCommitteeReward(publicKeyCommittee, beaconBlkRewardInfo.BeaconReward[key], key)
						if err != nil {
							return err
						}
					}
					continue

				case metadata.DevRewardRequestMeta:
					devRewardInfo, err := metadata.NewDevRewardInfoFromStr(l[3])
					if err != nil {
						return err
					}
					keyWalletDevAccount, err := wallet.Base58CheckDeserialize(common.DevAddress)
					if err != nil {
						return err
					}
					for key := range devRewardInfo.DevReward {
						err = db.AddCommitteeReward(keyWalletDevAccount.KeySet.PaymentAddress.Pk, devRewardInfo.DevReward[key], key)
						if err != nil {
							return err
						}
					}
					continue
				}
			}
			// switch metaType {
			// case metadata.ShardBlockRewardRequestMeta:
			// 	shardRewardInfo, err := metadata.NewShardBlockRewardInfoFromString(l[3])
			// 	if err != nil {
			// 		return err
			// 	}
			// 	if (!isInit) || (epoch != shardRewardInfo.Epoch) {
			// 		isInit = true
			// 		epoch = shardRewardInfo.Epoch
			// 		temp, err := blockchain.config.DataBase.FetchRewardReceiverByHeight(epoch * blockchain.config.ChainParams.Epoch)
			// 		if err != nil {
			// 			return err
			// 		}
			// 		json.Unmarshal(temp, &shardCommittee)
			// 	}
			// 	err = blockchain.shareRewardForShardCommittee(shardRewardInfo.Epoch, shardRewardInfo.ShardReward, shardCommittee[shardID])
			// 	if err != nil {
			// 		return err
			// 	}
			// 	continue
			// }

		}
	}
	return nil
}

func (blockchain *BlockChain) updateDatabaseWithBlockRewardInfo(beaconBlock *BeaconBlock) error {
	db := blockchain.config.DataBase
	for _, inst := range beaconBlock.Body.Instructions {
		if len(inst) <= 2 {
			continue
		}
		if inst[0] == SetAction || inst[0] == StakeAction || inst[0] == RandomAction || inst[0] == SwapAction || inst[0] == AssignAction {
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
			if val, ok := acceptedBlkRewardInfo.TxsFee[common.PRVCoinID]; ok {
				acceptedBlkRewardInfo.TxsFee[common.PRVCoinID] = val + blockchain.getRewardAmount(acceptedBlkRewardInfo.ShardBlockHeight)
			} else {
				if acceptedBlkRewardInfo.TxsFee == nil {
					acceptedBlkRewardInfo.TxsFee = map[common.Hash]uint64{}
				}
				acceptedBlkRewardInfo.TxsFee[common.PRVCoinID] = blockchain.getRewardAmount(acceptedBlkRewardInfo.ShardBlockHeight)
			}
			for key, value := range acceptedBlkRewardInfo.TxsFee {
				err = db.AddShardRewardRequest(beaconBlock.Header.Epoch, acceptedBlkRewardInfo.ShardID, value, key)
				if err != nil {
					return err
				}
			}
			continue
		}
	}
	return nil
}

func (blockchain *BlockChain) buildWithDrawTransactionResponse(
	txRequest *metadata.Transaction,
	blkProducerPrivateKey *privacy.PrivateKey,
) (
	metadata.Transaction,
	error,
) {
	if (*txRequest).GetMetadataType() != metadata.WithDrawRewardRequestMeta {
		return nil, errors.New("Can not understand this request!")
	}
	requestDetail := (*txRequest).GetMetadata().(*metadata.WithDrawRewardRequest)
	amount, err := blockchain.config.DataBase.GetCommitteeReward(requestDetail.PaymentAddress.Pk, requestDetail.TokenID)
	if (amount == 0) || (err != nil) {
		return nil, errors.New("Not enough reward")
	}
	responseMeta, err := metadata.NewWithDrawRewardResponse((*txRequest).Hash())
	if err != nil {
		return nil, err
	}
	return blockchain.InitTxSalaryByCoinID(
		&requestDetail.PaymentAddress,
		amount,
		blkProducerPrivateKey,
		blockchain.GetDatabase(),
		responseMeta,
		requestDetail.TokenID,
		common.GetShardIDFromLastByte(requestDetail.PaymentAddress.Pk[32]))
}

// mapPlusMap(src, dst): dst = dst + src
func mapPlusMap(src, dst *map[common.Hash]uint64) {
	if src != nil {
		for key, value := range *src {
			(*dst)[key] += value
		}
	}
}

// calculateMapReward(src, dst): dst = dst + src
func splitReward(
	totalReward *map[common.Hash]uint64,
	numberOfActiveShards int,
	forDev bool,
) (
	*map[common.Hash]uint64,
	*map[common.Hash]uint64,
	error,
) {
	hasValue := false
	rewardForBeacon := map[common.Hash]uint64{}
	rewardForDev := map[common.Hash]uint64{}
	if forDev {
		for key, value := range *totalReward {
			rewardForBeacon[key] = uint64(18*value) / ((uint64(numberOfActiveShards) + 2) * 10)
			rewardForDev[key] = value / uint64(10)
			(*totalReward)[key] = value - (rewardForBeacon[key] + rewardForDev[key])
			//fmt.Printf("[ndh] TokenID %+v - - Beacon: %+v; Dev: %+v; Shard: %+v;\n", key, rewardForBeacon[key], rewardForDev[key], (*totalReward)[key])
			if !hasValue {
				hasValue = true
			}
		}
		if !hasValue {
			//fmt.Printf("[ndh] not enough reward\n")
			return nil, nil, nil
		}
		return &rewardForBeacon, &rewardForDev, nil
	} else {
		for key, value := range *totalReward {
			rewardForBeacon[key] = uint64(2*value) / (uint64(numberOfActiveShards) + 2)
			(*totalReward)[key] = value - (rewardForBeacon[key])
			//fmt.Printf("[ndh] TokenID %+v - - Beacon: %+v; noDev; Shard: %+v;\n", key, rewardForBeacon[key], (*totalReward)[key])
			if !hasValue {
				hasValue = true
			}
		}
		if !hasValue {
			//fmt.Printf("[ndh] not enough reward\n")
			return nil, nil, nil
		}
		return &rewardForBeacon, nil, nil
	}
}

// func getRewardAmountForUserOfShard()

// func getRewardAmountOfShardCandidate(
// 	shardReward uint64,
// 	numberOfCandidates int,
// ) (
// 	reward uint64,
// 	err error,
// ) {
// 	reward :=
// 	return
// }
