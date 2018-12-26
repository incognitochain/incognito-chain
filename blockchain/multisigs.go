package blockchain

import (
	"errors"

	"github.com/ninjadotorg/constant/metadata"
)

func (blockGen *BlkTmplGenerator) registerMultiSigsAddresses(
	txs []metadata.Transaction,
) error {
	if len(txs) == 0 {
		return nil
	}
	msRegs := map[string]*metadata.MultiSigsRegistration{}
	sortedTxs := Txs(txs).SortTxs(false)
	for _, tx := range sortedTxs {
		meta := tx.GetMetadata()
		if meta == nil {
			continue
		}
		multiSigsReg, ok := meta.(*metadata.MultiSigsRegistration)
		if !ok {
			return errors.New("Could not parse MultiSigsRegistration metadata")
		}
		msRegs[string(multiSigsReg.PaymentAddress.Pk)] = multiSigsReg
	}
	// TODO: store msRegs to db
	return nil
}
