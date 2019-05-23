package blockchain

import (
	"encoding/json"
	"fmt"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/common/base58"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
	"github.com/constant-money/constant-chain/wallet"
	"github.com/pkg/errors"
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

	blockHash, index, err := blockgen.chain.config.DataBase.GetTransactionIndexById(txHash)
	if err != nil {
		abc := NewBlockChainError(UnExpectedError, err)
		Logger.log.Error(abc)
		return nil, abc
	}
	block, err1, _ := blockgen.chain.GetShardBlockByHash(blockHash)
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
		blockgen.chain.GetDatabase(),
		returnStakingMeta,
	)
	//modify the type of the salary transaction
	returnStakingTx.Type = common.TxReturnStakingType

	if err1 != nil {
		return nil, err1
	}
	return returnStakingTx, nil
}

func (blockgen *BlkTmplGenerator) buildBeaconSalaryRes(
	instType string,
	contentStr string,
	blkProducerPrivateKey *privacy.PrivateKey,
) ([]metadata.Transaction, error) {

	var beaconSalaryInfo metadata.BeaconSalaryInfo
	err := json.Unmarshal([]byte(contentStr), &beaconSalaryInfo)
	if err != nil {
		return nil, err
	}
	if beaconSalaryInfo.PayToAddress == nil || beaconSalaryInfo.InfoHash == nil {
		return nil, errors.Errorf("Can not Parse from contentStr")
	}

	salaryResMeta := metadata.NewBeaconBlockSalaryRes(
		beaconSalaryInfo.BeaconBlockHeight,
		beaconSalaryInfo.PayToAddress,
		beaconSalaryInfo.InfoHash,
		metadata.BeaconSalaryResponseMeta,
	)

	salaryResTx := new(transaction.Tx)
	err = salaryResTx.InitTxSalary(
		beaconSalaryInfo.BeaconSalary,
		beaconSalaryInfo.PayToAddress,
		blkProducerPrivateKey,
		blockgen.chain.GetDatabase(),
		salaryResMeta,
	)
	//fmt.Println("SA: beacon salary", beaconSalaryInfo, salaryResTx.CalculateTxValue(), salaryResTx.Hash().String())
	if err != nil {
		return nil, err
	}
	return []metadata.Transaction{salaryResTx}, nil
}

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
		reward /= 2
	}
	return reward
}
