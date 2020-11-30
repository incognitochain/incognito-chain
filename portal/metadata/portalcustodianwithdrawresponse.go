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

type PortalCustodianWithdrawResponse struct {
	basemeta.MetadataBase
	RequestStatus  string
	ReqTxID        common.Hash
	PaymentAddress string
	Amount         uint64
}

func NewPortalCustodianWithdrawResponse(
	requestStatus string,
	reqTxId common.Hash,
	paymentAddress string,
	amount uint64,
	metaType int,
) *PortalCustodianWithdrawResponse {
	metaDataBase := basemeta.MetadataBase{Type: metaType}

	return &PortalCustodianWithdrawResponse{
		MetadataBase:   metaDataBase,
		RequestStatus:  requestStatus,
		ReqTxID:        reqTxId,
		PaymentAddress: paymentAddress,
		Amount:         amount,
	}
}

func (responseMeta PortalCustodianWithdrawResponse) CheckTransactionFee(tr basemeta.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (responseMeta PortalCustodianWithdrawResponse) ValidateTxWithBlockChain(txr basemeta.Transaction, chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, shardID byte, db *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (responseMeta PortalCustodianWithdrawResponse) ValidateSanityData(chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, beaconHeight uint64, tx basemeta.Transaction) (bool, bool, error) {
	return false, true, nil
}

func (responseMeta PortalCustodianWithdrawResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return responseMeta.Type == basemeta.PortalCustodianWithdrawResponseMeta
}

func (responseMeta PortalCustodianWithdrawResponse) Hash() *common.Hash {
	record := responseMeta.MetadataBase.Hash().String()
	record += responseMeta.RequestStatus
	record += responseMeta.ReqTxID.String()
	record += responseMeta.PaymentAddress
	record += strconv.FormatUint(responseMeta.Amount, 10)
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (responseMeta *PortalCustodianWithdrawResponse) CalculateSize() uint64 {
	return basemeta.CalculateSize(responseMeta)
}

func (responseMeta PortalCustodianWithdrawResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
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
		if len(inst) < 4 { // this is not PortalRequestPTokens response instruction
			continue
		}

		instMetaType := inst[0]
		if instUsed[i] > 0 || instMetaType != strconv.Itoa(basemeta.PortalCustodianWithdrawRequestMeta) {
			continue
		}

		instDepositStatus := inst[2]
		if instDepositStatus != responseMeta.RequestStatus ||
			(instDepositStatus != common.PortalCustodianWithdrawRequestAcceptedChainStatus) {
			continue
		}

		var shardIDFromInst byte
		var txReqIDFromInst common.Hash
		var requesterAddrStrFromInst string
		var portingAmountFromInst uint64

		contentBytes := []byte(inst[3])
		var custodianWithdrawRequest PortalCustodianWithdrawRequestContent
		err := json.Unmarshal(contentBytes, &custodianWithdrawRequest)
		if err != nil {
			Logger.log.Error("WARNING - VALIDATION: an error occured while parsing custodian withdraw request content: ", err)
			continue
		}
		shardIDFromInst = custodianWithdrawRequest.ShardID
		txReqIDFromInst = custodianWithdrawRequest.TxReqID
		requesterAddrStrFromInst = custodianWithdrawRequest.PaymentAddress
		portingAmountFromInst = custodianWithdrawRequest.Amount
		receivingTokenIDStr := common.PRVCoinID.String()

		if !bytes.Equal(responseMeta.ReqTxID[:], txReqIDFromInst[:]) ||
			shardID != shardIDFromInst {
			continue
		}

		key, err := wallet.Base58CheckDeserialize(requesterAddrStrFromInst)
		if err != nil {
			Logger.log.Info("WARNING - VALIDATION: an error occured while deserializing receiver address string: ", err)
			continue
		}

		_, pk, amount, assetID := tx.GetTransferData()
		if !bytes.Equal(key.KeySet.PaymentAddress.Pk[:], pk[:]) ||
			portingAmountFromInst != amount ||
			receivingTokenIDStr != assetID.String() {
			continue
		}

		idx = i
		break
	}

	if idx == -1 { // not found the issuance request tx for this response
		return false, fmt.Errorf(fmt.Sprintf("no PortalCustodianWithdrawRequest instruction found for PortalCustodianWithdrawResponse tx %s", tx.Hash().String()))
	}
	instUsed[idx] = 1
	return true, nil
}
