package metadata

import (
	"fmt"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata/fromshardins"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/pkg/errors"
)

func NewSubmitProposalInfo(executeDuration uint64, explanation string, paymentAddress privacy.PaymentAddress, constitutionIndex uint32) *component.SubmitProposalInfo {
	return &component.SubmitProposalInfo{ExecuteDuration: executeDuration, Explanation: explanation, PaymentAddress: paymentAddress, ConstitutionIndex: constitutionIndex}
}

type SubmitDCBProposalMetadata struct {
	DCBParams          component.DCBParams
	SubmitProposalInfo component.SubmitProposalInfo

	MetadataBase
}

func NewSubmitDCBProposalMetadata(
	DCBParams component.DCBParams,
	executeDuration uint64,
	explanation string,
	address *privacy.PaymentAddress,
	constitutionIndex uint32,
) *SubmitDCBProposalMetadata {
	return &SubmitDCBProposalMetadata{
		DCBParams: DCBParams,
		SubmitProposalInfo: *NewSubmitProposalInfo(
			executeDuration,
			explanation,
			*address,
			constitutionIndex,
		),
		MetadataBase: *NewMetadataBase(SubmitDCBProposalMeta),
	}
}

func NewSubmitDCBProposalMetadataFromRPC(data map[string]interface{}) (Metadata, error) {
	dcbParams, err := component.NewDCBParamsFromJson(data["DCBParams"])
	if err != nil {
		return nil, err
	}
	return NewSubmitDCBProposalMetadata(
		*dcbParams,
		uint64(data["ExecuteDuration"].(float64)),
		data["Explanation"].(string),
		data["PaymentAddress"].(*privacy.PaymentAddress),
		uint32(data["ConstitutionIndex"].(float64)),
	), nil
}

func (submitDCBProposalMetadata *SubmitDCBProposalMetadata) Hash() *common.Hash {
	record := submitDCBProposalMetadata.DCBParams.Hash().String()
	record += string(submitDCBProposalMetadata.SubmitProposalInfo.ToBytes())

	record += submitDCBProposalMetadata.MetadataBase.Hash().String()
	hash := common.HashH([]byte(record))
	return &hash
}

func (submitDCBProposalMetadata *SubmitDCBProposalMetadata) BuildReqActions(
	tx Transaction,
	bcr BlockchainRetriever,
	shardID byte,
) ([][]string, error) {
	submitProposal := component.SubmitProposalData{
		ProposalTxID:      *tx.Hash(),
		ConstitutionIndex: submitDCBProposalMetadata.SubmitProposalInfo.ConstitutionIndex,
		SubmitterPayment:  submitDCBProposalMetadata.SubmitProposalInfo.PaymentAddress,
	}
	inst := fromshardins.NewSubmitProposalIns(common.DCBBoard, submitProposal)

	instStr, err := inst.GetStringFormat()
	fmt.Println("[ndh] - submitDCBProposalMetadata BuildReqActions: ", instStr)
	if err != nil {
		return nil, err
	}
	return [][]string{instStr}, nil
}

func (submitDCBProposalMetadata *SubmitDCBProposalMetadata) ValidateTxWithBlockChain(
	tx Transaction,
	br BlockchainRetriever,
	chainID byte,
	db database.DatabaseInterface,
) (bool, error) {
	if !submitDCBProposalMetadata.SubmitProposalInfo.ValidateTxWithBlockChain(common.DCBBoard, chainID, db) {
		return false, errors.Errorf("SubmitProposalInfo invalid")
	}

	fmt.Println("[db] validating dcb proposal")

	// Validate reserve data
	raiseReserveData := submitDCBProposalMetadata.DCBParams.RaiseReserveData
	for assetID, _ := range raiseReserveData {
		if br.GetAssetPrice(&assetID) == 0 {
			return false, errors.Errorf("cannot raise reserve without oracle price for asset %s", assetID.String())
		}
	}

	spendReserveData := submitDCBProposalMetadata.DCBParams.SpendReserveData
	for assetID, _ := range spendReserveData {
		if br.GetAssetPrice(&assetID) == 0 {
			return false, errors.Errorf("cannot spend reserve without oracle price for asset %s", assetID.String())
		}
	}
	return true, nil
}

func (submitDCBProposalMetadata *SubmitDCBProposalMetadata) ValidateSanityData(br BlockchainRetriever, tx Transaction) (bool, bool, error) {
	if !submitDCBProposalMetadata.DCBParams.ValidateSanityData() {
		return true, false, fmt.Errorf("invalid sanity data for DCB params")
	}
	if !submitDCBProposalMetadata.SubmitProposalInfo.ValidateSanityData() {
		return true, false, fmt.Errorf("invalid sanity data for proposal info")
	}
	return true, true, nil
}

func (submitDCBProposalMetadata *SubmitDCBProposalMetadata) ValidateMetadataByItself() bool {
	return true
}

type SubmitGOVProposalMetadata struct {
	GOVParams          component.GOVParams
	SubmitProposalInfo component.SubmitProposalInfo

	MetadataBase
}

func (submitGOVProposalMetadata *SubmitGOVProposalMetadata) BuildReqActions(
	tx Transaction,
	bcr BlockchainRetriever,
	shardID byte,
) ([][]string, error) {
	submitProposal := component.SubmitProposalData{
		ProposalTxID:      *tx.Hash(),
		ConstitutionIndex: submitGOVProposalMetadata.SubmitProposalInfo.ConstitutionIndex,
		SubmitterPayment:  submitGOVProposalMetadata.SubmitProposalInfo.PaymentAddress,
	}
	inst := fromshardins.NewSubmitProposalIns(common.GOVBoard, submitProposal)

	instStr, err := inst.GetStringFormat()
	fmt.Println("[ndh] - submitGOVProposalMetadata BuildReqActions: ", instStr)
	if err != nil {
		return nil, err
	}
	return [][]string{instStr}, nil
}

func NewSubmitGOVProposalMetadata(
	govParams component.GOVParams,
	executeDuration uint64,
	explanation string,
	address *privacy.PaymentAddress,
	constitutionIndex uint32,
) *SubmitGOVProposalMetadata {
	return &SubmitGOVProposalMetadata{
		GOVParams: govParams,
		SubmitProposalInfo: *NewSubmitProposalInfo(
			executeDuration,
			explanation,
			*address,
			constitutionIndex,
		),
		MetadataBase: *NewMetadataBase(SubmitGOVProposalMeta),
	}
}

func NewSubmitGOVProposalMetadataFromRPC(data map[string]interface{}) (Metadata, error) {
	return NewSubmitGOVProposalMetadata(
		*component.NewGOVParamsFromJson(data["GOVParams"]),
		uint64(data["ExecuteDuration"].(float64)),
		data["Explanation"].(string),
		data["PaymentAddress"].(*privacy.PaymentAddress),
		uint32(data["ConstitutionIndex"].(float64)),
	), nil
}

func (submitGOVProposalMetadata *SubmitGOVProposalMetadata) Hash() *common.Hash {
	record := submitGOVProposalMetadata.GOVParams.Hash().String()
	record += string(submitGOVProposalMetadata.SubmitProposalInfo.ToBytes())

	record += submitGOVProposalMetadata.MetadataBase.Hash().String()
	hash := common.HashH([]byte(record))
	return &hash
}

func (submitGOVProposalMetadata *SubmitGOVProposalMetadata) ValidateTxWithBlockChain(tx Transaction, br BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	beaconHeight := br.GetBeaconHeight()
	govParams := submitGOVProposalMetadata.GOVParams
	sellingBonds := govParams.SellingBonds
	if sellingBonds != nil {
		if sellingBonds.StartSellingAt+sellingBonds.Maturity < beaconHeight {
			return false, nil
		}
		if sellingBonds.StartSellingAt+sellingBonds.SellingWithin < beaconHeight {
			return false, nil
		}
	}

	sellingGOVTokens := govParams.SellingGOVTokens
	if sellingGOVTokens != nil {
		if sellingGOVTokens.StartSellingAt+sellingGOVTokens.SellingWithin < beaconHeight {
			return false, nil
		}
	}
	return true, nil
}

func (submitGOVProposalMetadata *SubmitGOVProposalMetadata) ValidateSanityData(br BlockchainRetriever, tx Transaction) (bool, bool, error) {
	if !submitGOVProposalMetadata.GOVParams.ValidateSanityData() {
		return true, false, errors.New("submitGOVProposalMetadata.GOVParams")
	}
	if !submitGOVProposalMetadata.SubmitProposalInfo.ValidateSanityData() {
		return true, false, errors.New("submitGOVProposalMetadata.SubmitProposalInfo")
	}
	return true, true, nil
}

func (submitGOVProposalMetadata *SubmitGOVProposalMetadata) ValidateMetadataByItself() bool {
	return true
}

func (submitGOVProposalMetadata *SubmitGOVProposalMetadata) CalculateSize() uint64 {
	return calculateSize(submitGOVProposalMetadata)
}
