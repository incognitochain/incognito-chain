package blockchain

/*
Use these function to validate common data in blockchain
*/

import (
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
)

func (self *BlockChain) GetAmountPerAccount(proposal *metadata.DividendProposal) (uint64, []string, []uint64, error) {
	tokenHolders, err := self.config.DataBase.GetCustomTokenPaymentAddressesBalanceUnreward(proposal.TokenID)
	if err != nil {
		return 0, nil, nil, err
	}

	// Get total token supply
	totalTokenSupply := uint64(0)
	for _, value := range tokenHolders {
		totalTokenSupply += value
	}

	// Get amount per account (only count unrewarded utxo)
	rewardHolders := []string{}
	amounts := []uint64{}
	for holder := range tokenHolders {
		paymentAddressInBytes, _, _ := base58.Base58Check{}.Decode(holder)
		keySet := cashec.KeySet{}
		keySet.PaymentAddress = privacy.PaymentAddress{}
		keySet.PaymentAddress.SetBytes(paymentAddressInBytes)
		vouts, err := self.GetUnspentTxCustomTokenVout(keySet, proposal.TokenID)
		if err != nil {
			return 0, nil, nil, err
		}
		amount := uint64(0)
		for _, vout := range vouts {
			amount += vout.Value
		}

		if amount > 0 {
			rewardHolders = append(rewardHolders, holder)
			amounts = append(amounts, amount)
		}
	}
	return totalTokenSupply, rewardHolders, amounts, nil
}
