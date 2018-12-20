package blockchain

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"sort"

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

type Evals []*Evaluation

func (p Evals) Len() int           { return len(p) }
func (p Evals) Less(i, j int) bool { return p[i].OracleFeed.Price < p[j].OracleFeed.Price }
func (p Evals) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p Evals) SortEvals(isDesc bool) Evals {
	if isDesc {
		sort.Sort(sort.Reverse(p))
	}
	sort.Sort(p)
	return p
}

func (blockGen *BlkTmplGenerator) groupOracleFeedTxsByOracleType(
	updateFrequency uint32,
) (map[string][]metadata.Transaction, error) {
	txsByOracleType := map[string][]metadata.Transaction{}
	header := blockGen.chain.BestState[0].BestBlock.Header
	blockHash := header.PrevBlockHash
	for i := updateFrequency; i > 0; i-- {
		blockBytes, err := blockGen.chain.config.DataBase.FetchBlock(&blockHash)
		if err != nil {
			return nil, err
		}
		block := Block{}
		err = json.Unmarshal(blockBytes, &block)
		if err != nil {
			return nil, err
		}
		for _, tx := range block.Transactions {
			meta := tx.GetMetadata()
			if meta.GetType() != metadata.OracleFeedMeta {
				continue
			}
			oracleFeed, ok := meta.(*metadata.OracleFeed)
			if !ok {
				return nil, errors.New("Could not parse OracleFeed metadata")
			}
			assetTypeStr := string(oracleFeed.AssetType[:])
			_, existed := txsByOracleType[assetTypeStr]
			if !existed {
				txsByOracleType[assetTypeStr] = []metadata.Transaction{tx}
			} else {
				txsByOracleType[assetTypeStr] = append(txsByOracleType[assetTypeStr], tx)
			}
		}
		blockHash = block.Header.PrevBlockHash
	}
	return txsByOracleType, nil
}

func computeRewards(
	evals []*Evaluation,
	oracleRewardMultiplier uint8,
) (uint64, []*Evaluation) {
	sortedEvals := Evals(evals).SortEvals(false)
	medPos := len(evals) / 2
	minPos := medPos / 2
	maxPos := medPos + minPos
	delta := math.Abs(float64(sortedEvals[minPos].OracleFeed.Price - sortedEvals[maxPos].OracleFeed.Price))
	selectedPrice := sortedEvals[medPos].OracleFeed.Price
	rewardedEvals := []*Evaluation{}
	for i, eval := range sortedEvals {
		if i < minPos || i > maxPos {
			continue
		}
		basePayout := eval.Tx.GetTxFee()
		eval.Reward = basePayout + uint64(oracleRewardMultiplier)*uint64(math.Abs(delta-float64(2*(eval.OracleFeed.Price-selectedPrice)))/delta)
		rewardedEvals = append(rewardedEvals, eval)
	}
	return selectedPrice, rewardedEvals
}

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

func (blockGen *BlkTmplGenerator) updateOracleValues() ([]*Evaluation, error) {
	header := blockGen.chain.BestState[0].BestBlock.Header
	govParams := header.GOVConstitution.GOVParams
	oracleNetwork := govParams.OracleNetwork
	if header.Height == 0 || uint32(header.Height)%oracleNetwork.UpdateFrequency != 0 {
		return []*Evaluation{}, nil
	}
	txsByOracleType, err := blockGen.groupOracleFeedTxsByOracleType(oracleNetwork.UpdateFrequency)
	if err != nil {
		return nil, err
	}
	rewardAndRefundEvals := []*Evaluation{}
	updatedValues := map[string]uint64{}
	for oracleType, txs := range txsByOracleType {
		txsLen := len(txs)
		if txsLen < int(oracleNetwork.Quorum) {
			refundEvals := refundOracleFeeders(txs)
			rewardAndRefundEvals = append(rewardAndRefundEvals, refundEvals...)
			continue
		}
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
		}
		selectedPrice, rewardedEvals := computeRewards(
			evals,
			oracleNetwork.OracleRewardMultiplier,
		)
		updatedValues[oracleType] = selectedPrice
		rewardAndRefundEvals = append(rewardAndRefundEvals, rewardedEvals...)
	}
	// update oracle values in block header
	update(header.Oracle, updatedValues)
	return rewardAndRefundEvals, nil
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
