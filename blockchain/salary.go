package blockchain

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

const (
	DurationHalfLifeRewardForDev = uint64(31536000) // 5 years, after 5 year, reward for devs = 0
)

// func getMiningReward(isForBeacon bool, blkHeight uint64) uint64 {
// 	return 0
// }

func (blockgen *BlkTmplGenerator) buildReturnStakingAmountTx(
	swaperPubKey string,
	blkProducerPrivateKey *privacy.PrivateKey,
) (metadata.Transaction, error) {
	addressBytes := blockgen.chain.config.UserKeySet.PaymentAddress.Pk
	//shardID := common.GetShardIDFromLastByte(addressBytes[len(addressBytes)-1])
	_, committeeShardID := blockgen.chain.BestState.Beacon.GetPubkeyRole(base58.Base58Check{}.Encode(addressBytes, 0x00), 0)

	fmt.Println("SA: get tx for ", swaperPubKey, GetBestStateShard(committeeShardID).StakingTx, committeeShardID)
	tx, ok := GetBestStateShard(committeeShardID).StakingTx[swaperPubKey]
	if !ok {
		return nil, NewBlockChainError(UnExpectedError, errors.New("No staking tx in best state"))
	}

	var txHash = &common.Hash{}
	(&common.Hash{}).Decode(txHash, tx)

	blockHash, index, err := blockgen.chain.config.DataBase.GetTransactionIndexById(*txHash)
	if err != nil {
		abc := NewBlockChainError(UnExpectedError, err)
		Logger.log.Error(abc)
		return nil, abc
	}
	block, _, err1 := blockgen.chain.GetShardBlockByHash(blockHash)
	if err1 != nil {
		Logger.log.Errorf("ERROR", err1, "NO Transaction in block with hash &+v", blockHash, "and index", index, "contains", block.Body.Transactions[index])
		return nil, NewBlockChainError(UnExpectedError, err1)
	}

	txData := block.Body.Transactions[index]

	keyWallet, _ := wallet.Base58CheckDeserialize(txData.GetMetadata().(*metadata.StakingMetadata).PaymentAddress)
	paymentShardID := common.GetShardIDFromLastByte(keyWallet.KeySet.PaymentAddress.Pk[len(keyWallet.KeySet.PaymentAddress.Pk)-1])

	fmt.Println("SA: build salary tx", txData.GetMetadata().(*metadata.StakingMetadata).PaymentAddress, paymentShardID, committeeShardID)

	if paymentShardID != committeeShardID {
		return nil, NewBlockChainError(UnExpectedError, errors.New("Not from this shard"))
	}

	returnStakingMeta := metadata.NewReturnStaking(
		tx,
		keyWallet.KeySet.PaymentAddress,
		metadata.ReturnStakingMeta,
	)

	returnStakingTx := new(transaction.Tx)
	err1 = returnStakingTx.InitTxSalary(
		txData.CalculateTxValue(),
		&keyWallet.KeySet.PaymentAddress,
		blkProducerPrivateKey,
		blockgen.chain.config.DataBase,
		returnStakingMeta,
	)
	//modify the type of the salary transaction
	returnStakingTx.Type = common.TxReturnStakingType

	if err1 != nil {
		return nil, err1
	}
	return returnStakingTx, nil
}

// func (blockgen *BlkTmplGenerator) buildBeaconSalaryRes(
// 	instType string,
// 	contentStr string,
// 	blkProducerPrivateKey *privacy.PrivateKey,
// ) ([]metadata.Transaction, error) {

// 	var beaconSalaryInfo metadata.BeaconSalaryInfo
// 	err := json.Unmarshal([]byte(contentStr), &beaconSalaryInfo)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if beaconSalaryInfo.PayToAddress == nil || beaconSalaryInfo.InfoHash == nil {
// 		return nil, errors.Errorf("Can not Parse from contentStr")
// 	}

// 	salaryResMeta := metadata.NewBeaconBlockSalaryRes(
// 		beaconSalaryInfo.BeaconBlockHeight,
// 		beaconSalaryInfo.PayToAddress,
// 		beaconSalaryInfo.InfoHash,
// 		metadata.BeaconSalaryResponseMeta,
// 	)

// 	salaryResTx := new(transaction.Tx)
// 	err = salaryResTx.InitTxSalary(
// 		beaconSalaryInfo.BeaconSalary,
// 		beaconSalaryInfo.PayToAddress,
// 		blkProducerPrivateKey,
// 		blockgen.chain.config.DataBase,
// 		salaryResMeta,
// 	)
// 	//fmt.Println("SA: beacon salary", beaconSalaryInfo, salaryResTx.CalculateTxValue(), salaryResTx.Hash().String())
// 	if err != nil {
// 		return nil, err
// 	}
// 	return []metadata.Transaction{salaryResTx}, nil
// }

// func (blockgen *BlkTmplGenerator) buildBeaconRewardTx(inst metadata.BeaconSalaryInfo, producerPriKey *privacy.PrivateKey) error {
// 	n := inst.BeaconBlockHeight / blockgen.chain.config.ChainParams.RewardHalflife
// 	reward := blockgen.chain.config.ChainParams.BasicReward
// 	for ; n > 0; n-- {
// 		reward /= 2
// 	}
// 	txCoinBase := new(transaction.Tx)
// 	// err := txCoinBase.InitTxSalary(reward, inst.PayToAddress, producerPriKey, db)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

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
	numberOfActiveShards := blockchain.BestState.Beacon.ActiveShards
	totalRewards := make([]uint64, numberOfActiveShards)
	totalRewardForBeacon := uint64(0)
	totalRewardForDev := uint64(0)
	var err error
	epochEndDevReward := DurationHalfLifeRewardForDev / common.EPOCH
	for ID := 0; ID < numberOfActiveShards; ID++ {
		totalRewards[ID], err = blockchain.GetDatabase().GetRewardOfShardByEpoch(epoch, byte(ID))
		if err != nil {
			return nil, err
		}
		if totalRewards[ID] == 0 {
			continue
		}
	}
	rewardForBeacon := uint64(0)
	for ID := 0; ID < numberOfActiveShards; ID++ {
		if epoch <= epochEndDevReward {
			rewardForBeacon = uint64(18*totalRewards[ID]) / ((uint64(numberOfActiveShards) + 2) * 10)
			fmt.Printf("[ndh] shardID %+v - - %+v %+v %+v\n", totalRewards[ID], rewardForBeacon, 18*totalRewards[ID], ((uint64(numberOfActiveShards) + 2) * 10))
			rewardForDev := totalRewards[ID] / 10
			totalRewards[ID] -= (rewardForBeacon + rewardForDev)
			totalRewardForDev += rewardForDev
		} else {
			rewardForBeacon = 2 * totalRewards[ID] / (uint64(numberOfActiveShards) + 2)
			totalRewards[ID] -= (rewardForBeacon)
		}
		totalRewardForBeacon += rewardForBeacon
	}
	fmt.Printf("[ndh] %+v\n", totalRewardForBeacon)
	var resInst [][]string
	var instRewardForBeacons [][]string
	var instRewardForDev [][]string
	var instRewardForShards [][]string
	if totalRewardForBeacon > 0 {
		instRewardForBeacons, err = blockchain.BuildInstRewardForBeacons(epoch, totalRewardForBeacon)
		if err != nil {
			return nil, err
		}
	}
	if totalRewardForDev > 0 {
		instRewardForDev, err = blockchain.BuildInstRewardForDev(epoch, totalRewardForDev)
		if err != nil {
			return nil, err
		}
	}
	instRewardForShards, err = blockchain.BuildInstRewardForShards(epoch, totalRewards)
	if err != nil {
		return nil, err
	}
	resInst = common.AppendSliceString(instRewardForBeacons, instRewardForDev, instRewardForShards)
	return resInst, nil
}

func (blockchain *BlockChain) shareRewardForShardCommittee(epoch, totalReward uint64, listCommitee []string) error {
	reward := totalReward / uint64(len(listCommitee))
	for i, committee := range listCommitee {
		committeeBytes, _, err := base58.Base58Check{}.Decode(committee)
		if err != nil {
			for j := i - 1; j >= 0; j-- {
				committeeBytes, _, _ := base58.Base58Check{}.Decode(listCommitee[j])
				_ = blockchain.config.DataBase.RemoveCommitteeReward(committeeBytes, reward)
			}
			return err
		}
		err = blockchain.config.DataBase.AddCommitteeReward(committeeBytes, reward)
		if err != nil {
			for j := i - 1; j >= 0; j-- {
				committeeBytes, _, _ := base58.Base58Check{}.Decode(listCommitee[j])
				_ = blockchain.config.DataBase.RemoveCommitteeReward(committeeBytes, reward)
			}
			return err
		}
	}
	return nil
}

func (blockchain *BlockChain) updateDatabaseFromBeaconInstructions(
	beaconBlocks []*BeaconBlock,
	shardID byte,
) error {

	shardCommittee := make(map[byte][]string)
	isInit := false
	epoch := uint64(0)
	db := blockchain.config.DataBase
	// listShardCommittee := blockchain.config.DataBase.FetchCommitteeByEpoch
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
			if shardToProcess == int(shardID) {
				metaType, err := strconv.Atoi(l[0])
				if err != nil {
					return err
				}
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
					err = db.AddCommitteeReward(publicKeyCommittee, beaconBlkRewardInfo.BeaconReward)
					if err != nil {
						return err
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
					err = db.AddCommitteeReward(keyWalletDevAccount.KeySet.PaymentAddress.Pk, devRewardInfo.DevReward)
					if err != nil {
						return err
					}
					continue

				case metadata.ShardBlockRewardRequestMeta:
					shardRewardInfo, err := metadata.NewShardBlockRewardInfoFromString(l[3])
					if err != nil {
						return err
					}
					if (!isInit) || (epoch != shardRewardInfo.Epoch) {
						isInit = true
						epoch = shardRewardInfo.Epoch
						temp, err := blockchain.config.DataBase.FetchCommitteeByEpoch(epoch)
						if err != nil {
							return err
						}
						json.Unmarshal(temp, &shardCommittee)
					}
					err = blockchain.shareRewardForShardCommittee(shardRewardInfo.Epoch, shardRewardInfo.ShardReward, shardCommittee[shardID])
					if err != nil {
						return err
					}
					continue
				}
			}
		}
	}
	return nil
}

func (blockchain *BlockChain) updateDatabaseFromBeaconBlock(
	beaconBlock *BeaconBlock,
) error {
	db := blockchain.config.DataBase
	for _, inst := range beaconBlock.Body.Instructions {
		if inst[0] == StakeAction || inst[0] == RandomAction {
			continue
		}
		if len(inst) <= 2 {
			continue
		}
		metaType, err := strconv.Atoi(inst[0])
		if err != nil {
			fmt.Printf("[ndh] error - - %+v\n", err)
			return err
		}
		switch metaType {
		case metadata.AcceptedBlockRewardInfoMeta:
			acceptedBlkRewardInfo, err := metadata.NewAcceptedBlockRewardInfoFromStr(inst[2])
			if err != nil {
				fmt.Printf("[ndh] error1 - - %+v\n", err)
				return err
			}
			totalReward := blockchain.getRewardAmount(acceptedBlkRewardInfo.ShardBlockHeight) + acceptedBlkRewardInfo.TxsFee
			err = db.AddShardRewardRequest(beaconBlock.Header.Epoch, acceptedBlkRewardInfo.ShardID, totalReward)
			if err != nil {
				return err
			}
			continue
		}
	}
	fmt.Printf("[ndh] non error \n")
	return nil
}

func (blockchain *BlockChain) buildWithDrawTransactionResponse(txRequest *metadata.Transaction, blkProducerPrivateKey *privacy.PrivateKey) (metadata.Transaction, error) {
	if (*txRequest).GetMetadataType() != metadata.WithDrawRewardRequestMeta {
		return nil, errors.New("Can not understand this request!")
	}
	// requestMeta := (*txRequest).GetMetadata().(*metadata.WithDrawRewardRequest)
	// requester := base58.Base58Check{}.Encode(requestMeta.PaymentAddress.Pk, VERSION)
	// receiverBytes := requestMeta.PaymentAddress.Pk
	requestDetail := (*txRequest).GetMetadata().(*metadata.WithDrawRewardRequest)
	// requester := base58.Base58Check{}.Encode(requestMeta.PaymentAddress.Pk, VERSION)
	// if len(receiverBytes) == 0 {
	// 	return nil, errors.New("Can not get payment address of request's sender")
	// }
	amount, err := blockchain.config.DataBase.GetCommitteeReward(requestDetail.PaymentAddress.Pk)
	if (amount == 0) || (err != nil) {
		return nil, errors.New("Not enough reward")
	}
	responseMeta, err := metadata.NewWithDrawRewardResponse((*txRequest).Hash())
	if err != nil {
		return nil, err
	}
	txRes := new(transaction.Tx)
	txRes.InitTxSalary(amount, &requestDetail.PaymentAddress, blkProducerPrivateKey, blockchain.config.DataBase, responseMeta)
	return txRes, nil
}
