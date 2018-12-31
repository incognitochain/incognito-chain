package blockchain

/*
Use these function to validate common data in blockchain
*/

import (
	"math"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common/base58"
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
	for holder, _ := range tokenHolders {
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

func isAnyBoardAddressInVins(customToken *transaction.TxCustomToken) bool {
	GOVAddressStr := string(common.GOVAddress)
	DCBAddressStr := string(common.DCBAddress)
	for _, vin := range customToken.TxTokenData.Vins {
		apkStr := string(vin.PaymentAddress.Pk[:])
		if apkStr == GOVAddressStr || apkStr == DCBAddressStr {
			return true
		}
	}
	return false
}

func isAllBoardAddressesInVins(
	customToken *transaction.TxCustomToken,
	boardAddrStr string,
) bool {
	for _, vin := range customToken.TxTokenData.Vins {
		apkStr := string(vin.PaymentAddress.Pk[:])
		if apkStr != boardAddrStr {
			return false
		}
	}
	return true
}

func verifySignatures(
	tx *transaction.TxCustomToken,
	boardPubKeys [][]byte,
) bool {
	boardLen := len(boardPubKeys)
	if boardLen == 0 {
		return false
	}

	signs := tx.BoardSigns
	verifiedSignCount := 0
	tx.BoardSigns = nil

	for _, pubKey := range boardPubKeys {
		sign, ok := signs[string(pubKey)]
		if !ok {
			continue
		}
		keyObj, err := wallet.Base58CheckDeserialize(string(pubKey))
		if err != nil {
			Logger.log.Info(err)
			continue
		}
		isValid, err := keyObj.KeySet.Verify(common.ToBytes(*tx), common.ToBytes(sign))
		if err != nil {
			Logger.log.Info(err)
			continue
		}
		if isValid {
			verifiedSignCount += 1
		}
	}

	if verifiedSignCount >= int(math.Floor(float64(boardLen/2)))+1 {
		return true
	}
	return false
}

func (bc *BlockChain) verifyByBoard(
	boardType uint8,
	customToken *transaction.TxCustomToken,
) bool {
	var address string
	var pubKeys [][]byte
	if boardType == common.DCB {
		address = string(common.DCBAddress)
		pubKeys = bc.BestState[0].BestBlock.Header.DCBGovernor.DCBBoardPubKeys
	} else if boardType == common.GOV {
		govAccount, _ := wallet.Base58CheckDeserialize(common.GOVAddress)
		address = string(govAccount.KeySet.PaymentAddress.Pk)
		pubKeys = bc.BestState[0].BestBlock.Header.GOVGovernor.GOVBoardPubKeys
	} else {
		return false
	}

	if !isAllBoardAddressesInVins(customToken, address) {
		return false
	}
	return verifySignatures(customToken, pubKeys)
}

// VerifyMultiSigByBoard: verify multisig if the tx is for board's spending
func (bc *BlockChain) VerifyCustomTokenSigns(tx metadata.Transaction) bool {
	customToken, ok := tx.(*transaction.TxCustomToken)
	if !ok {
		return false
	}

	boardType := customToken.BoardType
	if boardType == 0 { // this tx is not for board's spending so no need to verify multisig
		if isAnyBoardAddressInVins(customToken) {
			return false
		}
		return true
	}

	return bc.verifyByBoard(boardType, customToken)
}
