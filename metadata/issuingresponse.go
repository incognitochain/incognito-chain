package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/basemeta"

	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/pkg/errors"
)

type IssuingResponse struct {
	basemeta.MetadataBase
	RequestedTxID common.Hash
}

type IssuingResAction struct {
	IncTokenID *common.Hash `json:"incTokenID"`
}

func NewIssuingResponse(requestedTxID common.Hash, metaType int) *IssuingResponse {
	metadataBase := basemeta.MetadataBase{
		Type: metaType,
	}
	return &IssuingResponse{
		RequestedTxID: requestedTxID,
		MetadataBase:  metadataBase,
	}
}

func (iRes IssuingResponse) CheckTransactionFee(tr basemeta.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes IssuingResponse) ValidateTxWithBlockChain(tx basemeta.Transaction, chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID) in current block
	return false, nil
}

func (iRes IssuingResponse) ValidateSanityData(chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, beaconHeight uint64, tx basemeta.Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes IssuingResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (iRes IssuingResponse) Hash() *common.Hash {
	record := iRes.RequestedTxID.String()
	record += iRes.MetadataBase.Hash().String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *IssuingResponse) CalculateSize() uint64 {
	return basemeta.CalculateSize(iRes)
}

func (iRes IssuingResponse) VerifyMinerCreatedTxBeforeGettingInBlock(txsInBlock []basemeta.Transaction, txsUsed []int, insts [][]string, instUsed []int, shardID byte, tx basemeta.Transaction, chainRetriever basemeta.ChainRetriever, ac *basemeta.AccumulatedValues, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever) (bool, error) {
	idx := -1
	for i, inst := range insts {
		if len(inst) < 4 { // this is not IssuingETHRequest instruction
			continue
		}
		instMetaType := inst[0]
		if instUsed[i] > 0 ||
			instMetaType != strconv.Itoa(basemeta.IssuingRequestMeta) {
			continue
		}

		contentBytes, err := base64.StdEncoding.DecodeString(inst[3])
		if err != nil {
			Logger.log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
			continue
		}
		var issuingAcceptedInst IssuingAcceptedInst
		err = json.Unmarshal(contentBytes, &issuingAcceptedInst)
		if err != nil {
			Logger.log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
			continue
		}

		if !bytes.Equal(iRes.RequestedTxID[:], issuingAcceptedInst.TxReqID[:]) {
			continue
		}

		_, pk, amount, assetID := tx.GetTransferData()
		if !bytes.Equal(issuingAcceptedInst.ReceiverAddr.Pk[:], pk[:]) ||
			issuingAcceptedInst.DepositedAmount != amount ||
			!bytes.Equal(issuingAcceptedInst.IncTokenID[:], assetID[:]) ||
			issuingAcceptedInst.ShardID != shardID {
			continue
		}
		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, errors.Errorf("no IssuingRequest tx found for the IssuingResponse tx %s", tx.Hash().String())
	}
	instUsed[idx] = 1
	return true, nil
}
