package blockchain

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/pkg/errors"
	"strconv"
)

type returnStakingBeaconInfo struct {
	FunderAddress privacy.PaymentAddress
	SharedRandom  []byte
	StakingTx     []metadata.Transaction
	StakingAmount uint64
}

func (r *returnStakingBeaconInfo) ToString() string {
	res := ""
	res += r.FunderAddress.String()
	res += "-"
	res += string(r.SharedRandom)
	res += "-"
	for i, v := range r.StakingTx {
		res += strconv.Itoa(i)
		res += v.Hash().String()
		res += "-"
	}
	res += strconv.FormatUint(r.StakingAmount, 10)
	return common.HashH([]byte(res)).String()
}

func (blockchain *BlockChain) buildReturnBeaconStakingAmountTx(
	curView *ShardBestState,
	info *returnStakingBeaconInfo,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (
	metadata.Transaction,
	uint64,
	error,
) {
	txStaking := []string{}
	for _, tx := range info.StakingTx {
		txStaking = append(txStaking, tx.Hash().String())
	}
	returnStakingMeta := metadata.NewReturnBeaconStaking(
		txStaking,
		info.FunderAddress,
		metadata.ReturnStakingMeta,
	)

	txParam := transaction.TxSalaryOutputParams{
		Amount:          info.StakingAmount,
		ReceiverAddress: &info.FunderAddress,
		TokenID:         &common.PRVCoinID,
		Type:            common.TxReturnStakingType,
	}

	makeMD := func(c privacy.Coin) metadata.Metadata {
		if c != nil && c.GetSharedRandom() != nil {
			returnStakingMeta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return returnStakingMeta
	}
	returnStakingTx, err := txParam.BuildTxSalary(producerPrivateKey, curView.GetCopiedTransactionStateDB(), makeMD)
	if err != nil {
		return nil, 0, errors.Errorf("cannot init return staking tx. Error %v", err)
	}
	// returnStakingTx.SetType()
	return returnStakingTx, info.StakingAmount, nil
}
