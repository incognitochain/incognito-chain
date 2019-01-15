package blockchain

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"sort"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
)

type Evaluation struct {
	Tx               *transaction.Tx
	OracleFeederAddr *privacy.PaymentAddress
	OracleFeed       *metadata.OracleFeed
	Reward           uint64
}

func sortEvalsByPrice(evals []*Evaluation, isDesc bool) []*Evaluation {
	sort.Slice(evals, func(i, j int) bool {
		if isDesc {
			return evals[i].OracleFeed.Price > evals[j].OracleFeed.Price
		}
		return evals[i].OracleFeed.Price <= evals[j].OracleFeed.Price
	})
	return evals
}

func (blockGen *BlkTmplGenerator) groupOracleFeedTxsByOracleType(
	updateFrequency uint32,
	chainID byte,
) (map[string][]metadata.Transaction, error) {
	txsByOracleType := map[string][]metadata.Transaction{}
	header := blockGen.chain.BestState[chainID].BestBlock.Header
	blockHash := header.PrevBlockHash
	for i := updateFrequency; i > 0; i-- {
		if blockHash.String() == (common.Hash{}).String() {
			return txsByOracleType, nil
		}
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
			if meta == nil || meta.GetType() != metadata.OracleFeedMeta {
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
	sortedEvals := sortEvalsByPrice(evals, false)
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

func getSenderAddress(tx *transaction.Tx) (*privacy.PaymentAddress, error) {
	meta := tx.GetMetadata()
	if meta == nil || meta.GetType() != metadata.OracleFeedMeta {
		return nil, errors.New("Metadata from tx is not OracleFeedMeta type.")
	}
	oracleFeed, ok := meta.(*metadata.OracleFeed)
	if !ok {
		return nil, errors.New("Could not parse OracleFeedMeta metadata.")
	}
	return &oracleFeed.FeederAddress, nil
}

func refundOracleFeeders(txs []metadata.Transaction) []*Evaluation {
	evals := []*Evaluation{}
	for _, tx := range txs {
		normalTx, ok := tx.(*transaction.Tx)
		if !ok {
			continue
		}
		senderAddr, err := getSenderAddress(normalTx)
		if err != nil {
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

func (blockGen *BlkTmplGenerator) updateOracleValues(
	newBlock *Block,
	updatedValues map[string]uint64,
) {
	oracleValues := newBlock.Header.Oracle
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

func (blockGen *BlkTmplGenerator) buildRewardAndRefundEvals(
	block *Block,
	chainID byte,
) ([]*Evaluation, map[string]uint64, error) {
	header := block.Header
	govParams := header.GOVConstitution.GOVParams
	oracleNetwork := govParams.OracleNetwork
	if header.Height == 0 || uint32(header.Height)%oracleNetwork.UpdateFrequency != 0 {
		return []*Evaluation{}, map[string]uint64{}, nil
	}
	txsByOracleType, err := blockGen.groupOracleFeedTxsByOracleType(oracleNetwork.UpdateFrequency, chainID)
	if err != nil {
		return nil, nil, err
	}
	rewardAndRefundEvals := []*Evaluation{}
	updatedOracleValues := map[string]uint64{}
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
			senderAddr, err := getSenderAddress(normalTx)
			if err != nil {
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
		updatedOracleValues[oracleType] = selectedPrice
		rewardAndRefundEvals = append(rewardAndRefundEvals, rewardedEvals...)
	}
	// update oracle values in block header
	// update(header.Oracle, updatedOracleValues)
	return rewardAndRefundEvals, updatedOracleValues, nil
}

func (blockGen *BlkTmplGenerator) buildOracleRewardTxs(
	chainID byte,
	privatekey *privacy.SpendingKey,
) ([]metadata.Transaction, uint64, map[string]uint64, error) {
	bestBlock := blockGen.chain.BestState[chainID].BestBlock
	evals, updatedOracleValues, err := blockGen.buildRewardAndRefundEvals(bestBlock, chainID)
	if err != nil {
		return []metadata.Transaction{}, 0, map[string]uint64{}, err
	}

	totalRewards := uint64(0)
	oracleRewardTxs := []metadata.Transaction{}
	for _, eval := range evals {
		oracleReward := metadata.NewOracleReward(*eval.Tx.Hash(), metadata.OracleRewardMeta)
		oracleRewardTx := new(transaction.Tx)
		err := oracleRewardTx.InitTxSalary(eval.Reward, eval.OracleFeederAddr, privatekey, blockGen.chain.GetDatabase(), oracleReward)
		if err != nil {
			return []metadata.Transaction{}, 0, map[string]uint64{}, err
		}
		oracleRewardTxs = append(oracleRewardTxs, oracleRewardTx)
		totalRewards += eval.Reward
	}
	return oracleRewardTxs, totalRewards, updatedOracleValues, nil
}

func removeOraclePubKeys(
	oracleRemovePubKeys [][]byte,
	oracleBoardPubKeys [][]byte,
) [][]byte {
	pubKeys := [][]byte{}
	for _, boardPK := range oracleBoardPubKeys {
		isRemoved := false
		for _, removePK := range oracleRemovePubKeys {
			if bytes.Equal(removePK, boardPK) {
				isRemoved = true
				break
			}
		}
		if !isRemoved {
			pubKeys = append(pubKeys, boardPK)
		}
	}
	return pubKeys
}

func (blockGen *BlkTmplGenerator) updateOracleBoard(
	newBlock *Block,
	txs []metadata.Transaction,
) error {
	if len(txs) == 0 {
		return nil
	}
	oraclePubKeys := newBlock.Header.GOVConstitution.GOVParams.OracleNetwork.OraclePubKeys
	sortedTxs := transaction.SortTxsByLockTime(txs, false)
	for _, tx := range sortedTxs {
		meta := tx.GetMetadata()
		updatingOracleBoard, ok := meta.(*metadata.UpdatingOracleBoard)
		if !ok {
			return errors.New("Could not parse UpdatingOracleBoard metadata.")
		}
		action := updatingOracleBoard.Action
		if action == metadata.Add {
			oraclePubKeys = append(oraclePubKeys, updatingOracleBoard.OraclePubKeys...)
		} else if action == metadata.Remove {
			oraclePubKeys = removeOraclePubKeys(updatingOracleBoard.OraclePubKeys, oraclePubKeys)
		}
	}
	return nil
}
