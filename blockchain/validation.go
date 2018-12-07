package blockchain

/*
Use these function to validate common data in blockchain
*/

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wallet"
)

/*
IsSalaryTx determines whether or not a transaction is a salary.
*/
// func (self *BlockChain) IsSalaryTx(tx metadata.Transaction) bool {
// 	// Check normal tx(not an action tx)
// 	if tx.GetType() != common.TxSalaryType {
// 		normalTx, ok := tx.(*transaction.Tx)
// 		if !ok {
// 			return false
// 		}
// 		// Check nullifiers in every Descs
// 		descs := normalTx.Descs
// 		if len(descs) != 1 {
// 			return false
// 		} else {
// 			if descs[0].Reward > 0 {
// 				return true
// 			}
// 		}
// 		return false
// 	}
// 	return false
// }

// ValidateDoubleSpend - check double spend for any transaction type
func (self *BlockChain) ValidateDoubleSpend(tx metadata.Transaction, chainID byte) error {
	txHash := tx.Hash()
	txViewPoint, err := self.FetchTxViewPoint(chainID)
	if err != nil {
		str := fmt.Sprintf("Can not check double spend for tx")
		err := NewBlockChainError(CanNotCheckDoubleSpendError, errors.New(str))
		return err
	}
	nullifierDb := txViewPoint.ListNullifiers()
	var descs []*transaction.JoinSplitDesc
	if tx.GetType() == common.TxNormalType {
		descs = tx.(*transaction.Tx).Descs
	}
	for _, desc := range descs {
		for _, nullifer := range desc.Nullifiers {
			existed, err := common.SliceBytesExists(nullifierDb, nullifer)
			if err != nil {
				str := fmt.Sprintf("Can not check double spend for tx")
				err := NewBlockChainError(CanNotCheckDoubleSpendError, errors.New(str))
				return err
			}
			if existed {
				str := fmt.Sprintf("Nullifiers of transaction %+v already existed", txHash.String())
				err := NewBlockChainError(CanNotCheckDoubleSpendError, errors.New(str))
				return err
			}
		}
	}
	return nil
}

func (self *BlockChain) GetAmountPerAccount(proposal *metadata.DividendProposal) (uint64, []string, []uint64, error) {
	// TODO(@0xsirrush): cache list so that list of receivers is fixed across blocks
	tokenHolders, err := self.config.DataBase.GetCustomTokenListPaymentAddressesBalance(proposal.TokenID)
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
		temp, _ := hex.DecodeString(holder)
		paymentAddress := (&privacy.PaymentAddress{}).FromBytes(temp)
		utxos, err := self.config.DataBase.GetCustomTokenPaymentAddressUTXO(proposal.TokenID, *paymentAddress)
		if err != nil {
			return 0, nil, nil, err
		}
		amount := uint64(0)
		for _, vout := range utxos {
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
		isValid, err := keyObj.KeySet.Verify(common.ToBytes(tx), common.ToBytes(sign))
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

func (self *BlockChain) ValidateTxBuySellDCBRequest(tx metadata.Transaction, chainID byte) error {
	// Check if crowdsale existed
	requestTx, ok := tx.(*transaction.TxBuySellRequest)
	if !ok {
		return fmt.Errorf("Error parsing TxBuySellDCBRequest")
	}
	saleData, err := self.config.DataBase.LoadCrowdsaleData(requestTx.SaleID)
	if err != nil {
		return fmt.Errorf("SaleID not found")
	}

	// Check if sale is still valid
	if self.BestState[chainID].Height >= saleData.EndBlock {
		return fmt.Errorf("Sale ended")
	}

	dbcAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	if bytes.Equal(saleData.BuyingAsset[:8], BondTokenID[:8]) {
		for _, vout := range requestTx.TxTokenData.Vouts {
			if !bytes.Equal(vout.BuySellResponse.BondID, saleData.BuyingAsset[8:]) {
				return fmt.Errorf("Received asset id %s instead of %s", append(BondTokenID[:8], vout.BuySellResponse.BondID...), saleData.BuyingAsset)
			}

			// Check if receiving address is DCB's
			if !bytes.Equal(vout.PaymentAddress.Pk[:], dbcAccount.KeySet.PaymentAddress.Pk) {
				return fmt.Errorf("Sending payment to %x instead of %x", vout.PaymentAddress.Pk[:], common.DCBAddress)
			}
		}
	} else if bytes.Equal(saleData.BuyingAsset, ConstantID[:]) {
		for _, desc := range requestTx.Tx.Descs {
			for _, note := range desc.Note {
				if !bytes.Equal(note.Apk[:], dbcAccount.KeySet.PaymentAddress.Pk) {
					return fmt.Errorf("Sending payment to %x instead of %x", note.Apk[:], common.DCBAddress)
				}
			}
		}
	}
	return nil
}

func (self *BlockChain) ValidateTxBuySellDCBResponse(tx metadata.Transaction, chainID byte) error {
	// Check if crowdsale existed
	responseTx, ok := tx.(*transaction.TxBuySellDCBResponse)
	if !ok {
		return fmt.Errorf("Error parsing TxBuySellDCBResponse")
	}
	saleData, err := self.config.DataBase.LoadCrowdsaleData(responseTx.SaleID)
	if err != nil {
		return fmt.Errorf("SaleID not found")
	}

	// Check if sale is still valid
	if self.BestState[chainID].Height >= saleData.EndBlock {
		return fmt.Errorf("Sale ended")
	}

	if !bytes.Equal(saleData.SellingAsset[:8], BondTokenID[:8]) {
		return fmt.Errorf("Sending asset id %s instead of %s", BondTokenID, saleData.SellingAsset)
	}

	// TODO(@0xbunyip): validate amount of asset sent
	return nil
}

//validate voting transaction
func (bc *BlockChain) ValidateTxSubmitDCBProposal(tx metadata.Transaction, chainID byte) error {
	return nil
}

func (bc *BlockChain) ValidateTxAcceptDCBProposal(tx metadata.Transaction, chainID byte) error {
	return nil
}

func (bc *BlockChain) ValidateTxVoteDCBProposal(tx metadata.Transaction, chainID byte) error {
	return nil
}

func (bc *BlockChain) ValidateTxSubmitGOVProposal(tx metadata.Transaction, chainID byte) error {
	return nil
}

func (bc *BlockChain) ValidateTxAcceptGOVProposal(tx metadata.Transaction, chainID byte) error {
	return nil
}

func (bc *BlockChain) ValidateTxVoteGOVProposal(tx metadata.Transaction, chainID byte) error {
	return nil
}

func (self *BlockChain) ValidateDoubleSpendCustomToken(tx *transaction.TxCustomToken) error {
	listTxs, err := self.GetCustomTokenTxs(&tx.TxTokenData.PropertyID)
	if err != nil {
		return err
	}

	if len(listTxs) == 0 {
		if tx.TxTokenData.Type != transaction.CustomTokenInit {
			return errors.New("Not exist tx for this ")
		}
	}

	if len(listTxs) > 0 {
		for _, txInBlocks := range listTxs {
			err := self.ValidateDoubleSpendCustomTokenOnTx(tx, txInBlocks)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (self *BlockChain) ValidateDoubleSpendCustomTokenOnTx(tx *transaction.TxCustomToken, txInBlock metadata.Transaction) error {
	temp := txInBlock.(*transaction.TxCustomToken)
	for _, vin := range temp.TxTokenData.Vins {
		for _, item := range tx.TxTokenData.Vins {
			if vin.TxCustomTokenID.String() == item.TxCustomTokenID.String() {
				if vin.VoutIndex == item.VoutIndex {
					return errors.New("Double spend")
				}
			}
		}
	}
	return nil
}

func (self *BlockChain) ValidateBuyFromGOVRequestTx(
	tx metadata.Transaction,
	chainID byte,
) error {
	buySellReqTx, ok := tx.(*transaction.TxBuySellRequest)
	if !ok {
		return fmt.Errorf("Fail parsing TxBuySellRequest transaction")
	}

	// check double spending on fee + amount tx
	err := self.ValidateDoubleSpend(&buySellReqTx.Tx, chainID)
	if err != nil {
		return err
	}

	// TODO: support and validate for either bonds or govs buy requests

	sellingBondsParams := self.BestState[chainID].BestBlock.Header.GOVConstitution.GOVParams.SellingBonds
	if sellingBondsParams == nil {
		return errors.New("SellingBonds params are not existed.")
	}

	// check if buy price againsts SellingBonds params' BondPrice is correct or not
	if buySellReqTx.BuyPrice < sellingBondsParams.BondPrice {
		return errors.New("Requested buy price is under SellingBonds params' buy price.")
	}
	return nil
}

func (self *BlockChain) ValidateBuyBackRequestTx(
	tx metadata.Transaction,
	chainID byte,
) error {
	buyBackReqTx, ok := tx.(*transaction.TxBuyBackRequest)
	if !ok {
		return fmt.Errorf("Fail parsing TxBuyBackRequest transaction")
	}

	// check double spending on fee + amount tx
	err := self.ValidateDoubleSpend(buyBackReqTx.Tx, chainID)
	if err != nil {
		return err
	}

	return nil
}
