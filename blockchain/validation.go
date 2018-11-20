package blockchain

/*
Use these function to validate common data in blockchain
*/

import (
	"bytes"
	"errors"
	"fmt"
	"math"

	"encoding/hex"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wallet"
	"golang.org/x/crypto/sha3"
)

/*
IsSalaryTx determines whether or not a transaction is a salary.
*/
func (self *BlockChain) IsSalaryTx(tx transaction.Transaction) bool {
	// Check normal tx(not an action tx)
	if tx.GetType() != common.TxSalaryType {
		normalTx, ok := tx.(*transaction.Tx)
		if !ok {
			return false
		}
		// Check nullifiers in every Descs
		descs := normalTx.Descs
		if len(descs) != 1 {
			return false
		} else {
			if descs[0].Reward > 0 {
				return true
			}
		}
		return false
	}
	return false
}

// ValidateDoubleSpend - check double spend for any transaction type
func (self *BlockChain) ValidateDoubleSpend(tx transaction.Transaction, chainID byte) error {
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

func (self *BlockChain) ValidateTxLoanRequest(tx transaction.Transaction, chainID byte) error {
	txLoan, ok := tx.(*transaction.TxLoanRequest)
	if !ok {
		return fmt.Errorf("Fail parsing LoanRequest transaction")
	}

	// Check if loan's params are correct
	currentParams := self.BestState[chainID].BestBlock.Header.LoanParams
	if txLoan.Params != currentParams {
		return fmt.Errorf("LoanRequest transaction has incorrect params")
	}

	// Check if loan id is unique across all chains
	// TODO(@0xbunyip): should we check in db/chain or only in best state?
	for chainID, bestState := range self.BestState {
		for _, id := range bestState.LoanIDs {
			if bytes.Equal(txLoan.LoanID, id) {
				return fmt.Errorf("LoanID already existed on chain %d", chainID)
			}
		}
	}
	return nil
}

func (self *BlockChain) ValidateTxLoanResponse(tx transaction.Transaction, chainID byte) error {
	txResponse, ok := tx.(*transaction.TxLoanResponse)
	if !ok {
		return fmt.Errorf("Fail parsing LoanResponse transaction")
	}

	// Check if a loan request with the same id exists on any chain
	txHashes, err := self.config.DataBase.GetLoanTxs(txResponse.LoanID)
	if err != nil {
		return err
	}
	found := false
	for _, txHash := range txHashes {
		hash := &common.Hash{}
		copy(hash[:], txHash)
		_, _, _, txOld, err := self.GetTransactionByHash(hash)
		if txOld == nil || err != nil {
			return fmt.Errorf("Error finding corresponding loan request")
		}
		switch txOld.GetType() {
		case common.TxLoanResponse:
			{
				return fmt.Errorf("Loan already had response")
			}
		case common.TxLoanRequest:
			{
				_, ok := txOld.(*transaction.TxLoanRequest)
				if !ok {
					return fmt.Errorf("Error parsing loan request tx")
				}
				found = true
			}
		}
	}

	if found == false {
		return fmt.Errorf("Corresponding loan request not found")
	}
	return nil
}

func (self *BlockChain) ValidateTxLoanPayment(tx transaction.Transaction, chainID byte) error {
	txPayment, ok := tx.(*transaction.TxLoanPayment)
	if !ok {
		return fmt.Errorf("Fail parsing LoanPayment transaction")
	}

	// Check if a loan request with the same id exists on any chain
	txHashes, err := self.config.DataBase.GetLoanTxs(txPayment.LoanID)
	if err != nil {
		return err
	}
	found := false
	for _, txHash := range txHashes {
		hash := &common.Hash{}
		copy(hash[:], txHash)
		_, _, _, txOld, err := self.GetTransactionByHash(hash)
		if txOld == nil || err != nil {
			return fmt.Errorf("Error finding corresponding loan request")
		}
		switch txOld.GetType() {
		case common.TxLoanResponse:
			{
				found = true
			}
		}
	}

	if found == false {
		return fmt.Errorf("Corresponding loan response not found")
	}
	return nil
}

func (self *BlockChain) ValidateTxLoanWithdraw(tx transaction.Transaction, chainID byte) error {
	txWithdraw, ok := tx.(*transaction.TxLoanWithdraw)
	if !ok {
		return fmt.Errorf("Fail parsing LoanResponse transaction")
	}

	// Check if a loan response with the same id exists on any chain
	txHashes, err := self.config.DataBase.GetLoanTxs(txWithdraw.LoanID)
	if err != nil {
		return err
	}
	foundResponse := false
	keyCorrect := false
	for _, txHash := range txHashes {
		hash := &common.Hash{}
		copy(hash[:], txHash)
		_, _, _, txOld, err := self.GetTransactionByHash(hash)
		if txOld == nil || err != nil {
			return fmt.Errorf("Error finding corresponding loan request")
		}
		switch txOld.GetType() {
		case common.TxLoanRequest:
			{
				// Check if key is correct
				txRequest, ok := tx.(*transaction.TxLoanRequest)
				if !ok {
					return fmt.Errorf("Error parsing corresponding loan request")
				}
				h := make([]byte, 32)
				sha3.ShakeSum256(h, txWithdraw.Key)
				if bytes.Equal(h, txRequest.KeyDigest) {
					keyCorrect = true
				}
			}
		case common.TxLoanResponse:
			{
				// Check if loan is accepted
				txResponse, ok := tx.(*transaction.TxLoanResponse)
				if !ok {
					return fmt.Errorf("Error parsing corresponding loan response")
				}
				if txResponse.Response != transaction.Accept {
					foundResponse = true
				}
			}

		}
	}

	if !foundResponse {
		return fmt.Errorf("Corresponding loan response not found")
	} else if !keyCorrect {
		return fmt.Errorf("Provided key is incorrect")
	}
	return nil
}

func (self *BlockChain) GetAmountPerAccount(proposal *transaction.PayoutProposal) (uint64, [][]byte, []uint64, error) {
	// TODO(@0xbunyip): cache list so that list of receivers is fixed across blocks
	tokenHolders, err := self.GetListTokenHolders(proposal.TokenID)
	if err != nil {
		return 0, nil, nil, err
	}

	// Get total token supply
	totalTokenSupply := uint64(0)
	for holder, _ := range tokenHolders {
		temp, _ := hex.DecodeString(holder)
		paymentAddress := privacy.PaymentAddress{}
		paymentAddress.FromBytes(temp)
		utxos := self.GetAccountUTXO(paymentAddress.Pk[:])
		for i := 0; i < len(utxos); i += 1 {
			// TODO(@0xbunyip): get amount from utxo hash
			value := uint64(0)
			totalTokenSupply += value
		}
	}

	// Get amount per account
	rewardHolders := [][]byte{}
	amounts := []uint64{}
	for holder, _ := range tokenHolders {
		temp, _ := hex.DecodeString(holder)
		paymentAddress := privacy.PaymentAddress{}
		paymentAddress.FromBytes(temp)
		utxos := self.GetAccountUTXO(paymentAddress.Pk[:]) // Cached data
		amount := uint64(0)
		for i := 0; i < len(utxos); i += 1 {
			reward, err := self.GetUTXOReward(utxos[i]) // Data from latest block
			if err != nil {
				return 0, nil, nil, err
			}
			if reward < proposal.PayoutID {
				// TODO(@0xbunyip): get amount from utxo hash
				value := uint64(0)
				amount += value
			}
		}

		if amount > 0 {
			rewardHolders = append(rewardHolders, paymentAddress.Pk[:])
			amounts = append(amounts, amount)
		}
	}
	return totalTokenSupply, rewardHolders, amounts, nil
}

func (self *BlockChain) ValidateTxDividendPayout(tx transaction.Transaction, chainID byte) error {
	txPayout, ok := tx.(*transaction.TxDividendPayout)
	if !ok {
		return fmt.Errorf("Fail parsing DividendPayout transaction")
	}

	// Check if there's a proposal to pay dividend
	// TODO(@0xbunyip): get current proposal and check if it is dividend payout
	proposal := &transaction.PayoutProposal{}
	_, tokenHolders, amounts, err := self.GetAmountPerAccount(proposal)
	if err != nil {
		return err
	}

	// Check if user is not rewarded and amount is correct
	for _, desc := range txPayout.Descs {
		for _, note := range desc.Note {
			// Check if user is not rewarded
			utxos := self.GetAccountUTXO(note.Apk[:])
			for _, utxo := range utxos {
				reward, err := self.GetUTXOReward(utxo)
				if err != nil {
					return err
				}
				if reward >= proposal.PayoutID {
					return fmt.Errorf("UTXO %s has already received dividend payment", string(utxo))
				}
			}

			// Check amount
			found := 0
			for i, holder := range tokenHolders {
				if bytes.Equal(holder, note.Apk[:]) {
					found += 1
					if amounts[i] != note.Value {
						return fmt.Errorf("Payment amount for user %s incorrect, found %d instead of %d", holder, note.Value, amounts[i])
					}
				}
			}

			if found == 0 {
				return fmt.Errorf("User %s isn't eligible for receiving dividend", note.Apk[:])
			} else if found > 1 {
				return fmt.Errorf("Multiple dividend payments found for user %s", note.Apk[:])
			}
		}
	}

	return nil
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
	boardPubKeys []string,
) bool {
	boardLen := len(boardPubKeys)
	if boardLen == 0 {
		return false
	}

	signs := tx.BoardSigns
	verifiedSignCount := 0
	tx.BoardSigns = nil

	for _, pubKey := range boardPubKeys {
		sign, ok := signs[pubKey]
		if !ok {
			continue
		}
		keyObj, err := wallet.Base58CheckDeserialize(pubKey)
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
	var pubKeys []string
	if boardType == common.DCB {
		address = string(common.DCBAddress)
		pubKeys = bc.BestState[0].BestBlock.Header.DCBGovernor.DCBBoardPubKeys
	} else if boardType == common.GOV {
		address = string(common.GOVAddress)
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
func (bc *BlockChain) VerifyCustomTokenSigns(tx transaction.Transaction) bool {
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

func (self *BlockChain) ValidateTxBuyRequest(tx transaction.Transaction, chainID byte) error {
	return nil
}

//validate voting transaction
func (bc *BlockChain) ValidateTxSubmitDCBProposal(tx transaction.Transaction, chainID byte) error {
	return nil
}

func (bc *BlockChain) ValidateTxAcceptDCBProposal(tx transaction.Transaction, chainID byte) error {
	return nil
}

func (bc *BlockChain) ValidateTxVoteDCBProposal(tx transaction.Transaction, chainID byte) error {
	return nil
}

func (bc *BlockChain) ValidateTxSubmitGOVProposal(tx transaction.Transaction, chainID byte) error {
	return nil
}

func (bc *BlockChain) ValidateTxAcceptGOVProposal(tx transaction.Transaction, chainID byte) error {
	return nil
}

func (bc *BlockChain) ValidateTxVoteGOVProposal(tx transaction.Transaction, chainID byte) error {
	return nil
}

func (self *BlockChain) ValidateDoubleSpendCustomToken(tx *transaction.TxCustomToken) (error) {
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

func (self *BlockChain) ValidateDoubleSpendCustomTokenOnTx(tx *transaction.TxCustomToken, txInBlock transaction.Transaction) (error) {
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
