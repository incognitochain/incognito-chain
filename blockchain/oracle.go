package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"math"
	"sort"
	"strconv"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
)

type Evaluation struct {
	TxReqID          common.Hash
	OracleFeederAddr *privacy.PaymentAddress
	OracleFeed       *metadata.OracleFeed
	Reward           uint64
	TxFee            uint64
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
	TxFee   uint64              `json:"txFee"`
}

type UpdatingOracleBoardAction struct {
	Meta metadata.UpdatingOracleBoard `json:"meta"`
}

func buildInstForOracleFeedReq(
	shardID byte,
	contentStr string,
	beaconBestState *BestStateBeacon,
) ([][]string, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		return [][]string{}, err
	}
	var oracleFeedAction OracleFeedAction
	err = json.Unmarshal(contentBytes, &oracleFeedAction)
	if err != nil {
		return nil, err
	}
	md := oracleFeedAction.Meta
	feederPubKey := md.FeederAddress.Pk
	govParams := beaconBestState.StabilityInfo.GOVConstitution.GOVParams
	oraclePubKeys := govParams.OracleNetwork.OraclePubKeys
	instType := "notAccepted"
	for _, oraclePubKey := range oraclePubKeys {
		if oraclePubKey == hex.EncodeToString(feederPubKey) {
			instType = "accepted"
			break
		}
	}
	returnedInst := []string{
		strconv.Itoa(metadata.OracleFeedMeta),
		strconv.Itoa(int(shardID)),
		instType,
		contentStr,
	}
	return [][]string{returnedInst}, nil
}

func groupOracleFeedTxsByOracleType(
	bc *BlockChain,
	beaconBestState *BestStateBeacon,
	updateFrequency uint32,
) (map[string][][]string, error) {
	instsByOracleType := map[string][][]string{}
	blockHash := beaconBestState.BestBlock.Header.PrevBlockHash
	for i := updateFrequency; i > 0; i-- {
		if blockHash.String() == (common.Hash{}).String() {
			// return instsByOracleType, nil
			continue
		}
		blockBytes, err := bc.config.DataBase.FetchBlock(&blockHash)
		if err != nil {
			return nil, err
		}
		block := BeaconBlock{}
		err = json.Unmarshal(blockBytes, &block)
		if err != nil {
			return nil, err
		}
		for _, inst := range block.Body.Instructions {
			if len(inst) != 4 {
				continue
			}
			metaTypeStr := inst[0]
			instType := inst[2]
			if instType != "accepted" {
				continue
			}
			contentStr := inst[3]
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
				// assetTypeStr := string(oracleFeed.AssetType[:])
				assetTypeStr := oracleFeed.AssetType.String()
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
		basePayout := eval.TxFee
		if delta == 0 {
			eval.Reward = basePayout + uint64(oracleRewardMultiplier)
		} else {
			eval.Reward = basePayout + uint64(oracleRewardMultiplier)*uint64(math.Abs(delta-float64(2*(eval.OracleFeed.Price-selectedPrice)))/delta)
		}
		rewardedEvals = append(rewardedEvals, eval)
	}
	return selectedPrice, rewardedEvals
}

func refundOracleFeeders(insts [][]string) ([]*Evaluation, error) {
	evals := []*Evaluation{}
	for _, inst := range insts {
		contentStr := inst[3]
		contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
		if err != nil {
			return evals, err
		}
		var oracleFeedAction OracleFeedAction
		err = json.Unmarshal(contentBytes, &oracleFeedAction)
		if err != nil {
			return evals, err
		}
		md := oracleFeedAction.Meta
		paidFee := oracleFeedAction.TxFee
		eval := &Evaluation{
			TxReqID:          oracleFeedAction.TxReqID,
			OracleFeederAddr: &md.FeederAddress,
			OracleFeed:       nil,
			Reward:           paidFee,
			TxFee:            paidFee,
		}
		evals = append(evals, eval)
	}
	return evals, nil
}

func updateOracleValues(
	beaconBestState *BestStateBeacon,
	updatedValues map[string]uint64,
) {
	oracleValues := &bestStateBeacon.StabilityInfo.Oracle
	for oracleType, value := range updatedValues {
		// oracleTypeBytes := []byte(oracleType)
		oracleTypeBytes, _ := common.Hash{}.NewHashFromStr(oracleType)
		if bytes.Equal(oracleTypeBytes[:], common.DCBTokenID[:]) {
			oracleValues.DCBToken = value
			continue
		}
		if bytes.Equal(oracleTypeBytes[:], common.GOVTokenID[:]) {
			oracleValues.GOVToken = value
			continue
		}
		if bytes.Equal(oracleTypeBytes[:], common.ConstantID[:]) {
			oracleValues.Constant = value
			continue
		}
		if bytes.Equal(oracleTypeBytes[:], common.ETHAssetID[:]) {
			oracleValues.ETH = value
			continue
		}
		if bytes.Equal(oracleTypeBytes[:], common.BTCAssetID[:]) {
			oracleValues.BTC = value
			continue
		}
		if bytes.Equal(oracleTypeBytes[0:8], common.BondTokenID[0:8]) {
			oracleValues.Bonds[oracleType] = value
			continue
		}
	}
}

func buildRewardAndRefundEvals(
	bc *BlockChain,
	beaconBestState *BestStateBeacon,
) ([]*Evaluation, map[string]uint64, error) {
	beaconHeight := beaconBestState.BeaconHeight
	stabilityInfo := beaconBestState.StabilityInfo
	govParams := stabilityInfo.GOVConstitution.GOVParams
	oracleNetwork := govParams.OracleNetwork
	if beaconHeight == 0 || oracleNetwork == nil || oracleNetwork.UpdateFrequency == 0 || uint32(beaconHeight)%oracleNetwork.UpdateFrequency != 0 {
		return []*Evaluation{}, map[string]uint64{}, nil
	}
	instsByOracleType, err := groupOracleFeedTxsByOracleType(bc, beaconBestState, oracleNetwork.UpdateFrequency)
	if err != nil {
		return nil, nil, err
	}
	rewardAndRefundEvals := []*Evaluation{}
	updatedOracleValues := map[string]uint64{}
	for oracleType, insts := range instsByOracleType {
		instsLen := len(insts)
		if instsLen < int(oracleNetwork.Quorum) {
			refundEvals, err := refundOracleFeeders(insts)
			if err != nil {
				return nil, nil, err
			}
			rewardAndRefundEvals = append(rewardAndRefundEvals, refundEvals...)
			continue
		}

		evals := []*Evaluation{}
		for _, inst := range insts {
			contentStr := inst[3]
			contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
			if err != nil {
				return evals, updatedOracleValues, err
			}
			var oracleFeedAction OracleFeedAction
			err = json.Unmarshal(contentBytes, &oracleFeedAction)
			if err != nil {
				return evals, updatedOracleValues, err
			}
			md := oracleFeedAction.Meta
			paidFee := oracleFeedAction.TxFee
			eval := &Evaluation{
				TxReqID:          oracleFeedAction.TxReqID,
				OracleFeed:       &md,
				OracleFeederAddr: &md.FeederAddress,
				TxFee:            paidFee,
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
	return rewardAndRefundEvals, updatedOracleValues, nil
}

func (blockChain *BlockChain) buildOracleRewardInstructions(
	beaconBestState *BestStateBeacon,
) ([][]string, error) {
	evals, _, err := buildRewardAndRefundEvals(blockChain, beaconBestState)
	if err != nil {
		return nil, err
	}
	if len(evals) == 0 {
		return [][]string{}, nil
	}

	totalRewards := uint64(0)
	instructions := [][]string{}
	for _, eval := range evals {
		feederPk := eval.OracleFeederAddr.Pk[:]
		lastByte := feederPk[len(feederPk)-1]
		shardID := common.GetShardIDFromLastByte(lastByte)
		evalBytes, err := json.Marshal(eval)
		if err != nil {
			return nil, err
		}
		inst := []string{
			strconv.Itoa(metadata.OracleRewardMeta),
			strconv.Itoa(int(shardID)),
			"accepted",
			string(evalBytes),
		}
		instructions = append(instructions, inst)
		totalRewards += eval.Reward
	}
	return instructions, nil
}

func (bsb *BestStateBeacon) updateOracleParams(bc *BlockChain) error {
	evals, updatedOracleValues, err := buildRewardAndRefundEvals(bc, bsb)
	if err != nil {
		return err
	}
	if len(evals) == 0 {
		return nil
	}

	totalRewards := uint64(0)
	for _, eval := range evals {
		totalRewards += eval.Reward
	}

	if bestStateBeacon.StabilityInfo.SalaryFund < totalRewards {
		bestStateBeacon.StabilityInfo.SalaryFund = 0
	} else {
		bestStateBeacon.StabilityInfo.SalaryFund -= totalRewards
	}
	updateOracleValues(bsb, updatedOracleValues)
	return nil
}

func (blockGen *BlkTmplGenerator) buildOracleRewardTxs(
	evalStr string,
	privatekey *privacy.PrivateKey,
) ([]metadata.Transaction, error) {
	var eval Evaluation
	err := json.Unmarshal([]byte(evalStr), &eval)
	if err != nil {
		return nil, err
	}

	oracleReward := metadata.NewOracleReward(eval.TxReqID, metadata.OracleRewardMeta)
	oracleRewardTx := new(transaction.Tx)
	err = oracleRewardTx.InitTxSalary(eval.Reward, eval.OracleFeederAddr, privatekey, blockGen.chain.GetDatabase(), oracleReward)
	return []metadata.Transaction{oracleRewardTx}, err
}

func removeOraclePubKeys(
	oracleRemovePubKeys []string,
	oracleBoardPubKeys []string,
) []string {
	pubKeys := []string{}
	for _, boardPK := range oracleBoardPubKeys {
		isRemoved := false
		for _, removePK := range oracleRemovePubKeys {
			if removePK == boardPK {
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

func getGOVBoardPubKeys(beaconBestState *BestStateBeacon) [][]byte {
	govBoardAddresses := beaconBestState.StabilityInfo.GOVGovernor.GovernorInfo.BoardPaymentAddress
	pks := [][]byte{}
	for _, addr := range govBoardAddresses {
		pks = append(pks, addr.Pk[:])
	}
	return pks
}

func buildInstForUpdatingOracleBoardReq(
	shardID byte,
	contentStr string,
	beaconBestState *BestStateBeacon,
) ([][]string, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		return [][]string{}, err
	}
	var updatingOracleBoardAction UpdatingOracleBoardAction
	err = json.Unmarshal(contentBytes, &updatingOracleBoardAction)
	if err != nil {
		return nil, err
	}
	md := updatingOracleBoardAction.Meta
	govBoardPks := getGOVBoardPubKeys(beaconBestState)
	boardLen := len(govBoardPks)
	if boardLen == 0 {
		returnedInst := []string{
			strconv.Itoa(metadata.UpdatingOracleBoardMeta),
			strconv.Itoa(int(shardID)),
			"notAccepted",
			contentStr,
		}
		return [][]string{returnedInst}, nil
	}

	// // verify signs
	metaBytes := md.Hash()[:]
	signs := md.Signs
	verifiedSignCount := 0
	for _, pubKey := range govBoardPks {
		hexStrPk := hex.EncodeToString(pubKey)
		sign, existed := signs[hexStrPk]
		if !existed {
			continue
		}

		verKey := new(privacy.SchnPubKey)
		verKey.PK = new(privacy.EllipticPoint)
		err = verKey.PK.Decompress(pubKey)
		if err != nil {
			continue
		}

		verKey.G = new(privacy.EllipticPoint)
		verKey.G.Set(privacy.PedCom.G[privacy.SK].X, privacy.PedCom.G[privacy.SK].Y)

		verKey.H = new(privacy.EllipticPoint)
		verKey.H.Set(privacy.PedCom.G[privacy.RAND].X, privacy.PedCom.G[privacy.RAND].Y)

		// convert signature from byte array to SchnorrSign
		signature := new(privacy.SchnSignature)
		signature.SetBytes(sign)

		// verify signature
		res := verKey.Verify(signature, metaBytes)
		if res {
			verifiedSignCount += 1
		}
	}
	if verifiedSignCount < int(math.Floor(float64(boardLen/2)))+1 {
		returnedInst := []string{
			strconv.Itoa(metadata.UpdatingOracleBoardMeta),
			strconv.Itoa(int(shardID)),
			"notAccepted",
			contentStr,
		}
		return [][]string{returnedInst}, nil
	}
	mdBytes, err := json.Marshal(md)
	if err != nil {
		return nil, err
	}

	returnedInst := []string{
		strconv.Itoa(metadata.UpdatingOracleBoardMeta),
		strconv.Itoa(int(shardID)),
		"accepted",
		string(mdBytes),
	}
	return [][]string{returnedInst}, nil
}
