package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strconv"

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

type OracleFeedAction struct {
	TxReqID common.Hash         `json:"txReqId"`
	Meta    metadata.OracleFeed `json:"meta"`
}

func (blockGen *BlkTmplGenerator) groupOracleFeedTxsByOracleType(
	beaconBestState *BestStateBeacon,
	updateFrequency uint32,
) (map[string][][]string, error) {
	instsByOracleType := map[string][][]string{}
	blockHash := beaconBestState.BestBlock.Header.PrevBlockHash
	for i := updateFrequency; i > 0; i-- {
		if blockHash.String() == (common.Hash{}).String() {
			return instsByOracleType, nil
		}
		blockBytes, err := blockGen.chain.config.DataBase.FetchBlock(&blockHash)
		if err != nil {
			return nil, err
		}
		block := BeaconBlock{}
		err = json.Unmarshal(blockBytes, &block)
		if err != nil {
			return nil, err
		}
		for _, inst := range block.Body.Instructions {
			metaTypeStr := inst[0]
			contentStr := inst[1]
			if metaTypeStr == strconv.Itoa(metadata.OracleFeedMeta) {
				contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
				if err != nil {
					return nil, err
				}
				var oracleFeedAction OracleFeedAction
				err = json.Unmarshal(contentBytes, &oracleFeedAction)
				if err != nil {
					return nil, err
				}
				oracleFeed := oracleFeedAction.Meta
				assetTypeStr := string(oracleFeed.AssetType[:])
				_, existed := instsByOracleType[assetTypeStr]
				if !existed {
					instsByOracleType[assetTypeStr] = [][]string{inst}
				} else {
					instsByOracleType[assetTypeStr] = append(instsByOracleType[assetTypeStr], inst)
				}
			}
		}
		blockHash = block.Header.PrevBlockHash
	}
	return instsByOracleType, nil
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

// func (blockGen *BlkTmplGenerator) updateOracleValues(
// 	newBlock *Block,
// 	updatedValues map[string]uint64,
// ) {
// 	oracleValues := newBlock.Header.Oracle
// 	for oracleType, value := range updatedValues {
// 		oracleTypeBytes := []byte(oracleType)
// 		if bytes.Equal(oracleTypeBytes, common.DCBTokenID[:]) {
// 			oracleValues.DCBToken = value
// 			continue
// 		}
// 		if bytes.Equal(oracleTypeBytes, common.GOVTokenID[:]) {
// 			oracleValues.GOVToken = value
// 			continue
// 		}
// 		if bytes.Equal(oracleTypeBytes, common.ConstantID[:]) {
// 			oracleValues.Constant = value
// 			continue
// 		}
// 		if bytes.Equal(oracleTypeBytes, common.ETHAssetID[:]) {
// 			oracleValues.ETH = value
// 			continue
// 		}
// 		if bytes.Equal(oracleTypeBytes, common.BTCAssetID[:]) {
// 			oracleValues.BTC = value
// 			continue
// 		}
// 		if bytes.Equal(oracleTypeBytes[0:8], common.BondTokenID[0:8]) {
// 			oracleValues.Bonds[oracleType] = value
// 			continue
// 		}
// 	}
// }

func (blockGen *BlkTmplGenerator) buildRewardAndRefundEvals(
	beaconBestState *BestStateBeacon,
) ([]*Evaluation, map[string]uint64, error) {
	beaconHeight := beaconBestState.BeaconHeight
	stabilityInfo := beaconBestState.StabilityInfo
	govParams := stabilityInfo.GOVConstitution.GOVParams
	oracleNetwork := govParams.OracleNetwork
	if beaconHeight == 0 || uint32(beaconHeight)%oracleNetwork.UpdateFrequency != 0 {
		return []*Evaluation{}, map[string]uint64{}, nil
	}
	_, err := blockGen.groupOracleFeedTxsByOracleType(beaconBestState, oracleNetwork.UpdateFrequency)
	if err != nil {
		return nil, nil, err
	}
	rewardAndRefundEvals := []*Evaluation{}
	updatedOracleValues := map[string]uint64{}
	// update oracle values in block header
	// update(header.Oracle, updatedOracleValues)
	return rewardAndRefundEvals, updatedOracleValues, nil
}

func (blockGen *BlkTmplGenerator) buildOracleRewardTxs(
	beaconBestState *BestStateBeacon,
	privatekey *privacy.SpendingKey,
) ([]metadata.Transaction, uint64, map[string]uint64, error) {
	// bestBlock := beaconBestState.BestBlock
	evals, updatedOracleValues, err := blockGen.buildRewardAndRefundEvals(beaconBestState)
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
