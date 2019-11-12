package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"sort"
	"strconv"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/pkg/errors"
)

// build instructions at beacon chain before syncing to shards
func (blockchain *BlockChain) buildBridgeInstructions(
	shardID byte,
	shardBlockInstructions [][]string,
	beaconHeight uint64,
	db database.DatabaseInterface,
) ([][]string, error) {
	currentPDEState, err := InitCurrentPDEStateFromDB(db, beaconHeight-1)
	if err != nil {
		Logger.log.Error(err)
	}
	accumulatedValues := &metadata.AccumulatedValues{
		UniqETHTxsUsed:   [][]byte{},
		DBridgeTokenPair: map[string][]byte{},
		CBridgeTokens:    []*common.Hash{},
	}
	instructions := [][]string{}
	pdeContributionInsts := [][]string{}
	pdeTradeInsts := [][]string{}
	pdeWithdrawalInsts := [][]string{}
	for _, inst := range shardBlockInstructions {
		if len(inst) < 2 {
			continue
		}
		if inst[0] == SetAction || inst[0] == StakeAction || inst[0] == SwapAction || inst[0] == RandomAction || inst[0] == AssignAction {
			continue
		}

		metaType, err := strconv.Atoi(inst[0])
		if err != nil {
			continue
		}
		contentStr := inst[1]
		newInst := [][]string{}
		switch metaType {
		case metadata.ContractingRequestMeta:
			newInst, err = blockchain.buildInstructionsForContractingReq(contentStr, shardID, metaType)

		case metadata.IssuingRequestMeta:
			newInst, err = blockchain.buildInstructionsForIssuingReq(contentStr, shardID, metaType, accumulatedValues)

		case metadata.IssuingETHRequestMeta:
			newInst, err = blockchain.buildInstructionsForIssuingETHReq(contentStr, shardID, metaType, accumulatedValues)

		case metadata.BurningRequestMeta:
			burningConfirm := []string{}
			burningConfirm, err = buildBurningConfirmInst(inst, beaconHeight, db)
			newInst = [][]string{burningConfirm}

		case metadata.PDEContributionMeta:
			pdeContributionInsts = append(pdeContributionInsts, inst)

		case metadata.PDETradeRequestMeta:
			pdeTradeInsts = append(pdeTradeInsts, inst)

		case metadata.PDEWithdrawalRequestMeta:
			pdeWithdrawalInsts = append(pdeWithdrawalInsts, inst)

		default:
			continue
		}

		if err != nil {
			Logger.log.Error(err)
			continue
		}
		if len(newInst) > 0 {
			instructions = append(instructions, newInst...)
		}
	}
	pdeInsts, err := blockchain.handlePDEInsts(shardID, beaconHeight-1, currentPDEState, pdeContributionInsts, pdeTradeInsts, pdeWithdrawalInsts)
	if err != nil {
		Logger.log.Error(err)
		return instructions, nil
	}
	if len(pdeInsts) > 0 {
		instructions = append(instructions, pdeInsts...)
	}
	return instructions, nil
}

func sortPDETradeInstsByFee(
	beaconHeight uint64,
	currentPDEState *CurrentPDEState,
	pdeTradeInsts [][]string,
) [][]string {
	sortedInsts := [][]string{}
	wrongFormatInsts := [][]string{}
	tradesByPairs := make(map[string][]metadata.PDETradeRequestAction)
	for _, inst := range pdeTradeInsts {
		contentStr := inst[1]
		contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while decoding content string of pde trade action: %+v", err)
			wrongFormatInsts = append(wrongFormatInsts, inst)
			continue
		}
		var pdeTradeReqAction metadata.PDETradeRequestAction
		err = json.Unmarshal(contentBytes, &pdeTradeReqAction)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling pde trade action: %+v", err)
			wrongFormatInsts = append(wrongFormatInsts, inst)
			continue
		}
		tradeMeta := pdeTradeReqAction.Meta
		poolPairKey := string(lvdb.BuildPDEPoolForPairKey(beaconHeight, tradeMeta.TokenIDToBuyStr, tradeMeta.TokenIDToSellStr))
		tradesByPair, found := tradesByPairs[poolPairKey]
		if !found {
			tradesByPairs[poolPairKey] = []metadata.PDETradeRequestAction{pdeTradeReqAction}
		} else {
			tradesByPairs[poolPairKey] = append(tradesByPair, pdeTradeReqAction)
		}
	}
	sortedInsts = append(sortedInsts, wrongFormatInsts...)

	notExistingPairTradeActions := []metadata.PDETradeRequestAction{}
	sortedExistingPairTradeActions := []metadata.PDETradeRequestAction{}
	for poolPairKey, tradeActions := range tradesByPairs {
		poolPair, found := currentPDEState.PDEPoolPairs[poolPairKey]
		if !found || poolPair == nil {
			notExistingPairTradeActions = append(notExistingPairTradeActions, tradeActions...)
			continue
		}
		if poolPair.Token1PoolValue == 0 || poolPair.Token2PoolValue == 0 {
			notExistingPairTradeActions = append(notExistingPairTradeActions, tradeActions...)
			continue
		}

		// sort trade actions by trading fee
		sort.Slice(tradeActions, func(i, j int) bool {
			// comparing a/b to c/d is equivalent comparing a*d to c*b
			firstItemProportion := big.NewInt(0)
			firstItemProportion.Mul(
				big.NewInt(int64(tradeActions[i].Meta.TradingFee)),
				big.NewInt(int64(tradeActions[j].Meta.SellAmount)),
			)
			secondItemProportion := big.NewInt(0)
			secondItemProportion.Mul(
				big.NewInt(int64(tradeActions[j].Meta.TradingFee)),
				big.NewInt(int64(tradeActions[i].Meta.SellAmount)),
			)
			return firstItemProportion.Cmp(secondItemProportion) == 1
		})
		sortedExistingPairTradeActions = append(sortedExistingPairTradeActions, tradeActions...)
	}
	for _, action := range notExistingPairTradeActions {
		actionContentBytes, _ := json.Marshal(action)
		actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
		inst := []string{strconv.Itoa(metadata.PDETradeRequestMeta), actionContentBase64Str}
		sortedInsts = append(sortedInsts, inst)
	}
	for _, action := range sortedExistingPairTradeActions {
		actionContentBytes, _ := json.Marshal(action)
		actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
		inst := []string{strconv.Itoa(metadata.PDETradeRequestMeta), actionContentBase64Str}
		sortedInsts = append(sortedInsts, inst)
	}
	return sortedInsts
}

func (blockchain *BlockChain) handlePDEInsts(
	shardID byte,
	beaconHeight uint64,
	currentPDEState *CurrentPDEState,
	pdeContributionInsts [][]string,
	pdeTradeInsts [][]string,
	pdeWithdrawalInsts [][]string,
) ([][]string, error) {
	instructions := [][]string{}
	sortedTradesInsts := sortPDETradeInstsByFee(
		beaconHeight,
		currentPDEState,
		pdeTradeInsts,
	)
	for _, inst := range sortedTradesInsts {
		contentStr := inst[1]
		newInst, err := blockchain.buildInstructionsForPDETrade(contentStr, shardID, metadata.PDETradeRequestMeta, currentPDEState, beaconHeight)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		if len(newInst) > 0 {
			instructions = append(instructions, newInst...)
		}
	}
	for _, inst := range pdeWithdrawalInsts {
		contentStr := inst[1]
		newInst, err := blockchain.buildInstructionsForPDEWithdrawal(contentStr, shardID, metadata.PDEWithdrawalRequestMeta, currentPDEState, beaconHeight)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		if len(newInst) > 0 {
			instructions = append(instructions, newInst...)
		}
	}
	for _, inst := range pdeContributionInsts {
		contentStr := inst[1]
		newInst, err := blockchain.buildInstructionsForPDEContribution(contentStr, shardID, metadata.PDEContributionMeta, currentPDEState, beaconHeight)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		if len(newInst) > 0 {
			instructions = append(instructions, newInst...)
		}
	}
	return instructions, nil
}

// buildBurningConfirmInst builds on beacon an instruction confirming a tx burning bridge-token
func buildBurningConfirmInst(inst []string, height uint64, db database.DatabaseInterface) ([]string, error) {
	BLogger.log.Infof("Build BurningConfirmInst: %s", inst)
	// Parse action and get metadata
	var burningReqAction BurningReqAction
	err := decodeContent(inst[1], &burningReqAction)
	if err != nil {
		return nil, errors.Wrap(err, "invalid BurningRequest")
	}
	md := burningReqAction.Meta
	txID := burningReqAction.RequestedTxID // to prevent double-release token
	shardID := byte(common.BridgeShardID)

	// Convert to external tokenID
	tokenID, err := findExternalTokenID(&md.TokenID, db)
	if err != nil {
		return nil, err
	}

	// Convert amount to big.Int to get bytes later
	amount := big.NewInt(0).SetUint64(md.BurningAmount)
	if bytes.Equal(tokenID, rCommon.HexToAddress(common.EthAddrStr).Bytes()) {
		// Convert Gwei to Wei for Ether
		amount = amount.Mul(amount, big.NewInt(1000000000))
	}

	// Convert height to big.Int to get bytes later
	h := big.NewInt(0).SetUint64(height)

	return []string{
		strconv.Itoa(metadata.BurningConfirmMeta),
		strconv.Itoa(int(shardID)),
		base58.Base58Check{}.Encode(tokenID, 0x00),
		md.RemoteAddress,
		base58.Base58Check{}.Encode(amount.Bytes(), 0x00),
		txID.String(),
		base58.Base58Check{}.Encode(md.TokenID[:], 0x00),
		base58.Base58Check{}.Encode(h.Bytes(), 0x00),
	}, nil
}

// findExternalTokenID finds the external tokenID for a bridge token from database
func findExternalTokenID(tokenID *common.Hash, db database.DatabaseInterface) ([]byte, error) {
	allBridgeTokensBytes, err := db.GetAllBridgeTokens()
	if err != nil {
		return nil, err
	}
	var allBridgeTokens []*lvdb.BridgeTokenInfo
	err = json.Unmarshal(allBridgeTokensBytes, &allBridgeTokens)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for _, token := range allBridgeTokens {
		if token.TokenID.IsEqual(tokenID) && len(token.ExternalTokenID) > 0 {
			return token.ExternalTokenID, nil
		}
	}
	return nil, errors.New("invalid tokenID")
}
