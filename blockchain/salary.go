package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/core/rawdb"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incdb"
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
	blockHash, index, err := rawdb.GetTransactionIndexById(blockGenerator.chain.config.DataBase, *txHash)
	if err != nil {
		return nil, NewBlockChainError(GetTransactionFromDatabaseError, err)
	}
	shardBlock, _, err := blockGenerator.chain.GetShardBlockByHash(blockHash)
	if err != nil || shardBlock == nil {
		Logger.log.Error("ERROR", err, "NO Transaction in block with hash", blockHash, "and index", index, "contains", shardBlock.Body.Transactions[index])
		return nil, NewBlockChainError(FetchShardBlockError, err)
	}
	txData := shardBlock.Body.Transactions[index]
	keyWallet, err := wallet.Base58CheckDeserialize(txData.GetMetadata().(*metadata.StakingMetadata).FunderPaymentAddress)
	if err != nil {
		Logger.log.Error("SA: cannot get payment address", txData.GetMetadata().(*metadata.StakingMetadata), committeeShardID)
		return nil, NewBlockChainError(WalletKeySerializedError, err)
	}
	Logger.log.Info("SA: build salary tx", txData.GetMetadata().(*metadata.StakingMetadata).FunderPaymentAddress, committeeShardID)
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
	blockBeaconInterval := blockchain.config.ChainParams.MinBeaconBlockInterval.Seconds()
	blockInYear := getNoBlkPerYear(uint64(blockBeaconInterval))
	n := blkHeight / blockInYear
	reward := uint64(blockchain.config.ChainParams.BasicReward)
	for ; n > 0; n-- {
		reward *= 91
		reward /= 100
	}
	return reward
}

func (blockchain *BlockChain) BuildRewardInstructionByEpoch(blkHeight, epoch uint64) ([][]string, error) {
	var resInst [][]string
	var instRewardForBeacons [][]string
	var instRewardForIncDAO [][]string
	var instRewardForShards [][]string
	numberOfActiveShards := blockchain.BestState.Beacon.ActiveShards
	allCoinID, err := blockchain.GetAllCoinID()
	if err != nil {
		return nil, err
	}
	blkPerYear := getNoBlkPerYear(uint64(blockchain.config.ChainParams.MaxBeaconBlockCreation.Seconds()))
	percentForIncognitoDAO := getPercentForIncognitoDAO(blkHeight, blkPerYear)
	totalRewards := make([]map[common.Hash]uint64, numberOfActiveShards)
	totalRewardForBeacon := map[common.Hash]uint64{}
	totalRewardForIncDAO := map[common.Hash]uint64{}
	for ID := 0; ID < numberOfActiveShards; ID++ {
		if totalRewards[ID] == nil {
			totalRewards[ID] = map[common.Hash]uint64{}
		}
		for _, coinID := range allCoinID {
			totalRewards[ID][coinID], err = rawdb.GetRewardOfShardByEpoch(blockchain.GetDatabase(), epoch, byte(ID), coinID)
			if err != nil {
				return nil, err
			}
			if totalRewards[ID][coinID] == 0 {
				delete(totalRewards[ID], coinID)
			}
		}
		rewardForBeacon, rewardForIncDAO, err := splitReward(&totalRewards[ID], numberOfActiveShards, percentForIncognitoDAO)
		if err != nil {
			Logger.log.Infof("\n------------------------------------\nNot enough reward in epoch %v\n------------------------------------\n", err)
		}
		mapPlusMap(rewardForBeacon, &totalRewardForBeacon)
		mapPlusMap(rewardForIncDAO, &totalRewardForIncDAO)
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

	if len(totalRewardForIncDAO) > 0 {
		instRewardForIncDAO, err = blockchain.BuildInstRewardForIncDAO(epoch, totalRewardForIncDAO)
		if err != nil {
			return nil, err
		}
	}
	resInst = common.AppendSliceString(instRewardForBeacons, instRewardForIncDAO, instRewardForShards)
	return resInst, nil
}

func (blockchain *BlockChain) updateDatabaseFromBeaconInstructions(beaconBlocks []*BeaconBlock, shardID byte) error {
	rewardReceivers := make(map[string]string)
	committee := make(map[byte][]incognitokey.CommitteePublicKey)
	isInit := false
	epoch := uint64(0)
	for _, beaconBlock := range beaconBlocks {
		//fmt.Printf("RewardLog Process BeaconBlock %v\n", beaconBlock.GetHeight())
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
					// fmt.Printf("RewardLog Process Beacon %v\n", l)
					beaconBlkRewardInfo, err := metadata.NewBeaconBlockRewardInfoFromStr(l[3])
					if err != nil {
						return err
					}
					publicKeyCommittee, _, err := base58.Base58Check{}.Decode(beaconBlkRewardInfo.PayToPublicKey)
					if err != nil {
						return err
					}
					for key := range beaconBlkRewardInfo.BeaconReward {
						err = rawdb.AddCommitteeReward(blockchain.GetDatabase(), publicKeyCommittee, beaconBlkRewardInfo.BeaconReward[key], key)
						if err != nil {
							return err
						}
					}
					continue

				case metadata.IncDAORewardRequestMeta:
					fmt.Printf("RewardLog Process Dev %v\n", l)
					incDAORewardInfo, err := metadata.NewIncDAORewardInfoFromStr(l[3])
					if err != nil {
						return err
					}
					keyWalletDevAccount, err := wallet.Base58CheckDeserialize(blockchain.config.ChainParams.IncognitoDAOAddress)
					if err != nil {
						return err
					}
					for key := range incDAORewardInfo.IncDAOReward {
						err = rawdb.AddCommitteeReward(blockchain.GetDatabase(), keyWalletDevAccount.KeySet.PaymentAddress.Pk, incDAORewardInfo.IncDAOReward[key], key)
						if err != nil {
							return err
						}
					}
					continue
				}
			}
			switch metaType {
			case metadata.ShardBlockRewardRequestMeta:
				//fmt.Printf("RewardLog Process Shard %v\n", l)
				shardRewardInfo, err := metadata.NewShardBlockRewardInfoFromString(l[3])
				if err != nil {
					return err
				}
				if (!isInit) || (epoch != shardRewardInfo.Epoch) {
					isInit = true
					epoch = shardRewardInfo.Epoch
					rewardReceiverBytes, err := rawdb.FetchRewardReceiverByHeight(blockchain.GetDatabase(), epoch*blockchain.config.ChainParams.Epoch)
					if err != nil {
						return err
					}
					json.Unmarshal(rewardReceiverBytes, &rewardReceivers)
					committeeBytes, err := rawdb.FetchShardCommitteeByHeight(blockchain.GetDatabase(), epoch*blockchain.config.ChainParams.Epoch)
					if err != nil {
						return err
					}
					json.Unmarshal(committeeBytes, &committee)
				}
				err = blockchain.getRewardAmountForUserOfShard(shardID, shardRewardInfo, committee[byte(shardToProcess)], &rewardReceivers, false)
				if err != nil {
					return err
				}
				continue
			}

		}
	}
	return nil
}

func (blockchain *BlockChain) updateDatabaseWithBlockRewardInfo(beaconBlock *BeaconBlock, bd *[]incdb.BatchData) error {
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
				if value != 0 {
					err = rawdb.AddShardRewardRequest(blockchain.GetDatabase(), beaconBlock.Header.Epoch, acceptedBlkRewardInfo.ShardID, value, key, bd)
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
	amount, err := rawdb.GetCommitteeReward(blockchain.GetDatabase(), requestDetail.PaymentAddress.Pk, requestDetail.TokenID)
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
		common.GetShardIDFromLastByte(requestDetail.PaymentAddress.Pk[common.PublicKeySize-1]))
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
	devPercent int,
) (
	*map[common.Hash]uint64,
	*map[common.Hash]uint64,
	error,
) {
	hasValue := false
	rewardForBeacon := map[common.Hash]uint64{}
	rewardForIncDAO := map[common.Hash]uint64{}
	for key, value := range *totalReward {
		rewardForBeacon[key] = 2 * (uint64(100-devPercent) * value) / ((uint64(numberOfActiveShards) + 2) * 100)
		rewardForIncDAO[key] = uint64(devPercent) * value / uint64(100)
		(*totalReward)[key] = value - (rewardForBeacon[key] + rewardForIncDAO[key])
		if !hasValue {
			hasValue = true
		}
	}
	if !hasValue {
		//fmt.Printf("[ndh] not enough reward\n")
		return nil, nil, NewBlockChainError(NotEnoughRewardError, errors.New("Not enough reward"))
	}
	return &rewardForBeacon, &rewardForIncDAO, nil
}

func getNoBlkPerYear(blockCreationTimeSeconds uint64) uint64 {
	//31536000 =
	return (365 * 24 * 60 * 60) / blockCreationTimeSeconds
}

func getPercentForIncognitoDAO(blockHeight, blkPerYear uint64) int {
	year := blockHeight / blkPerYear
	if year > (UpperBoundPercentForIncDAO - LowerBoundPercentForIncDAO) {
		return LowerBoundPercentForIncDAO
	} else {
		return UpperBoundPercentForIncDAO - int(year)
	}
}

func (blockchain *BlockChain) getRewardAmountForUserOfShard(
	selfShardID byte,
	rewardInfoShardToProcess *metadata.ShardBlockRewardInfo,
	committeeOfShardToProcess []incognitokey.CommitteePublicKey,
	rewardReceiver *map[string]string,
	forBackup bool,
) (
	err error,
) {
	committeeSize := len(committeeOfShardToProcess)
	// wg := sync.WaitGroup{}
	// done := make(chan bool, 1)
	// errChan := make(chan error, 1)
	for _, candidate := range committeeOfShardToProcess {
		// wg.Add(1)
		// go func() {
		// 	defer wg.Done()
		wl, err := wallet.Base58CheckDeserialize((*rewardReceiver)[candidate.GetIncKeyBase58()])
		if err != nil {
			// errChan <- err
			return err
		}
		if common.GetShardIDFromLastByte(wl.KeySet.PaymentAddress.Pk[common.PublicKeySize-1]) == selfShardID {
			for key, value := range rewardInfoShardToProcess.ShardReward {
				if forBackup {
					err = rawdb.BackupCommitteeReward(blockchain.GetDatabase(), wl.KeySet.PaymentAddress.Pk, key)
				} else {
					err = rawdb.AddCommitteeReward(blockchain.GetDatabase(), wl.KeySet.PaymentAddress.Pk, value/uint64(committeeSize), key)
				}
				if err != nil {
					// errChan <- err
					return err
				}
			}
		}
		// }()

	}
	// go func() {
	// 	wg.Wait()
	// 	close(done)
	// }()
	// select {
	// case <-done:
	// case err = <-errChan:
	// 	if err != nil {
	// 		return err
	// 	}
	return nil
}
