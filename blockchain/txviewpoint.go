package blockchain

import (
	"errors"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/transaction"
	"math/big"
	"github.com/ninjadotorg/constant/privacy-protocol/zero-knowledge"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/common/base58"
)

type TxViewPoint struct {
	chainID           byte
	listSerialNumbers [][]byte // array serialNumbers
	listCommitments   [][]byte
	mapCommitments    map[string][][]byte //map[base58check.encode{pubkey}]([]([]byte-commitment))
	mapOutputCoins    map[string][]privacy.OutputCoin

	listSnD        []big.Int
	customTokenTxs map[int32]*transaction.TxCustomToken

	// hash of best block in current
	currentBestBlockHash common.Hash
}

/*
ListSerialNumbers returns list nullifers which is contained in TxViewPoint
*/
// #1: joinSplitDescType is "Coin" Or "Bond" or other token
func (view *TxViewPoint) ListSerialNumbers() [][]byte {
	return view.listSerialNumbers
}

/*
ListCommitments returns list commitments which is contained in TxViewPoint
*/
// #1: joinSplitDescType is "Coin" Or "Bond"
func (view *TxViewPoint) ListCommitments() [][]byte {
	return view.listCommitments
}

func (view *TxViewPoint) ListSnDerivators() []big.Int {
	return view.listSnD
}

func (view *TxViewPoint) ListSerialNumnbersEclipsePoint() []*privacy.EllipticPoint {
	result := []*privacy.EllipticPoint{}
	for _, commitment := range view.listSerialNumbers {
		point := &privacy.EllipticPoint{}
		point.Decompress(commitment)
		result = append(result, point)
	}
	return result
}

/*
CurrentBestBlockHash returns the hash of the best block in the chain the view currently
represents.
*/
func (view *TxViewPoint) CurrentBestBlockHash() *common.Hash {
	return &view.currentBestBlockHash
}

// fetch from desc of tx to get nullifiers and commitments
func (view *TxViewPoint) processFetchTxViewPoint(chainID byte, db database.DatabaseInterface, proof *zkp.PaymentProof) ([][]byte, map[string][][]byte, map[string][]privacy.OutputCoin, []big.Int, error) {
	acceptedNullifiers := make([][]byte, 0)
	acceptedCommitments := make(map[string][][]byte)
	acceptedOutputcoins := make(map[string][]privacy.OutputCoin)
	acceptedSnD := make([]big.Int, 0)
	if proof == nil {
		return acceptedNullifiers, acceptedCommitments, acceptedOutputcoins, acceptedSnD, nil
	}
	for _, item := range proof.InputCoins {
		serialNum := item.CoinDetails.SerialNumber.Compress()
		ok, err := db.HasSerialNumber(serialNum, chainID)
		if err != nil {
			return acceptedNullifiers, acceptedCommitments, acceptedOutputcoins, acceptedSnD, err
		}
		if !ok {
			acceptedNullifiers = append(acceptedNullifiers, serialNum)
		}
	}
	for _, item := range proof.OutputCoins {
		commitment := item.CoinDetails.CoinCommitment.Compress()
		pubkey := item.CoinDetails.PublicKey.Compress()
		ok, err := db.HasCommitment(commitment, chainID)
		if err != nil {
			return acceptedNullifiers, acceptedCommitments, acceptedOutputcoins, acceptedSnD, err
		}
		if !ok {
			pubkeyStr := base58.Base58Check{}.Encode(pubkey, byte(0x00))
			if acceptedCommitments[pubkeyStr] == nil {
				acceptedCommitments[pubkeyStr] = make([][]byte, 0)
			}
			acceptedCommitments[pubkeyStr] = append(acceptedCommitments[pubkeyStr], item.CoinDetails.CoinCommitment.Compress())
			if acceptedOutputcoins[pubkeyStr] == nil {
				acceptedOutputcoins[pubkeyStr] = make([]privacy.OutputCoin, 0)
			}
			acceptedOutputcoins[pubkeyStr] = append(acceptedOutputcoins[pubkeyStr], *item)
		}

		snD := item.CoinDetails.SNDerivator
		ok, err = db.HasSNDerivator(*snD, chainID)
		if !ok && err == nil {
			acceptedSnD = append(acceptedSnD, *snD)
		}
	}
	return acceptedNullifiers, acceptedCommitments, acceptedOutputcoins, acceptedSnD, nil
}

/*
fetchTxViewPointFromBlock get list nullifiers and commitments from txs in block and check if they are not in Main chain db
return a tx view point which contains list new nullifiers and new commitments from block
*/
func (view *TxViewPoint) fetchTxViewPointFromBlock(db database.DatabaseInterface, block *Block) error {
	transactions := block.Transactions
	// Loop through all of the transaction descs (except for the salary tx)
	acceptedSerialNumbers := make([][]byte, 0)
	acceptedCommitments := make(map[string][][]byte)
	acceptedOutputcoins := make(map[string][]privacy.OutputCoin)
	acceptedSnD := make([]big.Int, 0)
	for indexTx, tx := range transactions {
		switch tx.GetType() {
		case common.TxNormalType:
			{
				normalTx := tx.(*transaction.Tx)
				temp1, temp2, temp22, temp3, err := view.processFetchTxViewPoint(block.Header.ChainID, db, normalTx.Proof)
				acceptedSerialNumbers = append(acceptedSerialNumbers, temp1...)
				for pubkey, data := range temp2 {
					if acceptedCommitments[pubkey] == nil {
						acceptedCommitments[pubkey] = make([][]byte, 0)
					}
					acceptedCommitments[pubkey] = append(acceptedCommitments[pubkey], data...)
				}
				for pubkey, data := range temp22 {
					if acceptedOutputcoins[pubkey] == nil {
						acceptedOutputcoins[pubkey] = make([]privacy.OutputCoin, 0)
					}
					acceptedOutputcoins[pubkey] = append(acceptedOutputcoins[pubkey], data...)
				}
				acceptedSnD = append(acceptedSnD, temp3...)
				if err != nil {
					return NewBlockChainError(UnExpectedError, err)
				}
			}
		case common.TxSalaryType:
			{
				normalTx := tx.(*transaction.Tx)
				temp1, temp2, temp22, temp3, err := view.processFetchTxViewPoint(block.Header.ChainID, db, normalTx.Proof)
				acceptedSerialNumbers = append(acceptedSerialNumbers, temp1...)
				for pubkey, data := range temp2 {
					if acceptedCommitments[pubkey] == nil {
						acceptedCommitments[pubkey] = make([][]byte, 0)
					}
					acceptedCommitments[pubkey] = append(acceptedCommitments[pubkey], data...)
				}
				for pubkey, data := range temp22 {
					if acceptedOutputcoins[pubkey] == nil {
						acceptedOutputcoins[pubkey] = make([]privacy.OutputCoin, 0)
					}
					acceptedOutputcoins[pubkey] = append(acceptedOutputcoins[pubkey], data...)
				}
				acceptedSnD = append(acceptedSnD, temp3...)
				if err != nil {
					return NewBlockChainError(UnExpectedError, err)
				}
			}
		case common.TxCustomTokenType:
			{
				tx := tx.(*transaction.TxCustomToken)
				temp1, temp2, temp22, temp3, err := view.processFetchTxViewPoint(block.Header.ChainID, db, tx.Proof)
				acceptedSerialNumbers = append(acceptedSerialNumbers, temp1...)
				for pubkey, data := range temp2 {
					if acceptedCommitments[pubkey] == nil {
						acceptedCommitments[pubkey] = make([][]byte, 0)
					}
					acceptedCommitments[pubkey] = append(acceptedCommitments[pubkey], data...)
				}
				for pubkey, data := range temp22 {
					if acceptedOutputcoins[pubkey] == nil {
						acceptedOutputcoins[pubkey] = make([]privacy.OutputCoin, 0)
					}
					acceptedOutputcoins[pubkey] = append(acceptedOutputcoins[pubkey], data...)
				}
				acceptedSnD = append(acceptedSnD, temp3...)
				if err != nil {
					return NewBlockChainError(UnExpectedError, err)
				}
				view.customTokenTxs[int32(indexTx)] = tx
			}
		default:
			{
				return NewBlockChainError(UnExpectedError, errors.New("TxNormal type is invalid"))
			}
		}
	}

	if len(acceptedSerialNumbers) > 0 {
		view.listSerialNumbers = acceptedSerialNumbers
	}
	if len(acceptedCommitments) > 0 {
		view.mapCommitments = acceptedCommitments
	}
	if len(acceptedOutputcoins) > 0 {
		view.mapOutputCoins = acceptedOutputcoins
	}
	if len(acceptedSnD) > 0 {
		view.listSnD = acceptedSnD
	}
	return nil
}

/*
Create a TxNormal view point, which contains data about nullifiers and commitments
*/
func NewTxViewPoint(chainId byte) *TxViewPoint {
	return &TxViewPoint{
		chainID:           chainId,
		listSerialNumbers: make([][]byte, 0),
		listCommitments:   make([][]byte, 0),
		mapCommitments:    make(map[string][][]byte, 0),
		mapOutputCoins:    make(map[string][]privacy.OutputCoin, 0),
		listSnD:           make([]big.Int, 0),
		customTokenTxs:    make(map[int32]*transaction.TxCustomToken, 0),
	}
}
