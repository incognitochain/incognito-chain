package metadata

import (
	"bytes"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/pkg/errors"
)

type CMBDepositContract struct {
	MaturityAt    uint64
	TotalInterest uint64
	DepositValue  uint64
	NoticePeriod  uint64
	Receiver      privacy.PaymentAddress // address of user who wants to deposit
	CMBAddress    privacy.PaymentAddress // address of CMB, must be the same as the one creating this tx

	ValidUntil uint64
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
	maturity := uint64(data["MaturityAt"].(float64))
	value := uint64(data["DepositValue"].(float64))
	interest := uint64(data["TotalInterest"].(float64))
	notice := uint64(data["NoticePeriod"].(float64))
	validUntil := uint64(data["ValidUntil"].(float64))
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
	record += dc.Receiver.String()
	record += dc.CMBAddress.String()

	// final hash
	record += dc.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (dc *CMBDepositContract) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	lastByte := dc.Receiver.Pk[len(dc.Receiver.Pk)-1]
	chainID, err := common.GetTxSenderChain(lastByte)
	receiverChainHeight := bcr.GetChainHeight(chainID)
	if err != nil || receiverChainHeight+1 >= dc.ValidUntil {
		return common.FalseValue, errors.Errorf("ValidUntil must be bigger than current block height of receiver")
	}

	// CMBAddress must be valid
	if !bytes.Equal(txr.GetSigPubKey(), dc.CMBAddress.Pk[:]) {
		return common.FalseValue, errors.Errorf("CMBAddress must be the one creating this tx")
	}
	_, _, _, _, _, _, err = bcr.GetCMB(dc.CMBAddress.Bytes())
	if err != nil {
		return common.FalseValue, err
	}
	return common.TrueValue, nil
}

func (dc *CMBDepositContract) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if dc.ValidUntil >= dc.MaturityAt {
		return common.FalseValue, common.FalseValue, errors.Errorf("Deposit maturity must be greater than ValidUntil")
	}
	if len(dc.Receiver.Pk) <= 0 {
		return common.FalseValue, common.FalseValue, errors.Errorf("Receiver must be set")
	}
	return common.TrueValue, common.TrueValue, nil // continue to check for fee
}

func (dc *CMBDepositContract) ValidateMetadataByItself() bool {
	return common.TrueValue
}
