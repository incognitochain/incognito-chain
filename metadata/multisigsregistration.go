package metadata

import (
	"errors"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
)

type MultiSigsRegistration struct {
	PaymentAddress   privacy.PaymentAddress // registering address
	SpendableMembers [][]byte               // pub keys of spendable members™
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

func (msReg *MultiSigsRegistration) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	chainID byte,
	db database.DatabaseInterface,
) (bool, error) {
	_, err := db.GetMultiSigsRegistration(msReg.PaymentAddress.Pk)
	if err == nil { // found
		return false, errors.New("The payment address's public key is already existed.")
	}
	if err != lvdberr.ErrNotFound {
		return false, err
	}
	return true, nil
}

func (msReg *MultiSigsRegistration) ValidateSanityData(
	bcr BlockchainRetriever,
	txr Transaction,
) (bool, bool, error) {
	if len(msReg.PaymentAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's payment address")
	}
	if len(msReg.PaymentAddress.Tk) == 0 {
		return false, false, errors.New("Wrong request info's payment address")
	}
	if len(msReg.SpendableMembers) == 0 {
		return false, false, errors.New("Wrong request info's spendable members")
	}
	for _, pk := range msReg.SpendableMembers {
		if len(pk) == 0 {
			return false, false, errors.New("Wrong request info's spendable members")
		}
	}

	return true, true, nil
}

func (msReg *MultiSigsRegistration) ValidateMetadataByItself() bool {
	if msReg.Type != MultiSigsRegistrationMeta {
		return false
	}
	return true
}

func (msReg *MultiSigsRegistration) Hash() *common.Hash {
	record := msReg.PaymentAddress.String()
	for _, pk := range msReg.SpendableMembers {
		record += string(pk)
	}
	record += msReg.MetadataBase.Hash().String()

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}
