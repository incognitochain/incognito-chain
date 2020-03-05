package metadata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/wallet"
	"strconv"
)

type PortalRequestUnlockCollateralResponse struct {
	MetadataBase
	RequestStatus    string
	ReqTxID          common.Hash
	CustodianAddrStr string
	UnlockedAmount   uint64
	MintedTokenID    string
}

func NewPortalRequestUnlockCollateralResponse(
	reqStatus string,
	reqTxID common.Hash,
	requesterAddressStr string,
	amount uint64,
	tokenID string,
	metaType int,
) *PortalRequestUnlockCollateralResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &PortalRequestUnlockCollateralResponse{
		RequestStatus:    reqStatus,
		ReqTxID:          reqTxID,
		MetadataBase:     metadataBase,
		CustodianAddrStr: requesterAddressStr,
		UnlockedAmount:   amount,
		MintedTokenID:    tokenID,
	}
}

func (iRes PortalRequestUnlockCollateralResponse) CheckTransactionFee(tr Transaction, minFee uint64, beaconHeight int64, db database.DatabaseInterface) bool {
	// no need to have fee for this tx
	return true
}

func (iRes PortalRequestUnlockCollateralResponse) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (iRes PortalRequestUnlockCollateralResponse) ValidateSanityData(bcr BlockchainRetriever, txr Transaction, beaconHeight uint64) (bool, bool, error) {
	return false, true, nil
}

func (iRes PortalRequestUnlockCollateralResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return iRes.Type == PortalRequestUnlockCollateralResponseMeta
}

func (iRes PortalRequestUnlockCollateralResponse) Hash() *common.Hash {
	record := iRes.MetadataBase.Hash().String()
	record += iRes.RequestStatus
	record += iRes.ReqTxID.String()
	record += iRes.CustodianAddrStr
	record += strconv.FormatUint(iRes.UnlockedAmount, 10)
	record += iRes.MintedTokenID
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *PortalRequestUnlockCollateralResponse) CalculateSize() uint64 {
	return calculateSize(iRes)
}

func (iRes PortalRequestUnlockCollateralResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
	txsInBlock []Transaction,
	txsUsed []int,
	insts [][]string,
	instUsed []int,
	shardID byte,
	tx Transaction,
	bcr BlockchainRetriever,
	ac *AccumulatedValues,
) (bool, error) {
	idx := -1
	for i, inst := range insts {
		if len(inst) < 4 { // this is not PortalReqUnlockCollateral instruction
			continue
		}
		instMetaType := inst[0]
		if instUsed[i] > 0 ||
			instMetaType != strconv.Itoa(PortalRequestUnlockCollateralMeta) {
			continue
		}
		instDepositStatus := inst[2]
		if instDepositStatus != iRes.RequestStatus ||
			(instDepositStatus != common.PortalReqUnlockCollateralAcceptedChainStatus) {
			continue
		}

		var shardIDFromInst byte
		var txReqIDFromInst common.Hash
		var custodianAddrStrFromInst string
		var unlockAmountFromInst uint64
		//var tokenIDStrFromInst string

		contentBytes := []byte(inst[3])
		var reqUnlockContent PortalRequestUnlockCollateralContent
		err := json.Unmarshal(contentBytes, &reqUnlockContent)
		if err != nil {
			Logger.log.Error("WARNING - VALIDATION: an error occured while parsing portal request unlock collateral content: ", err)
			continue
		}
		shardIDFromInst = reqUnlockContent.ShardID
		txReqIDFromInst = reqUnlockContent.TxReqID
		custodianAddrStrFromInst = reqUnlockContent.CustodianAddressStr
		unlockAmountFromInst = reqUnlockContent.UnlockAmount

		if !bytes.Equal(iRes.ReqTxID[:], txReqIDFromInst[:]) ||
			shardID != shardIDFromInst {
			continue
		}

		if custodianAddrStrFromInst != iRes.CustodianAddrStr {
			Logger.log.Errorf("Error - VALIDATION: Custodian address %v is not matching to Custodian address in instruction %v", iRes.CustodianAddrStr, custodianAddrStrFromInst)
			continue
		}

		if unlockAmountFromInst != iRes.UnlockedAmount {
			Logger.log.Errorf("Error - VALIDATION: Unlocked amount %v is not matching to unlocked amount in instruction %v", iRes.UnlockedAmount, unlockAmountFromInst)
			continue
		}

		key, err := wallet.Base58CheckDeserialize(custodianAddrStrFromInst)
		if err != nil {
			Logger.log.Errorf("WARNING - VALIDATION: an error occured while deserializing custodian address string: ", err)
			continue
		}

		PRVIDStr := common.PRVCoinID.String()
		if iRes.MintedTokenID != PRVIDStr {
			Logger.log.Errorf("WARNING - VALIDATION: Minted token Id must be PRV: ")
			continue
		}

		_, pk, paidAmount, assetID := tx.GetTransferData()
		if !bytes.Equal(key.KeySet.PaymentAddress.Pk[:], pk[:]) ||
			unlockAmountFromInst != paidAmount ||
			PRVIDStr != assetID.String() {
			continue
		}
		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, fmt.Errorf(fmt.Sprintf("no PortalReqUnlockCollateral instruction found for PortalReqUnlockCollateralResponse tx %s", tx.Hash().String()))
	}
	instUsed[idx] = 1
	return true, nil
}
