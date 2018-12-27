package metadata

import (
	"bytes"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	privacy "github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/pkg/errors"
)

type CMBDepositContract struct {
	MaturityAt    int32
	TotalInterest uint64
	DepositValue  uint64
	NoticePeriod  int32
	Receiver      privacy.PaymentAddress // address of user who wants to deposit
	CMBAddress    privacy.PaymentAddress // address of CMB, must be the same as the one creating this tx

	ValidUntil int32
	MetadataBase
}

func NewCMBDepositContract(data map[string]interface{}) *CMBDepositContract {
	key, err := wallet.Base58CheckDeserialize(data["Receiver"].(string))
	if err != nil {
		return nil
	}
	keyCMB, err := wallet.Base58CheckDeserialize(data["CMBAddress"].(string))
	if err != nil {
		return nil
	}
	maturity := int32(data["MaturityAt"].(float64))
	value := uint64(data["DepositValue"].(float64))
	interest := uint64(data["TotalInterest"].(float64))
	notice := int32(data["NoticePeriod"].(float64))
	validUntil := int32(data["ValidUntil"].(float64))
	result := CMBDepositContract{
		MaturityAt:    maturity,
		TotalInterest: interest,
		DepositValue:  value,
		NoticePeriod:  notice,
		Receiver:      key.KeySet.PaymentAddress,
		CMBAddress:    keyCMB.KeySet.PaymentAddress,
		ValidUntil:    validUntil,
	}

	result.Type = CMBDepositContractMeta
	return &result
}

func (dc *CMBDepositContract) Hash() *common.Hash {
	record := string(dc.MaturityAt)
	record += string(dc.TotalInterest)
	record += string(dc.DepositValue)
	record += string(dc.NoticePeriod)
	record += string(dc.Receiver.ToBytes())
	record += string(dc.CMBAddress.ToBytes())

	// final hash
	record += string(dc.MetadataBase.Hash()[:])
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (dc *CMBDepositContract) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	if bcr.GetHeight()+1 >= dc.ValidUntil {
		return false, errors.Errorf("ValidUntil must be larger than block height")
	}

	// CMBAddress must be valid
	if !bytes.Equal(txr.GetJSPubKey(), dc.CMBAddress.Pk[:]) {
		return false, errors.Errorf("CMBAddress must be the one creating this tx")
	}
	_, _, _, _, _, _, err := bcr.GetCMB(dc.CMBAddress.ToBytes())
	if err != nil {
		return false, err
	}
	return true, nil
}

func (dc *CMBDepositContract) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if dc.ValidUntil >= dc.MaturityAt {
		return false, false, errors.Errorf("Deposit maturity must be greater than ValidUntil")
	}
	return true, true, nil // continue to check for fee
}

func (dc *CMBDepositContract) ValidateMetadataByItself() bool {
	return true
}
