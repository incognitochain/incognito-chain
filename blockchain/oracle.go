package blockchain

import (
	"bytes"

	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	privacy "github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/transaction"
)

type Evaluation struct {
	Tx               *transaction.Tx
	OracleFeederAddr *privacy.PaymentAddress
	OracleFeed       *metadata.OracleFeed
	Reward           uint64
}

func (blockGen *BlkTmplGenerator) groupOracleFeedTxsByType() map[string][]metadata.Transaction {
	return map[string][]metadata.Transaction{}
}

// func computeMedian(txs []metadata.Transaction) uint64 {
// 	txsLen := len(txs)
// 	if txsLen == 0 {
// 		return 0
// 	}
// 	sum := 0
// 	for _, tx := range txs {
// 		meta := tx.GetMetadata()
// 		oracleFeed, ok := meta.(*metadata.OracleFeed)
// 		if !ok {
// 			continue
// 		}
// 		sum += oracleFeed.Price
// 	}
// 	return sum / txsLen
// }

func getSenderAddress(tx *transaction.Tx) *privacy.PaymentAddress {
	if tx.Proof == nil || len(tx.Proof.InputCoins) == 0 {
		return nil
	}
	coin := tx.Proof.InputCoins[0].CoinDetails
	if coin == nil {
		return nil
	}
	pk := coin.PublicKey.Compress()
	return &privacy.PaymentAddress{
		Pk: pk,
	}
}

func computeReward(feedPrice uint64, median uint64) uint64 {
	// TODO: update here
	return 100
}

func refundOracleFeeders(txs []metadata.Transaction) []*Evaluation {
	evals := []*Evaluation{}
	for _, tx := range txs {
		normalTx, ok := tx.(*transaction.Tx)
		if !ok {
			continue
		}
		senderAddr := getSenderAddress(normalTx)
		if senderAddr == nil {
			continue
		}
		eval := &Evaluation{
			Tx:               normalTx,
			OracleFeederAddr: senderAddr,
			OracleFeed:       nil,
			Reward:           normalTx.GetTxFee(),
		}
		evals = append(evals, eval)
	}
	return evals
}

func update(oracleValues *params.Oracle, updatedValues map[string]uint64) {
	for oracleType, value := range updatedValues {
		oracleTypeBytes := []byte(oracleType)
		if bytes.Equal(oracleTypeBytes, common.DCBTokenID[:]) {
			oracleValues.DCBToken = value
			continue
		}
		if bytes.Equal(oracleTypeBytes, common.GOVTokenID[:]) {
			oracleValues.GOVToken = value
			continue
		}
		if bytes.Equal(oracleTypeBytes, common.ConstantID[:]) {
			oracleValues.Constant = value
			continue
		}
		if bytes.Equal(oracleTypeBytes, common.ETHAssetID[:]) {
			oracleValues.ETH = value
			continue
		}
		if bytes.Equal(oracleTypeBytes, common.BTCAssetID[:]) {
			oracleValues.BTC = value
			continue
		}
		if bytes.Equal(oracleTypeBytes[0:8], common.BondTokenID[0:8]) {
			oracleValues.Bonds[oracleType] = value
			continue
		}
	}
}

func (blockGen *BlkTmplGenerator) updateOracleValues() []*Evaluation {
	header := blockGen.chain.BestState[0].BestBlock.Header
	oracleNetwork := header.GOVConstitution.GOVParams.OracleNetwork
	if header.Height == 0 || uint32(header.Height)%oracleNetwork.UpdateFrequency != 0 {
		return []*Evaluation{}
	}
	txsByType := blockGen.groupOracleFeedTxsByType()
	rewardAndRefundEvals := []*Evaluation{}
	updatedValues := map[string]uint64{}
	for oracleType, txs := range txsByType {
		txsLen := len(txs)
		if txsLen < int(oracleNetwork.Quorum) {
			refundEvals := refundOracleFeeders(txs)
			rewardAndRefundEvals = append(rewardAndRefundEvals, refundEvals...)
			continue
		}
		evals := []*Evaluation{}
		sum := uint64(0)
		for _, tx := range txs {
			normalTx, ok := tx.(*transaction.Tx)
			if !ok {
				continue
			}
			senderAddr := getSenderAddress(normalTx)
			if senderAddr == nil {
				continue
			}
			meta := tx.GetMetadata()
			oracleFeed, ok := meta.(*metadata.OracleFeed)
			if !ok {
				continue
			}
			eval := &Evaluation{
				Tx:               normalTx,
				OracleFeed:       oracleFeed,
				OracleFeederAddr: senderAddr,
			}
			evals = append(evals, eval)
			sum += oracleFeed.Price
		}
		median := sum / uint64(txsLen)
		updatedValues[oracleType] = median
		for _, eval := range evals {
			feedPrice := eval.OracleFeed.Price
			reward := computeReward(feedPrice, median)
			if reward > 0 {
				rewardAndRefundEvals = append(rewardAndRefundEvals, eval)
			}
		}
	}
	// update oracle values in block header
	update(header.Oracle, updatedValues)
	return rewardAndRefundEvals
}

func (blockGen *BlkTmplGenerator) buildOracleRewardTxs(
	evals []*Evaluation,
	chainID byte,
	privatekey *privacy.SpendingKey,
) ([]*transaction.Tx, uint64, error) {
	totalRewards := uint64(0)
	oracleRewardTxs := []*transaction.Tx{}
	for _, eval := range evals {
		oracleRewardTx, err := transaction.CreateTxSalary(eval.Reward, eval.OracleFeederAddr, privatekey, blockGen.chain.GetDatabase())
		if err != nil {
			return []*transaction.Tx{}, 0, err
		}
		oracleReward := metadata.NewOracleReward(*eval.Tx.Hash(), metadata.OracleRewardMeta)
		oracleRewardTx.SetMetadata(oracleReward)
		oracleRewardTxs = append(oracleRewardTxs, oracleRewardTx)
		totalRewards += eval.Reward
	}
	return oracleRewardTxs, totalRewards, nil
}
