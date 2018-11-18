package blockchain

/*
Use these function to validate common data in blockchain
*/

import (
	"errors"
	"fmt"
	"math"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wallet"
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

func isAnyBoardAddressInVins(customToken *transaction.TxCustomToken) bool {
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

func isAllBoardAddressesInVins(
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

func verifyByBoard(
	bc *BlockChain,
	boardType uint8,
	customToken *transaction.TxCustomToken,
) bool {
	var address string
	var pubKeys []string
	if boardType == common.DCB {
		address = string(common.DCBAddress)
		pubKeys = bc.BestState[0].BestBlock.Header.DCDParams.DCBBoardPubKeys
	} else if boardType == common.GOV {
		address = string(common.GOVAddress)
		pubKeys = bc.BestState[0].BestBlock.Header.GOVConstitution.GOVBoardPubKeys
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

	return verifyByBoard(bc, boardType, customToken)
}
