package metadata

import (
	"bytes"
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
	if bytes.Equal(iReq.AssetType[:], common.DCBTokenID[:]) {
		saleDBCTOkensByUSDData := bcr.GetDCBParams().SaleDBCTOkensByUSDData
		if bcr.GetHeight()+1 > saleDBCTOkensByUSDData.EndBlock {
			return false, nil
		}
		oracleParams := bcr.GetOracleParams()
		reqAmt := iReq.DepositedAmount / oracleParams.DCBToken
		if saleDBCTOkensByUSDData.Amount < reqAmt {
			return false, nil
		}
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
	if len(iReq.AssetType) != common.HashSize {
		return false, false, errors.New("Wrong request info's asset type")
	}
	return false, true, nil
}

func (iReq *IssuingRequest) ValidateMetadataByItself() bool {
	if iReq.Type != IssuingRequestMeta {
		return false
	}
	if !bytes.Equal(iReq.AssetType[:], common.DCBTokenID[:]) &&
		!bytes.Equal(iReq.AssetType[:], common.ConstantID[:]) {
		return false
	}
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
