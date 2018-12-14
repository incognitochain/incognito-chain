package metadata

import (
	"errors"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	privacy "github.com/ninjadotorg/constant/privacy-protocol"
)

type IssuingRequest struct {
	ReceiverAddress privacy.PaymentAddress
	DepositedAmount uint64      // in US dollar
	AssetType       common.Hash // token id (one of types: Constant, BANK)
	// TODO: need an ID to verify with PrimeTrust
	MetadataBase
}

func NewIssuingRequest(
	receiverAddress privacy.PaymentAddress,
	depositedAmount uint64,
	assetType common.Hash,
	metaType int,
) *IssuingRequest {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	issuingReq := &IssuingRequest{
		ReceiverAddress: receiverAddress,
		DepositedAmount: depositedAmount,
		AssetType:       assetType,
	}
	issuingReq.MetadataBase = metadataBase
	return issuingReq
}

func (iReq *IssuingRequest) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	chainID byte,
	db database.DatabaseInterface,
) (bool, error) {
	// check double spending on fee tx
	err := txr.ValidateConstDoubleSpendWithBlockchain(bcr, chainID, db)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (iReq *IssuingRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	ok, err := txr.ValidateSanityData(bcr)
	if err != nil || !ok {
		return false, ok, err
	}
	if len(iReq.ReceiverAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's receiver address")
	}
	if iReq.DepositedAmount == 0 {
		return false, false, errors.New("Wrong request info's deposited amount")
	}
	if iReq.Type == IssuingRequestMeta {
		return false, false, errors.New("Wrong request info's meta type")
	}
	assetTypeStr := string(iReq.AssetType[:])
	if assetTypeStr != string(common.ConstantID[:]) && assetTypeStr != string(common.DCBTokenID[:]) {
		return false, false, errors.New("Wrong request info's asset type")
	}
	return false, true, nil
}

func (iReq *IssuingRequest) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (iReq *IssuingRequest) Hash() *common.Hash {
	record := string(iReq.ReceiverAddress.ToBytes())
	record += iReq.AssetType.String()
	record += string(iReq.DepositedAmount)
	record += string(iReq.MetadataBase.Hash()[:])

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}
