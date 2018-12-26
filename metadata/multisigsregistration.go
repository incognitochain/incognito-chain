package metadata

import (
	"errors"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

type MultiSigsRegistration struct {
	PaymentAddress   privacy.PaymentAddress // registing address
	SpendableMembers [][]byte
	MetadataBase
}

func NewMultiSigsRegistration(
	paymentAddress privacy.PaymentAddress,
	spendableMembers [][]byte,
	metaType int,
) *MultiSigsRegistration {
	metaBase := MetadataBase{
		Type: metaType,
	}
	return &MultiSigsRegistration{
		PaymentAddress:   paymentAddress,
		SpendableMembers: spendableMembers,
		MetadataBase:     metaBase,
	}
}

func (msr *MultiSigsRegistration) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	chainID byte,
	db database.DatabaseInterface,
) (bool, error) {
	// TODO: check registing address is existed or not
	return true, nil
}

func (msr *MultiSigsRegistration) ValidateSanityData(
	bcr BlockchainRetriever,
	txr Transaction,
) (bool, bool, error) {
	if len(msr.PaymentAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's payment address")
	}
	if len(msr.PaymentAddress.Tk) == 0 {
		return false, false, errors.New("Wrong request info's payment address")
	}
	if len(msr.SpendableMembers) == 0 {
		return false, false, errors.New("Wrong request info's spendable members")
	}
	for _, pk := range msr.SpendableMembers {
		if len(pk) == 0 {
			return false, false, errors.New("Wrong request info's spendable members")
		}
	}

	return true, true, nil
}

func (msr *MultiSigsRegistration) ValidateMetadataByItself() bool {
	if msr.Type != MultiSigsRegistrationMeta {
		return false
	}
	return true
}

func (msr *MultiSigsRegistration) Hash() *common.Hash {
	record := string(msr.PaymentAddress.Bytes())
	for _, pk := range msr.SpendableMembers {
		record += string(pk)
	}
	record += string(msr.MetadataBase.Hash()[:])

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}
