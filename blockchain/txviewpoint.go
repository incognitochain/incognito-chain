package blockchain

import (
	"errors"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/transaction"
	"math/big"
	"github.com/ninjadotorg/constant/privacy-protocol/zero-knowledge"
	"github.com/ninjadotorg/constant/privacy-protocol"
)

type TxViewPoint struct {
	chainID         byte
	listNullifiers  [][]byte
	listCommitments [][]byte
	listSnD         []big.Int
	customTokenTxs  map[int32]*transaction.TxCustomToken

	// hash of best block in current
	currentBestBlockHash common.Hash
}

/*
ListNullifiers returns list nullifers which is contained in TxViewPoint
*/
// #1: joinSplitDescType is "Coin" Or "Bond" or other token
func (view *TxViewPoint) ListNullifiers() [][]byte {
	return view.listNullifiers
}

/*
ListNullifiers returns list commitments which is contained in TxViewPoint
*/
// #1: joinSplitDescType is "Coin" Or "Bond"
func (view *TxViewPoint) ListCommitments() [][]byte {
	return view.listCommitments
}

func (view *TxViewPoint) ListSnDerivators() []big.Int {
	return view.listSnD
}

func (view *TxViewPoint) ListNullifiersEclipsePoint() []*privacy.EllipticPoint {
	result := []*privacy.EllipticPoint{}
	for _, commitment := range view.listNullifiers {
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
func (view *TxViewPoint) processFetchTxViewPoint(block *Block, db database.DatabaseInterface, proof *zkp.PaymentProof) ([][]byte, [][]byte, []big.Int, error) {
	acceptedNullifiers := make([][]byte, 0)
	acceptedCommitments := make([][]byte, 0)
	acceptedSnD := make([]big.Int, 0)
	if proof == nil {
		return acceptedNullifiers, acceptedCommitments, acceptedSnD, nil
	}
	for _, item := range proof.InputCoins {
		serialNum := item.CoinDetails.SerialNumber.Compress()
		temp, err := db.HasSerialNumber(serialNum, block.Header.ChainID)
		if err != nil {
			return nil, nil, nil, err
		}
		if !temp {
			acceptedNullifiers = append(acceptedNullifiers, serialNum)
		}
	}
	for _, item := range proof.OutputCoins {
		commitment := item.CoinDetails.CoinCommitment.Compress()
		temp, err := db.HasCommitment(commitment, block.Header.ChainID)
		if err != nil {
			return nil, nil, nil, err
		}
		if !temp {
			acceptedCommitments = append(acceptedCommitments, item.CoinDetails.CoinCommitment.Compress())
		}

		snD := item.CoinDetails.SNDerivator
		// TODO
		//temp, err := db.HasSND(snD, block.Header.ChainID)
		if !temp {
			acceptedSnD = append(acceptedSnD, *snD)
		}
	}
	return acceptedNullifiers, acceptedCommitments, acceptedSnD, nil
}

/*
fetchTxViewPointFromBlock get list nullifiers and commitments from txs in block and check if they are not in Main chain db
return a tx view point which contains list new nullifiers and new commitments from block
*/
func (view *TxViewPoint) fetchTxViewPointFromBlock(db database.DatabaseInterface, block *Block) error {
	transactions := block.Transactions
	// Loop through all of the transaction descs (except for the salary tx)
	acceptedNullifiers := make([][]byte, 0)
	acceptedCommitments := make([][]byte, 0)
	acceptedSnD := make([]big.Int, 0)
	for indexTx, tx := range transactions {
		switch tx.GetType() {
		case common.TxNormalType:
			{
				normalTx := tx.(*transaction.Tx)
				temp1, temp2, temp3, err := view.processFetchTxViewPoint(block, db, normalTx.Proof)
				acceptedNullifiers = append(acceptedNullifiers, temp1...)
				acceptedCommitments = append(acceptedCommitments, temp2...)
				acceptedSnD = append(acceptedSnD, temp3...)
				if err != nil {
					return NewBlockChainError(UnExpectedError, err)
				}
			}
		case common.TxSalaryType:
			{
				normalTx := tx.(*transaction.Tx)
				temp1, temp2, temp3, err := view.processFetchTxViewPoint(block, db, normalTx.Proof)
				acceptedNullifiers = append(acceptedNullifiers, temp1...)
				acceptedCommitments = append(acceptedCommitments, temp2...)
				acceptedSnD = append(acceptedSnD, temp3...)
				if err != nil {
					return NewBlockChainError(UnExpectedError, err)
				}
			}
		case common.TxCustomTokenType:
			{
				tx := tx.(*transaction.TxCustomToken)
				temp1, temp2, temp3, err := view.processFetchTxViewPoint(block, db, tx.Proof)
				acceptedNullifiers = append(acceptedNullifiers, temp1...)
				acceptedCommitments = append(acceptedCommitments, temp2...)
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

	if len(acceptedNullifiers) > 0 {
		for _, item := range acceptedNullifiers {
			view.listNullifiers = append(view.listNullifiers, item)
		}
	}
	if len(acceptedCommitments) > 0 {
		for _, item := range acceptedCommitments {
			view.listCommitments = append(view.listCommitments, item)
		}
	}
	if len(acceptedSnD) > 0 {
		for _, item := range acceptedSnD {
			view.listSnD = append(view.listSnD, item)
		}
	}
	return nil
}

/*
Create a TxNormal view point, which contains data about nullifiers and commitments
*/
func NewTxViewPoint(chainId byte) *TxViewPoint {
	return &TxViewPoint{
		chainID:         chainId,
		listNullifiers:  make([][]byte, 0),
		listCommitments: make([][]byte, 0),
		listSnD:         make([]big.Int, 0),
		customTokenTxs:  make(map[int32]*transaction.TxCustomToken, 0),
	}
}
