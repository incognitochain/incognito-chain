package blockchain

import (
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
)

func (bc *BlockChain) GetAmountPerAccount(tokenID *common.Hash) (uint64, []privacy.PaymentAddress, []uint64, error) {
	tokenHoldersMap, err := bc.config.DataBase.GetCustomTokenPaymentAddressesBalanceUnreward(tokenID)
	if err != nil {
		return 0, nil, nil, err
	}

	// Get total token supply
	totalTokenAmount := uint64(0)
	for _, value := range tokenHoldersMap {
		totalTokenAmount += value
	}

	// Get amount per account (only count unrewarded utxo)
	tokenHolders := []privacy.PaymentAddress{}
	amounts := []uint64{}
	for holder, amount := range tokenHoldersMap {
		paymentAddressInBytes, _, _ := base58.Base58Check{}.Decode(holder)
		keySet := cashec.KeySet{}
		keySet.PaymentAddress = privacy.PaymentAddress{}
		keySet.PaymentAddress.SetBytes(paymentAddressInBytes)
		//vouts, err := bc.GetUnspentTxCustomTokenVout(keySet, tokenID)
		//if err != nil {
		//	return 0, nil, nil, err
		//}
		//amount := uint64(0)
		//for _, vout := range vouts {
		//	amount += vout.Value
		//}

		if amount > 0 {
			tokenHolders = append(tokenHolders, keySet.PaymentAddress)
			amounts = append(amounts, amount)
		}
	}
	return totalTokenAmount, tokenHolders, amounts, nil
}

func (blockgen *BlkTmplGenerator) buildInstitutionDividendSubmitTx(forDCB bool) (metadata.Transaction, error) {
	// Get latest dividend proposal id and amount
	id, _ := blockgen.chain.BestState.Beacon.GetLatestDividendProposal(forDCB)
	if id == 0 {
		return nil, nil // No Dividend proposal found
	}

	// Check in shard state if DividendSubmit tx has been created
	_, _, hasValue, err := blockgen.chain.config.DataBase.GetDividendReceiversForID(id, forDCB)
	if err != nil {
		return nil, err
	}
	if hasValue {
		return nil, nil // Already created DividendSubmit tx in previous blocks
	}

	tokenID := &common.DCBTokenID
	if !forDCB {
		tokenID = &common.GOVTokenID
	}
	_, _, _, err = blockgen.chain.GetAmountPerAccount(tokenID)
	if err != nil {
		return nil, err
	}

	return nil, nil
	//infos := []metadata.DividendPayment{}
	//// Build tx to pay dividend to each holder
	//for i, holder := range tokenHolders {
	//	holderAddrInBytes, _, err := base58.Base58Check{}.Decode(holder)
	//	if err != nil {
	//		return nil, 0, err
	//	}
	//	holderAddress := (&privacy.PaymentAddress{}).SetBytes(holderAddrInBytes)
	//	info := metadata.DividendPayment{
	//		TokenHolder: *holderAddress,
	//		Amount:      amounts[i] / totalTokenAmount,
	//	}
	//	payoutAmount += info.Amount
	//	infos = append(infos, info)

	//	if len(infos) > metadata.MaxDivTxsPerBlock {
	//		break // Pay dividend to only some token holders in this block
	//	}
	//}

	//dividendTxs, err = transaction.BuildDividendTxs(infos, proposal, producerPrivateKey, blockgen.chain.GetDatabase())
	//if err != nil {
	//	return nil, 0, err
	//}
}

func (blockgen *BlkTmplGenerator) buildDividendSubmitTx() ([]metadata.Transaction, error) {
	// For DCB
	dcbDividendSubmitTx, err := blockgen.buildInstitutionDividendSubmitTx(true)
	if err != nil {
		return nil, err
	}

	// TODO: For GOV

	return []metadata.Transaction{dcbDividendSubmitTx}, nil
}
