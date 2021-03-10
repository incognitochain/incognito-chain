package metadata

import (
	"bytes"
	"encoding/json"
	"fmt"
	portalcommonv4 "github.com/incognitochain/incognito-chain/portal/portalv4/common"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
)

type PortalV4UnshieldResponse struct {
	MetadataBase
	RequestStatus    string
	ReqTxID          common.Hash
	RequesterAddrStr string
	UnshieldAmount   uint64
	IncTokenID       string
}

func NewPortalV4UnshieldResponse(
	requestStatus string,
	reqTxID common.Hash,
	requesterAddressStr string,
	amount uint64,
	tokenID string,
	metaType int,
) *PortalV4UnshieldResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &PortalV4UnshieldResponse{
		RequestStatus:    requestStatus,
		ReqTxID:          reqTxID,
		MetadataBase:     metadataBase,
		RequesterAddrStr: requesterAddressStr,
		UnshieldAmount:   amount,
		IncTokenID:       tokenID,
	}
}

func (iRes PortalV4UnshieldResponse) CheckTransactionFee(tr Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes PortalV4UnshieldResponse) ValidateTxWithBlockChain(txr Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, db *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (iRes PortalV4UnshieldResponse) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, txr Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes PortalV4UnshieldResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return iRes.Type == PortalV4UnshieldingResponseMeta
}

func (iRes PortalV4UnshieldResponse) Hash() *common.Hash {
	record := iRes.MetadataBase.Hash().String()
	record += iRes.RequestStatus
	record += iRes.ReqTxID.String()
	record += iRes.RequesterAddrStr
	record += strconv.FormatUint(iRes.UnshieldAmount, 10)
	record += iRes.IncTokenID
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *PortalV4UnshieldResponse) CalculateSize() uint64 {
	return calculateSize(iRes)
}

func (iRes PortalV4UnshieldResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
	txsInBlock []Transaction,
	txsUsed []int,
	insts [][]string,
	instUsed []int,
	shardID byte,
	tx Transaction,
	chainRetriever ChainRetriever,
	ac *AccumulatedValues,
	shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever,
) (bool, error) {
	idx := -1
	for i, inst := range insts {
		if len(inst) < 4 { // this is not PortalV4UnshieldResponse instruction
			continue
		}
		instMetaType := inst[0]
		if instUsed[i] > 0 || (instMetaType != strconv.Itoa(PortalV4UnshieldingRequestMeta)) {
			continue
		}
		instReqStatus := inst[2]
		if iRes.RequestStatus != "rejected" ||
			(instReqStatus != portalcommonv4.PortalV4RequestRejectedChainStatus) {
			continue
		}

		var shardIDFromInst byte
		var txReqIDFromInst common.Hash
		var requesterAddrStrFromInst string
		var unshieldAmountFromInst uint64
		var tokenIDStrFromInst string

		contentBytes := []byte(inst[3])
		var unshieldReqContent PortalUnshieldRequestContent
		err := json.Unmarshal(contentBytes, &unshieldReqContent)
		if err != nil {
			Logger.log.Error("WARNING - VALIDATION: an error occured while parsing portal v4 unshield request content: ", err)
			continue
		}
		shardIDFromInst = unshieldReqContent.ShardID
		txReqIDFromInst = unshieldReqContent.TxReqID
		requesterAddrStrFromInst = unshieldReqContent.IncAddressStr
		unshieldAmountFromInst = unshieldReqContent.UnshieldAmount
		tokenIDStrFromInst = unshieldReqContent.TokenID

		if !bytes.Equal(iRes.ReqTxID[:], txReqIDFromInst[:]) ||
			shardID != shardIDFromInst {
			continue
		}
		if requesterAddrStrFromInst != iRes.RequesterAddrStr {
			Logger.log.Errorf("Error - VALIDATION: Requester address %v is not matching to Requester address in instruction %v", iRes.RequesterAddrStr, requesterAddrStrFromInst)
			continue
		}

		if unshieldAmountFromInst != iRes.UnshieldAmount {
			Logger.log.Errorf("Error - VALIDATION: Unshield amount %v is not matching to unshield amount in instruction %v", iRes.UnshieldAmount, unshieldAmountFromInst)
			continue
		}

		key, err := wallet.Base58CheckDeserialize(requesterAddrStrFromInst)
		if err != nil {
			Logger.log.Info("WARNING - VALIDATION: an error occured while deserializing requester address string: ", err)
			continue
		}

		_, pk, paidAmount, assetID := tx.GetTransferData()
		if !bytes.Equal(key.KeySet.PaymentAddress.Pk[:], pk[:]) ||
			unshieldAmountFromInst != paidAmount ||
			tokenIDStrFromInst != assetID.String() {
			continue
		}
		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, fmt.Errorf(fmt.Sprintf("no PortalV4UnshieldRequest instruction found for PortalV4UnshieldResponse tx %s", tx.Hash().String()))
	}
	instUsed[idx] = 1
	return true, nil
}
