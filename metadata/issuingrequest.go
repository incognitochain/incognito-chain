package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	privacy "github.com/ninjadotorg/constant/privacy"
)

type IssuingRequest struct {
	ReceiverAddress privacy.PaymentAddress
	DepositedAmount uint64
	AssetType       common.Hash // token id (one of types: Constant, BANK)
	CurrencyType    common.Hash // USD or ETH for now
	MetadataBase
}

func NewIssuingRequest(
	receiverAddress privacy.PaymentAddress,
	depositedAmount uint64,
	assetType common.Hash,
	currencyType common.Hash,
	metaType int,
) *IssuingRequest {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	issuingReq := &IssuingRequest{
		ReceiverAddress: receiverAddress,
		DepositedAmount: depositedAmount,
		AssetType:       assetType,
		CurrencyType:    currencyType,
	}
	issuingReq.MetadataBase = metadataBase
	return issuingReq
}

func (iReq *IssuingRequest) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	if !bytes.Equal(txr.GetSigPubKey(), common.CentralizedWebsitePubKey) {
		return false, errors.New("The issuance request must be called by centralized website.")
	}
	return true, nil
}

func (iReq *IssuingRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if len(iReq.ReceiverAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's receiver address")
	}
	if iReq.DepositedAmount == 0 {
		return false, false, errors.New("Wrong request info's deposited amount")
	}
	if iReq.Type == IssuingRequestMeta {
		return false, false, errors.New("Wrong request info's meta type")
	}
	if len(iReq.AssetType) != common.HashSize {
		return false, false, errors.New("Wrong request info's asset type")
	}
	return true, true, nil
}

func (iReq *IssuingRequest) ValidateMetadataByItself() bool {
	if iReq.Type != IssuingRequestMeta {
		return false
	}
	if !bytes.Equal(iReq.CurrencyType[:], common.USDAssetID[:]) &&
		!bytes.Equal(iReq.CurrencyType[:], common.ETHAssetID[:]) {
		return false
	}
	if !bytes.Equal(iReq.AssetType[:], common.ConstantID[:]) &&
		!bytes.Equal(iReq.AssetType[:], common.DCBTokenID[:]) {
		return false
	}
	if bytes.Equal(iReq.CurrencyType[:], common.ETHAssetID[:]) &&
		!bytes.Equal(iReq.AssetType[:], common.DCBTokenID[:]) {
		return false
	}
	return true
}

func (iReq *IssuingRequest) Hash() *common.Hash {
	record := iReq.ReceiverAddress.String()
	record += iReq.AssetType.String()
	record += iReq.CurrencyType.String()
	record += string(iReq.DepositedAmount)
	record += iReq.MetadataBase.Hash().String()

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (iReq *IssuingRequest) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	actionContent := map[string]interface{}{
		"txReqId": *(tx.Hash()),
		"meta":    *iReq,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(IssuingRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}
