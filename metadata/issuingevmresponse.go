package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wallet"
)

type IssuingEVMResponse struct {
	MetadataBase
	RequestedTxID   common.Hash
	UniqTx          []byte
	ExternalTokenID []byte
}

type IssuingEVMResAction struct {
	Meta       *IssuingEVMResponse `json:"meta"`
	IncTokenID *common.Hash        `json:"incTokenID"`
}

func NewIssuingEVMResponse(
	requestedTxID common.Hash,
	uniqTx []byte,
	externalTokenID []byte,
	metaType int,
) *IssuingEVMResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &IssuingEVMResponse{
		RequestedTxID:   requestedTxID,
		UniqTx:          uniqTx,
		ExternalTokenID: externalTokenID,
		MetadataBase:    metadataBase,
	}
}

func (iRes IssuingEVMResponse) CheckTransactionFee(tr Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes IssuingEVMResponse) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID) in current block
	return false, nil
}

func (iRes IssuingEVMResponse) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes IssuingEVMResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (iRes IssuingEVMResponse) Hash() *common.Hash {
	record := iRes.RequestedTxID.String()
	record += string(iRes.UniqTx)
	record += string(iRes.ExternalTokenID)
	record += iRes.MetadataBase.Hash().String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *IssuingEVMResponse) CalculateSize() uint64 {
	return calculateSize(iRes)
}

func (iRes IssuingEVMResponse) VerifyMinerCreatedTxBeforeGettingInBlock(txsInBlock []Transaction, txsUsed []int, insts [][]string, instUsed []int, shardID byte, tx Transaction, chainRetriever ChainRetriever, ac *AccumulatedValues, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever) (bool, error) {
	idx := -1
	for i, inst := range insts {
		if len(inst) < 4 { // this is not IssuingEVMRequest instruction
			continue
		}
		instMetaType := inst[0]
		if instUsed[i] > 0 ||
			(instMetaType != strconv.Itoa(IssuingETHRequestMeta) && instMetaType != strconv.Itoa(IssuingBSCRequestMeta)) {
			continue
		}

		contentBytes, err := base64.StdEncoding.DecodeString(inst[3])
		if err != nil {
			Logger.log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
			continue
		}
		var issuingETHAcceptedInst IssuingEVMAcceptedInst
		err = json.Unmarshal(contentBytes, &issuingETHAcceptedInst)
		if err != nil {
			Logger.log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
			continue
		}

		if !bytes.Equal(iRes.RequestedTxID[:], issuingETHAcceptedInst.TxReqID[:]) ||
			!bytes.Equal(iRes.UniqTx, issuingETHAcceptedInst.UniqTx) ||
			!bytes.Equal(iRes.ExternalTokenID, issuingETHAcceptedInst.ExternalTokenID) ||
			shardID != issuingETHAcceptedInst.ShardID {
			continue
		}

		addressStr := issuingETHAcceptedInst.ReceiverAddrStr
		key, err := wallet.Base58CheckDeserialize(addressStr)
		if err != nil {
			Logger.log.Info("WARNING - VALIDATION: an error occured while deserializing receiver address string: ", err)
			continue
		}
		_, pk, paidAmount, assetID := tx.GetTransferData()
		if !bytes.Equal(key.KeySet.PaymentAddress.Pk[:], pk[:]) ||
			issuingETHAcceptedInst.IssuingAmount != paidAmount ||
			!bytes.Equal(issuingETHAcceptedInst.IncTokenID[:], assetID[:]) {
			continue
		}
		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, errors.New(fmt.Sprintf("no IssuingETHRequest tx found for IssuingEVMResponse tx %s", tx.Hash().String()))
	}
	instUsed[idx] = 1
	return true, nil
}
