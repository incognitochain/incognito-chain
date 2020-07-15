package metadata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
	"strconv"
)

type PortalCustodianWithdrawResponse struct {
	MetadataBase
	RequestStatus  string
	ReqTxID        common.Hash
	PaymentAddress string
	Amount         uint64
	SharedRandom       []byte
}

func NewPortalCustodianWithdrawResponse(
	requestStatus string,
	reqTxId common.Hash,
	paymentAddress string,
	amount uint64,
	metaType int,
) *PortalCustodianWithdrawResponse {
	metaDataBase := MetadataBase{Type: metaType}

	return &PortalCustodianWithdrawResponse{
		MetadataBase:   metaDataBase,
		RequestStatus:  requestStatus,
		ReqTxID:        reqTxId,
		PaymentAddress: paymentAddress,
		Amount:         amount,
	}
}

func (responseMeta PortalCustodianWithdrawResponse) CheckTransactionFee(tr Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (responseMeta PortalCustodianWithdrawResponse) ValidateTxWithBlockChain(txr Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, db *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (responseMeta PortalCustodianWithdrawResponse) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	return false, true, nil
}

func (responseMeta PortalCustodianWithdrawResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return responseMeta.Type == PortalCustodianWithdrawResponseMeta
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
	return calculateSize(responseMeta)
}

func (responseMeta PortalCustodianWithdrawResponse) VerifyMinerCreatedTxBeforeGettingInBlock(mintData *MintData, shardID byte, tx Transaction, chainRetriever ChainRetriever, ac *AccumulatedValues, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever) (bool, error) {
	idx := -1

	for i, inst := range mintData.Insts {
		if len(inst) < 4 { // this is not PortalRequestPTokens response instruction
			continue
		}

		instMetaType := inst[0]
		if mintData.InstsUsed[i] > 0 || instMetaType != strconv.Itoa(PortalCustodianWithdrawRequestMeta) {
			continue
		}

		instDepositStatus := inst[2]
		if instDepositStatus != responseMeta.RequestStatus ||
			(instDepositStatus != common.PortalCustodianWithdrawRequestAcceptedStatus) {
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
			Logger.log.Info("WARNING - VALIDATION: Error occured while deserializing receiver address string: ", err)
			continue
		}

		isMinted, mintCoin, coinID, err := tx.GetTxMintData()
		if err != nil || !isMinted {
			Logger.log.Info("WARNING - VALIDATION: Error occured while validate tx mint.  ", err)
			continue
		}

		if coinID.String() != receivingTokenIDStr {
			Logger.log.Info("WARNING - VALIDATION: Receive Token ID in tx mint maybe not correct.")
			continue
		}
		if ok := mintCoin.CheckCoinValid(key.KeySet.PaymentAddress, responseMeta.SharedRandom, portingAmountFromInst); !ok {
			Logger.log.Info("WARNING - VALIDATION: Error occured while check receiver and amount. CheckCoinValid return false ")
			continue
		}

		idx = i
		break
	}

	if idx == -1 { // not found the issuance request tx for this response
		return false, fmt.Errorf(fmt.Sprintf("no PortalCustodianWithdrawRequest instruction found for PortalCustodianWithdrawResponse tx %s", tx.Hash().String()))
	}
	mintData.InstsUsed[idx] = 1
	return true, nil
}

func (responseMeta *PortalCustodianWithdrawResponse) SetSharedRandom(r []byte) {
	responseMeta.SharedRandom = r
}