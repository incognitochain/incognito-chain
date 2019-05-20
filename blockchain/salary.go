package blockchain

import (
	"fmt"

	"github.com/constant-money/constant-chain/common/base58"
	"github.com/constant-money/constant-chain/wallet"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
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
