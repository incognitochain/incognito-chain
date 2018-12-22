package metadata

import (
	"bytes"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy-protocol"
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
	record := string(cres.MainAccount.ToBytes())

	// final hash
	record += string(cres.MetadataBase.Hash()[:])
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (cres *CMBInitResponse) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	// Check if cmb init request existed
	meta, ok := txr.GetMetadata().(*CMBInitResponse)
	if !ok {
		return false, errors.Errorf("error parsing cmb init response metadata")
	}
	_, _, txHash, state, err := bcr.GetCMB(meta.MainAccount.ToBytes())
	if err != nil {
		return false, err
	}

	// Check state of cmb
	if state != CMBRequested {
		return false, errors.Errorf("cmb state is not CMBRequested, cannot create response")
	}

	// Check if only board members created this tx
	if !txCreatedByDCBBoardMember(txr, bcr) {
		return false, errors.Errorf("Tx must be created by DCB Governor")
	}

	// Check if this member hasn't responded to this request
	memberResponded, err := bcr.GetCMBResponse(meta.MainAccount.ToBytes())
	if err != nil {
		return false, errors.Errorf("error getting list of old cmb init responses")
	}
	for _, member := range memberResponded {
		if bytes.Equal(txr.GetJSPubKey(), member) {
			return false, errors.Errorf("each board member can only response once to each cmb init request")
		}
	}

	// Check if response time is not over
	_, blockHash, _, _, err := bcr.GetTransactionByHash(txHash)
	if err != nil {
		return false, err
	}
	reqBlockHeight, _, err := bcr.GetBlockHeightByBlockHash(blockHash)
	curBlockHeight := bcr.GetHeight()
	if curBlockHeight-reqBlockHeight >= CMBInitRefundPeriod {
		return false, errors.Errorf("response time is over for this cmb init request")
	}
	return true, nil
}

func (cres *CMBInitResponse) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return true, false, nil // DCB takes care of fee
}

func (cres *CMBInitResponse) ValidateMetadataByItself() bool {
	return true
}
