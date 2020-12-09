package metadata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/basemeta"
	"github.com/incognitochain/incognito-chain/wallet"
	"strconv"
)

type PortalRedeemFromLiquidationPoolResponseV3 struct {
	basemeta.MetadataBase
	RequestStatus       string
	ReqTxID             common.Hash
	RequesterAddrStr    string
	ExternalAddress     string
	RedeemAmount        uint64
	MintedPRVCollateral uint64
	TokenID             string
}

func NewPortalRedeemFromLiquidationPoolResponseV3(
	requestStatus string,
	reqTxID common.Hash,
	requesterAddressStr string,
	redeemAmount uint64,
	amount uint64,
	tokenID string,
	metaType int,
) *PortalRedeemFromLiquidationPoolResponseV3 {
	metadataBase := basemeta.MetadataBase{
		Type: metaType,
	}
	return &PortalRedeemFromLiquidationPoolResponseV3{
		RequestStatus:       requestStatus,
		ReqTxID:             reqTxID,
		MetadataBase:        metadataBase,
		RequesterAddrStr:    requesterAddressStr,
		RedeemAmount:        redeemAmount,
		MintedPRVCollateral: amount,
		TokenID:             tokenID,
	}
}

func (iRes PortalRedeemFromLiquidationPoolResponseV3) CheckTransactionFee(tr basemeta.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes PortalRedeemFromLiquidationPoolResponseV3) ValidateTxWithBlockChain(txr basemeta.Transaction, chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, shardID byte, db *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (iRes PortalRedeemFromLiquidationPoolResponseV3) ValidateSanityData(chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, beaconHeight uint64, txr basemeta.Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes PortalRedeemFromLiquidationPoolResponseV3) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return iRes.Type == basemeta.PortalRedeemFromLiquidationPoolResponseMetaV3
}

func (iRes PortalRedeemFromLiquidationPoolResponseV3) Hash() *common.Hash {
	record := iRes.MetadataBase.Hash().String()
	record += iRes.RequestStatus
	record += iRes.ReqTxID.String()
	record += iRes.RequesterAddrStr
	record += strconv.FormatUint(iRes.RedeemAmount, 10)
	record += strconv.FormatUint(iRes.MintedPRVCollateral, 10)
	record += iRes.TokenID
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *PortalRedeemFromLiquidationPoolResponseV3) CalculateSize() uint64 {
	return basemeta.CalculateSize(iRes)
}

func (iRes PortalRedeemFromLiquidationPoolResponseV3) VerifyMinerCreatedTxBeforeGettingInBlock(
	txsInBlock []basemeta.Transaction,
	txsUsed []int,
	insts [][]string,
	instUsed []int,
	shardID byte,
	tx basemeta.Transaction,
	chainRetriever basemeta.ChainRetriever,
	ac *basemeta.AccumulatedValues,
	shardViewRetriever basemeta.ShardViewRetriever,
	beaconViewRetriever basemeta.BeaconViewRetriever,
) (bool, error) {
	idx := -1
	for i, inst := range insts {
		if len(inst) < 4 { // this is not PortalRedeemFromLiquidationPoolMeta response instruction
			continue
		}
		instMetaType := inst[0]
		if instUsed[i] > 0 ||
			instMetaType != strconv.Itoa(basemeta.PortalRedeemFromLiquidationPoolMetaV3) {
			continue
		}
		instReqStatus := inst[2]
		if instReqStatus != iRes.RequestStatus {
			Logger.log.Errorf("WARNING - VALIDATION: instReqStatus %v is different from iRes.RequestStatus %v", instReqStatus, iRes.RequestStatus)
			continue
		}
		if (instReqStatus != common.PortalRedeemFromLiquidationPoolSuccessChainStatus) &&
			(instReqStatus != common.PortalRedeemFromLiquidationPoolRejectedChainStatus) {
			Logger.log.Errorf("WARNING - VALIDATION: instReqStatus is not correct %v", instReqStatus)
			continue
		}

		var shardIDFromInst byte
		var txReqIDFromInst common.Hash
		var requesterAddrStrFromInst string
		var redeemAmountFromInst uint64
		var mintedPRVCollateral uint64
		//var tokenIDStrFromInst string

		contentBytes := []byte(inst[3])
		var redeemReqContent PortalRedeemFromLiquidationPoolContentV3
		err := json.Unmarshal(contentBytes, &redeemReqContent)
		if err != nil {
			Logger.log.Error("WARNING - VALIDATION: an error occurred while parsing portal redeem liquidate exchange rates content: ", err)
			continue
		}

		shardIDFromInst = redeemReqContent.ShardID
		txReqIDFromInst = redeemReqContent.TxReqID
		requesterAddrStrFromInst = redeemReqContent.RedeemerIncAddressStr
		redeemAmountFromInst = redeemReqContent.RedeemAmount
		mintedPRVCollateral = redeemReqContent.MintedPRVCollateral
		//tokenIDStrFromInst = redeemReqContent.TokenID

		if !bytes.Equal(iRes.ReqTxID[:], txReqIDFromInst[:]) ||
			shardID != shardIDFromInst {
			continue
		}

		if requesterAddrStrFromInst != iRes.RequesterAddrStr {
			Logger.log.Errorf("Error - VALIDATION: Requester address %v is not matching to Requester address in instruction %v", iRes.RequesterAddrStr, requesterAddrStrFromInst)
			continue
		}

		if mintedPRVCollateral != iRes.MintedPRVCollateral {
			Logger.log.Errorf("Error - VALIDATION:  mintedPRVCollateral %v is not matching to  TotalPTokenReceived in instruction %v", iRes.MintedPRVCollateral, redeemAmountFromInst)
			continue
		}

		if redeemAmountFromInst != iRes.RedeemAmount {
			Logger.log.Errorf("Error - VALIDATION: Redeem amount %v is not matching to redeem amount in instruction %v", iRes.RedeemAmount, redeemAmountFromInst)
			continue
		}

		key, err := wallet.Base58CheckDeserialize(requesterAddrStrFromInst)
		if err != nil {
			Logger.log.Info("WARNING - VALIDATION: an error occurred while deserializing requester address string: ", err)
			continue
		}

		mintedTokenID := common.PRVCoinID.String()
		mintedAmount := mintedPRVCollateral
		if instReqStatus == common.PortalRedeemFromLiquidationPoolRejectedChainStatus {
			mintedTokenID = redeemReqContent.TokenID
			mintedAmount = redeemAmountFromInst
		}

		_, pk, paidAmount, assetID := tx.GetTransferData()
		if !bytes.Equal(key.KeySet.PaymentAddress.Pk[:], pk[:]) ||
			mintedAmount != paidAmount ||
			mintedTokenID != assetID.String() {
			continue
		}
		idx = i
		break
	}

	if idx == -1 { // not found the issuance request tx for this response
		return false, fmt.Errorf(fmt.Sprintf("no PortalRedeemFromLiquidationPoolMetaV3 instruction found for PortalRedeemFromLiquidationPoolResponseV3 tx %s", tx.Hash().String()))
	}

	instUsed[idx] = 1
	return true, nil
}
