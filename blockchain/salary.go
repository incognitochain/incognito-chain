package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdb"
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

func isTestnetWrongToken(tokenID common.Hash) bool {
	testnetWrongToken := []string{
		"045b42b8a4f0e500f66d54aa327f9bd43fcb4fe3e3f2a31c519b5a7705e1d9bf",
		"058598884a6301ca4b7a36698e1d2bb40703ff41794c102a31178863310edc68",
		"0e349240130047bc844488f30855f0439148b4e6b0c5cd27cc9392de61833e4c",
		"0e990e48cc02a2fecffe8dd6b1bc12cd3d8a5c2ae493dc4b6bfb4331e08ad3fc",
		"15ffa7bc0bf626de2772977065421903de7d5e5fedddf279d791915f02c46426",
		"18b365667ec9e7f9bb7974c62f8370e20bba7e21f79adc7e5f5396b08f85f0c2",
		"18e820c5a799ce4fb369b6243bb6cd1f14a6f08279041cc6d0bfc18687602ad1",
		"3c966cbd4498d4ed5419a0613a0404e0ae326558afd3035a49c0ae37d9895538",
		"3cfb98d364f3bf715124c94e2f5856c5406b22fbe7ba7e1b4200e9f395bbb185",
		"3ea38346d39e7e1c73fb37336b6bede668639400d1e62ec1abad6fb64c438e37",
		"41b9c3dccedc241e52654dc674bd6879391935d86175a73b75256d8dbec10a9a",
		"4cf0a2fe3d12f8e25cd824a4e9e0bf83bbb3e18cb662920e24b14fbab8cf59b8",
		"4e0237fc0e1d8638aad45e8adcc2f1a24be24d765fdfa3e657394c6f9305584a",
		"683d8673f8e3b5a90ae89b269e6b286cd6729b5086563093e2525a70bbe5f1e7",
		"68aa8e11a4751ec9318042b7f7864b547f39d9e1f4cecb5769db94d082f84676",
		"69c0892139b83a29509cdafe9ec6a064a8f11fcb5045f543a1bfa467612da4a7",
		"6dee6fa8d29dc2d1c286b86a9ba99ff70a183ddf7a7339ed7dd9f471b8ce91a6",
		"71462e0981f383858ca72f759bcc07fdc48d21c9c1a4c4e415458904a1cfd99a",
		"91ec8decb876fbfd75fa5f5fbccb8e6720b5ae764160ebd8cc2d6010c7e23a4e",
		"996a0076f9ea2c74f0dc8c46a9e619a58f8709cf65ffef63bb3a85b711386d76",
		"a658b737fc82cf2d8c1e038f1a6681cb2ac5b8dc53675b4d8eee651f686eb093",
		"af0e4b1fbe7dd3c252df6b00279dfed02aa8537eaa62abaab7b3a442ce1bef38",
		"af7ecb9bfb261ce33b85f457599ff9cb2b86540e0c966fade27dbed68672466d",
		"cf0d4e267de468fe9f74f8c575fcce234d4f0e00b3f55df25e78379cfe21dcd3",
		"d8c5cc41572728355d725ec7c3078519e7a7937e7f9a81cad3fff3b221556abf",
		"e2d8849b37924167c7fb828e364c5761e2a81099b457d64df43671276f23a23e",
		"e49f4c4c33dfb78eca41d3f423f5eb39c9a9130fcb00a7c40d255badd7b40784",
		"f08312d24a12225d02b50031451818ed4c100b71977d7d8a10c2816a587f1a83",
	}
	for _, tokenString := range testnetWrongToken {
		if tokenID.String() == tokenString {
			return true
		}
	}
	return false
}

func removeTestnetWrongToken(sliceToken []common.Hash) []common.Hash {
	i := 0
	for _, tokenID := range sliceToken {
		if isTestnetWrongToken(tokenID) {
			continue
		}
		sliceToken[i] = tokenID
		i++
	}
	return sliceToken[:i]
}

func (blockchain *BlockChain) BuildRewardInstructionByEpoch(blkHeight, epoch uint64) ([][]string, error) {
	var resInst [][]string
	var instRewardForBeacons [][]string
	var instRewardForIncDAO [][]string
	var instRewardForShards [][]string
	numberOfActiveShards := blockchain.BestState.Beacon.ActiveShards
	allCoinID, err := blockchain.config.DataBase.GetAllTokenIDForReward(epoch)

	if blockchain.config.ChainParams.Net == Testnet {
		// istestnet
		allCoinID = removeTestnetWrongToken(allCoinID)
	}

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
