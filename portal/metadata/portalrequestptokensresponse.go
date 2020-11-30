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

type PortalRequestPTokensResponse struct {
	basemeta.MetadataBase
	RequestStatus    string
	ReqTxID          common.Hash
	RequesterAddrStr string
	Amount           uint64
	IncTokenID       string
}

func NewPortalRequestPTokensResponse(
	depositStatus string,
	reqTxID common.Hash,
	requesterAddressStr string,
	amount uint64,
	tokenID string,
	metaType int,
) *PortalRequestPTokensResponse {
	metadataBase := basemeta.MetadataBase{
		Type: metaType,
	}
	return &PortalRequestPTokensResponse{
		RequestStatus:    depositStatus,
		ReqTxID:          reqTxID,
		MetadataBase:     metadataBase,
		RequesterAddrStr: requesterAddressStr,
		Amount:           amount,
		IncTokenID:       tokenID,
	}
}

func (iRes PortalRequestPTokensResponse) CheckTransactionFee(tr basemeta.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes PortalRequestPTokensResponse) ValidateTxWithBlockChain(txr basemeta.Transaction, chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, shardID byte, db *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (iRes PortalRequestPTokensResponse) ValidateSanityData(chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, beaconHeight uint64, txr basemeta.Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes PortalRequestPTokensResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return iRes.Type == basemeta.PortalUserRequestPTokenResponseMeta
}

func (iRes PortalRequestPTokensResponse) Hash() *common.Hash {
	record := iRes.MetadataBase.Hash().String()
	record += iRes.RequestStatus
	record += iRes.ReqTxID.String()
	record += iRes.RequesterAddrStr
	record += strconv.FormatUint(iRes.Amount, 10)
	record += iRes.IncTokenID
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *PortalRequestPTokensResponse) CalculateSize() uint64 {
	return basemeta.CalculateSize(iRes)
}

func (iRes PortalRequestPTokensResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
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
		if instUsed[i] > 0 ||
			instMetaType != strconv.Itoa(basemeta.PortalUserRequestPTokenMeta) {
			continue
		}
		instDepositStatus := inst[2]
		if instDepositStatus != iRes.RequestStatus ||
			(instDepositStatus != common.PortalReqPTokensAcceptedChainStatus) {
			continue
		}

		var shardIDFromInst byte
		var txReqIDFromInst common.Hash
		var requesterAddrStrFromInst string
		var portingAmountFromInst uint64
		var tokenIDStrFromInst string

		contentBytes := []byte(inst[3])
		var reqPTokensContent PortalRequestPTokensContent
		err := json.Unmarshal(contentBytes, &reqPTokensContent)
		if err != nil {
			Logger.log.Error("WARNING - VALIDATION: an error occured while parsing portal request ptokens content: ", err)
			continue
		}
		shardIDFromInst = reqPTokensContent.ShardID
		txReqIDFromInst = reqPTokensContent.TxReqID
		requesterAddrStrFromInst = reqPTokensContent.IncogAddressStr
		portingAmountFromInst = reqPTokensContent.PortingAmount
		tokenIDStrFromInst = reqPTokensContent.TokenID

		if !bytes.Equal(iRes.ReqTxID[:], txReqIDFromInst[:]) ||
			shardID != shardIDFromInst {
			continue
		}
		key, err := wallet.Base58CheckDeserialize(requesterAddrStrFromInst)
		if err != nil {
			Logger.log.Info("WARNING - VALIDATION: an error occured while deserializing receiver address string: ", err)
			continue
		}

		_, pk, paidAmount, assetID := tx.GetTransferData()
		if !bytes.Equal(key.KeySet.PaymentAddress.Pk[:], pk[:]) ||
			portingAmountFromInst != paidAmount ||
			tokenIDStrFromInst != assetID.String() {
			continue
		}
		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, fmt.Errorf(fmt.Sprintf("no PortalReqPtokens instruction found for PortalReqPtokensResponse tx %s", tx.Hash().String()))
	}
	instUsed[idx] = 1
	return true, nil
}
