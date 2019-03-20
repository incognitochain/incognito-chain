package metadata

import (
	"bytes"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/wallet"
	"github.com/pkg/errors"
)

type CMBInitResponse struct {
	MainAccount privacy.PaymentAddress

	MetadataBase
}

func NewCMBInitResponse(data map[string]interface{}) *CMBInitResponse {
	keyWalletMainKey, err := wallet.Base58CheckDeserialize(data["MainAccount"].(string))
	if err != nil {
		return nil
	}
	result := CMBInitResponse{
		MainAccount: keyWalletMainKey.KeySet.PaymentAddress,
	}
	result.Type = CMBInitResponseMeta
	return &result
}

func (cres *CMBInitResponse) Hash() *common.Hash {
	record := cres.MainAccount.String()

	// final hash
	record += cres.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (cres *CMBInitResponse) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// Check if cmb init request existed
	_, _, _, txHash, state, _, err := bcr.GetCMB(cres.MainAccount.Bytes())
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
	memberResponded, err := bcr.GetCMBResponse(cres.MainAccount.Bytes())
	if err != nil {
		return false, errors.Errorf("error getting list of old cmb init responses")
	}
	for _, member := range memberResponded {
		if bytes.Equal(txr.GetSigPubKey(), member) {
			return false, errors.Errorf("each board member can only response once to each cmb init request")
		}
	}

	// Check if response time is not over
	_, blockHash, _, _, err := bcr.GetTransactionByHash(txHash)
	if err != nil {
		return false, err
	}
	reqBlockHeight, _, _ := bcr.GetBlockHeightByBlockHash(blockHash)
	curBlockHeight, err := bcr.GetTxChainHeight(txr)
	if err != nil || curBlockHeight-reqBlockHeight >= CMBInitRefundPeriod {
		return false, errors.Errorf("response time is over for this cmb init request")
	}
	return true, nil
}

func (cres *CMBInitResponse) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	// TODO(@0xbunyip)
	return true, false, nil // DCB takes care of fee
}

func (cres *CMBInitResponse) ValidateMetadataByItself() bool {
	// TODO(@0xbunyip)
	return true
}

func (cres *CMBInitResponse) CalculateSize() uint64 {
	return calculateSize(cres)
}
