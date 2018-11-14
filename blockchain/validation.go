package blockchain

/*
Use these function to validate common data in blockchain
*/

import (
	"errors"
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/transaction"
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
	} else if tx.GetType() == common.TxRegisterCandidateType {
		descs = tx.(*transaction.TxRegisterCandidate).Descs
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

func IsAnyBoardAddressInVins(customToken *transaction.TxCustomToken) bool {
	GOVAddressStr := string(common.GOVAddress)
	DCBAddressStr := string(common.DCBAddress)
	for _, vin := range customToken.TxTokenData.Vins {
		apkStr := string(vin.PaymentAddress.Apk[:])
		if apkStr == GOVAddressStr || apkStr == DCBAddressStr {
			return true
		}
	}
	return false
}

func IsAllBoardAddressInVins(
	customToken *transaction.TxCustomToken,
	boardAddrStr string,
) bool {
	for _, vin := range customToken.TxTokenData.Vins {
		apkStr := string(vin.PaymentAddress.Apk[:])
		if apkStr != boardAddrStr {
			return false
		}
	}
	return true
}

// VerifyMultiSigByBoard: verify multisig if the tx is for board's spending
func (bc *BlockChain) VerifyCustomTokenSigns(tx transaction.Transaction) bool {
	customToken, ok := tx.(*transaction.TxCustomToken)
	if !ok {
		return false
	}

	boardType := customToken.BoardType
	if boardType == 0 { // this tx is not for board's spending so no need to verify multisig
		if IsAnyBoardAddressInVins(customToken) {
			return false
		}
		return true
	}

	if boardType == common.DCB {
		// verify addresses in vins
		if !IsAllBoardAddressInVins(customToken, string(common.DCBAddress)) {
			return false
		}

		// verify signs
		pubKeysByBoard := bc.BestState[0].BestBlock.DCBBoardPubKeys
		fmt.Println("pubKeysByBoard: ", pubKeysByBoard)
		// TODO: do validation here
		return true

	} else if boardType == common.GOV {
		// verify addresses in vins
		if !IsAllBoardAddressInVins(customToken, string(common.GOVAddress)) {
			return false
		}

		pubKeysByBoard := bc.BestState[0].BestBlock.GOVBoardPubKeys
		fmt.Println("pubKeysByBoard: ", pubKeysByBoard)
		// TODO: do validation here
		return true

	} else {
		return false
	}
}
