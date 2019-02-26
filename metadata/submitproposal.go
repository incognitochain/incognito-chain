package metadata

import (
	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/pkg/errors"
)

type SubmitProposalInfo struct {
	ExecuteDuration uint64
	Explanation     string
	PaymentAddress  privacy.PaymentAddress
}

func (submitProposalInfo SubmitProposalInfo) ToBytes() []byte {
	record := string(common.Uint64ToBytes(submitProposalInfo.ExecuteDuration))
	record += submitProposalInfo.Explanation
	record += string(submitProposalInfo.PaymentAddress.Bytes())
	return []byte(record)
}

func (submitProposalInfo SubmitProposalInfo) ValidateSanityData(
	br BlockchainRetriever,
	tx Transaction,
) bool {
	if submitProposalInfo.ExecuteDuration < common.MinimumBlockOfProposalDuration ||
		submitProposalInfo.ExecuteDuration > common.MaximumBlockOfProposalDuration {
		return false
	}
	if len(submitProposalInfo.Explanation) > common.MaximumProposalExplainationLength {
		return false
	}
	return true
}

func (submitProposalInfo SubmitProposalInfo) ValidateTxWithBlockChain(
	boardType BoardType,
	tx Transaction,
	br BlockchainRetriever,
	chainID byte,
	db database.DatabaseInterface,
) bool {
	if br.GetConstitutionEndHeight(DCBBoard, chainID)+submitProposalInfo.ExecuteDuration+common.MinimumBlockOfProposalDuration >
		br.GetBoardEndHeight(boardType, chainID) {
		return false
	}
	return true
}

func NewSubmitProposalInfo(
	executeDuration uint64,
	explanation string,
	paymentAddress privacy.PaymentAddress,
) *SubmitProposalInfo {
	return &SubmitProposalInfo{
		ExecuteDuration: executeDuration,
		Explanation:     explanation,
		PaymentAddress:  paymentAddress,
	}
}

type SubmitDCBProposalMetadata struct {
	DCBParams          params.DCBParams
	SubmitProposalInfo SubmitProposalInfo

	MetadataBase
}

func NewSubmitDCBProposalMetadata(DCBParams params.DCBParams, executeDuration uint64, explanation string, address *privacy.PaymentAddress) *SubmitDCBProposalMetadata {
	return &SubmitDCBProposalMetadata{
		DCBParams: DCBParams,
		SubmitProposalInfo: *NewSubmitProposalInfo(
			executeDuration,
			explanation,
			*address,
		),
		MetadataBase: *NewMetadataBase(SubmitDCBProposalMeta),
	}
}

func NewSubmitDCBProposalMetadataFromRPC(data map[string]interface{}) (Metadata, error) {
	meta := NewSubmitDCBProposalMetadata(
		*params.NewDCBParamsFromJson(data["DCBParams"]),
		uint64(data["ExecuteDuration"].(float64)),
		data["Explanation"].(string),
		data["PaymentAddress"].(*privacy.PaymentAddress),
	)
	return meta, nil
}

func (submitDCBProposalMetadata *SubmitDCBProposalMetadata) Hash() *common.Hash {
	record := submitDCBProposalMetadata.DCBParams.Hash().String()
	record += string(submitDCBProposalMetadata.SubmitProposalInfo.ToBytes())

	record += submitDCBProposalMetadata.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (submitDCBProposalMetadata *SubmitDCBProposalMetadata) ValidateTxWithBlockChain(
	tx Transaction,
	br BlockchainRetriever,
	chainID byte,
	db database.DatabaseInterface,
) (bool, error) {
	if !submitDCBProposalMetadata.SubmitProposalInfo.ValidateTxWithBlockChain(DCBBoard, tx, br, chainID, db) {
		return false, nil
	}

	raiseReserveData := submitDCBProposalMetadata.DCBParams.RaiseReserveData
	for assetID, _ := range raiseReserveData {
		if br.GetAssetPrice(&assetID) == 0 {
			return false, errors.Errorf("Cannot raise reserve without oracle price for asset %x", assetID)
		}
	}

	spendReserveData := submitDCBProposalMetadata.DCBParams.SpendReserveData
	for assetID, _ := range spendReserveData {
		if br.GetAssetPrice(&assetID) == 0 {
			return false, errors.Errorf("Cannot spend reserve without oracle price for asset %x", assetID)
		}
	}
	return true, nil
}

func (submitDCBProposalMetadata *SubmitDCBProposalMetadata) ValidateSanityData(br BlockchainRetriever, tx Transaction) (bool, bool, error) {
	if !submitDCBProposalMetadata.DCBParams.ValidateSanityData() {
		return true, false, nil
	}
	if !submitDCBProposalMetadata.SubmitProposalInfo.ValidateSanityData(br, tx) {
		return true, false, nil
	}
	return true, true, nil
}

func (submitDCBProposalMetadata *SubmitDCBProposalMetadata) ValidateMetadataByItself() bool {
	return true
}

type SubmitGOVProposalMetadata struct {
	GOVParams          params.GOVParams
	SubmitProposalInfo SubmitProposalInfo

	MetadataBase
}

func NewSubmitGOVProposalMetadata(
	govParams params.GOVParams,
	executeDuration uint64,
	explanation string,
	address *privacy.PaymentAddress,
) *SubmitGOVProposalMetadata {
	return &SubmitGOVProposalMetadata{
		GOVParams: govParams,
		SubmitProposalInfo: *NewSubmitProposalInfo(
			executeDuration,
			explanation,
			*address,
		),
		MetadataBase: *NewMetadataBase(SubmitGOVProposalMeta),
	}
}

func NewSubmitGOVProposalMetadataFromRPC(data map[string]interface{}) (Metadata, error) {
	return NewSubmitGOVProposalMetadata(
		*params.NewGOVParamsFromJson(data["GOVParams"]),
		uint64(data["ExecuteDuration"].(float64)),
		data["Explanation"].(string),
		data["PaymentAddress"].(*privacy.PaymentAddress),
	), nil
}

func (submitGOVProposalMetadata *SubmitGOVProposalMetadata) Hash() *common.Hash {
	record := submitGOVProposalMetadata.GOVParams.Hash().String()
	record += string(submitGOVProposalMetadata.SubmitProposalInfo.ToBytes())

	record += submitGOVProposalMetadata.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (submitGOVProposalMetadata *SubmitGOVProposalMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte, database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (submitGOVProposalMetadata *SubmitGOVProposalMetadata) ValidateSanityData(br BlockchainRetriever, tx Transaction) (bool, bool, error) {
	if !submitGOVProposalMetadata.GOVParams.ValidateSanityData() {
		return true, false, nil
	}
	if !submitGOVProposalMetadata.SubmitProposalInfo.ValidateSanityData(br, tx) {
		return true, false, nil
	}
	return true, true, nil
}

func (submitGOVProposalMetadata *SubmitGOVProposalMetadata) ValidateMetadataByItself() bool {
	return true
}
