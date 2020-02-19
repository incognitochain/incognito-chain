package metadata

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"strconv"
)

type PortalRequestPTokensResponse struct {
	MetadataBase
	RequestStatus    string
	ReqTxID          common.Hash
	RequesterAddrStr string
	Amount           uint64
}

func NewPortalRequestPTokensResponse(
	depositStatus string,
	reqTxID common.Hash,
	requesterAddressStr string,
	amount uint64,
	metaType int,
) *PortalRequestPTokensResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &PortalRequestPTokensResponse{
		RequestStatus:    depositStatus,
		ReqTxID:          reqTxID,
		MetadataBase:     metadataBase,
		RequesterAddrStr: requesterAddressStr,
		Amount:           amount,
	}
}

func (iRes PortalRequestPTokensResponse) CheckTransactionFee(tr Transaction, minFee uint64, beaconHeight int64, db database.DatabaseInterface) bool {
	// no need to have fee for this tx
	return true
}

func (iRes PortalRequestPTokensResponse) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (iRes PortalRequestPTokensResponse) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes PortalRequestPTokensResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return iRes.Type == PortalUserRequestPTokenResponseMeta
}

func (iRes PortalRequestPTokensResponse) Hash() *common.Hash {
	record := iRes.MetadataBase.Hash().String()
	record += iRes.RequestStatus
	record += iRes.ReqTxID.String()
	record += iRes.RequesterAddrStr
	record += strconv.FormatUint(iRes.Amount, 10)
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *PortalRequestPTokensResponse) CalculateSize() uint64 {
	return calculateSize(iRes)
}

//todo:
func (iRes PortalRequestPTokensResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
	txsInBlock []Transaction,
	txsUsed []int,
	insts [][]string,
	instUsed []int,
	shardID byte,
	tx Transaction,
	bcr BlockchainRetriever,
	ac *AccumulatedValues,
) (bool, error) {
	//idx := -1
	//for i, inst := range insts {
	//	if len(inst) < 4 { // this is not PortalCustodianDeposit instruction
	//		continue
	//	}
	//	instMetaType := inst[0]
	//	if instUsed[i] > 0 ||
	//		instMetaType != strconv.Itoa(PDEContributionMeta) {
	//		continue
	//	}
	//	instDepositStatus := inst[2]
	//	if instDepositStatus != iRes.RequestStatus ||
	//		(instDepositStatus != common.PortalCustodianDepositAcceptedChainStatus &&
	//			instDepositStatus != common.PortalCustodianDepositRefundChainStatus) {
	//		continue
	//	}
	//
	//	var shardIDFromInst byte
	//	var txReqIDFromInst common.Hash
	//	var receiverAddrStrFromInst string
	//	var receivingAmtFromInst uint64
	//	var receivingTokenIDStr string
	//
	//	if instDepositStatus == common.PDEContributionRefundChainStatus {
	//		contentBytes := []byte(inst[3])
	//		var refundContribution PDERefundContribution
	//		err := json.Unmarshal(contentBytes, &refundContribution)
	//		if err != nil {
	//			Logger.log.Error("WARNING - VALIDATION: an error occured while parsing refund contribution content: ", err)
	//			continue
	//		}
	//		shardIDFromInst = refundContribution.ShardID
	//		txReqIDFromInst = refundContribution.TxReqID
	//		receiverAddrStrFromInst = refundContribution.ContributorAddressStr
	//		receivingTokenIDStr = refundContribution.TokenIDStr
	//		receivingAmtFromInst = refundContribution.ContributedAmount
	//
	//	} else { // matched and returned
	//		contentBytes := []byte(inst[3])
	//		var matchedNReturnedContrib PDEMatchedNReturnedContribution
	//		err := json.Unmarshal(contentBytes, &matchedNReturnedContrib)
	//		if err != nil {
	//			Logger.log.Error("WARNING - VALIDATION: an error occured while parsing matched and returned contribution content: ", err)
	//			continue
	//		}
	//		shardIDFromInst = matchedNReturnedContrib.ShardID
	//		txReqIDFromInst = matchedNReturnedContrib.TxReqID
	//		receiverAddrStrFromInst = matchedNReturnedContrib.ContributorAddressStr
	//		receivingTokenIDStr = matchedNReturnedContrib.TokenIDStr
	//		receivingAmtFromInst = matchedNReturnedContrib.ReturnedContributedAmount
	//	}
	//
	//	if !bytes.Equal(iRes.ReqTxID[:], txReqIDFromInst[:]) ||
	//		shardID != shardIDFromInst {
	//		continue
	//	}
	//	key, err := wallet.Base58CheckDeserialize(receiverAddrStrFromInst)
	//	if err != nil {
	//		Logger.log.Info("WARNING - VALIDATION: an error occured while deserializing receiver address string: ", err)
	//		continue
	//	}
	//	_, pk, paidAmount, assetID := tx.GetTransferData()
	//	if !bytes.Equal(key.KeySet.PaymentAddress.Pk[:], pk[:]) ||
	//		receivingAmtFromInst != paidAmount ||
	//		receivingTokenIDStr != assetID.String() {
	//		continue
	//	}
	//	idx = i
	//	break
	//}
	//if idx == -1 { // not found the issuance request tx for this response
	//	return false, fmt.Errorf(fmt.Sprintf("no PDEContribution instruction found for PDEContributionResponse tx %s", tx.Hash().String()))
	//}
	//instUsed[idx] = 1
	return true, nil
}
