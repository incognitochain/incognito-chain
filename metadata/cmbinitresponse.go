package metadata

import (
	"bytes"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/pkg/errors"
)

type CMBInitResponse struct {
	MainAccount privacy.PaymentAddress

	MetadataBase
}

func NewCMBInitResponse(data map[string]interface{}) *CMBInitResponse {
	mainKey, err := wallet.Base58CheckDeserialize(data["MainAccount"].(string))
	if err != nil {
		return nil
	}
	result := CMBInitResponse{
		MainAccount: mainKey.KeySet.PaymentAddress,
	}
	result.Type = CMBInitResponseMeta
	return &result
}

func (cres *CMBInitResponse) Hash() *common.Hash {
	record := string(cres.MainAccount.Bytes())

	// final hash
	record += cres.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (cres *CMBInitResponse) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	// Check if cmb init request existed
	_, _, _, txHash, state, _, err := bcr.GetCMB(cres.MainAccount.Bytes())
	if err != nil {
		return common.FalseValue, err
	}

	// Check state of cmb
	if state != CMBRequested {
		return common.FalseValue, errors.Errorf("cmb state is not CMBRequested, cannot create response")
	}

	// Check if only board members created this tx
	if !txCreatedByDCBBoardMember(txr, bcr) {
		return common.FalseValue, errors.Errorf("Tx must be created by DCB Governor")
	}

	// Check if this member hasn't responded to this request
	memberResponded, err := bcr.GetCMBResponse(cres.MainAccount.Bytes())
	if err != nil {
		return common.FalseValue, errors.Errorf("error getting list of old cmb init responses")
	}
	for _, member := range memberResponded {
		if bytes.Equal(txr.GetSigPubKey(), member) {
			return common.FalseValue, errors.Errorf("each board member can only response once to each cmb init request")
		}
	}

	// Check if response time is not over
	_, blockHash, _, _, err := bcr.GetTransactionByHash(txHash)
	if err != nil {
		return common.FalseValue, err
	}
	reqBlockHeight, _, _ := bcr.GetBlockHeightByBlockHash(blockHash)
	curBlockHeight, err := bcr.GetTxChainHeight(txr)
	if err != nil || curBlockHeight-reqBlockHeight >= CMBInitRefundPeriod {
		return common.FalseValue, errors.Errorf("response time is over for this cmb init request")
	}
	return common.TrueValue, nil
}

func (cres *CMBInitResponse) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	// TODO(@0xbunyip)
	return common.TrueValue, common.FalseValue, nil // DCB takes care of fee
}

func (cres *CMBInitResponse) ValidateMetadataByItself() bool {
	// TODO(@0xbunyip)
	return common.TrueValue
}
