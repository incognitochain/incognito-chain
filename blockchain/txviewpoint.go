package blockchain

import (
	"errors"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"math/big"
	"github.com/ninjadotorg/constant/privacy-protocol/zero-knowledge"
)

type TxViewPoint struct {
	chainID         byte
	listNullifiers  [][]byte
	listCommitments [][]byte
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
ListNullifiers returns list nullifers which is contained in TxViewPoint
*/
// #1: joinSplitDescType is "Coin" Or "Bond"
func (view *TxViewPoint) ListCommitments() [][]byte {
	return view.listCommitments
}

/*
CurrentBestBlockHash returns the hash of the best block in the chain the view currently
represents.
*/
func (view *TxViewPoint) CurrentBestBlockHash() *common.Hash {
	return &view.currentBestBlockHash
}

// fetch from desc of tx to get nullifiers and commitments
func (view *TxViewPoint) processFetchTxViewPoint(block *Block, db database.DatabaseInterface, proof *zkp.PaymentProof) ([]byte, []byte, big.Int, error) {
	acceptedNullifiers := []byte{}
	acceptedCommitments := []byte{}
	acceptedSnD := big.Int{}
	temp, err := db.HasNullifier(proof.InputCoins.SerialNumber.Compress(), block.Header.ChainID)
	if err != nil {
		return nil, nil, 0, err
	}
	if !temp {
		acceptedNullifiers = outcoin.CoinDetails.SerialNumber.Compress()
	}
	temp, err = db.HasCommitment(outcoin.CoinDetails.CoinCommitment.Compress(), block.Header.ChainID)
	if err != nil {
		return nil, nil, 0, err
	}
	if !temp {
		acceptedCommitments = outcoin.CoinDetails.CoinCommitment.Compress()
	}
	return acceptedNullifiers, acceptedCommitments, outcoin.CoinDetails., nil
}

/*
fetchTxViewPoint get list nullifiers and commitments from txs in block and check if they are not in Main chain db
return a tx view point which contains list new nullifiers and new commitments from block
*/
func (view *TxViewPoint) fetchTxViewPoint(db database.DatabaseInterface, block *Block) error {
	transactions := block.Transactions
	// Loop through all of the transaction descs (except for the salary tx)
	acceptedNullifiers := make([][]byte, 0)
	acceptedCommitments := make([][]byte, 0)
	for indexTx, tx := range transactions {
		switch tx.GetType() {
		case common.TxNormalType:
			{
				normalTx := tx.(*transaction.TxNormal)
				for _, desc := range normalTx.Descs {
					temp1, temp2, err := view.processFetchTxViewPoint(block, db, desc)
					acceptedNullifiers = append(acceptedNullifiers, temp1...)
					acceptedCommitments = append(acceptedCommitments, temp2...)
					if err != nil {
						return NewBlockChainError(UnExpectedError, err)
					}
				}
			}
		case common.TxSalaryType:
			{
				normalTx := tx.(*transaction.TxNormal)
				for _, desc := range normalTx.Descs {
					temp1, temp2, err := view.processFetchTxViewPoint(block, db, desc)
					acceptedNullifiers = append(acceptedNullifiers, temp1...)
					acceptedCommitments = append(acceptedCommitments, temp2...)
					if err != nil {
						return NewBlockChainError(UnExpectedError, err)
					}
				}
			}
		case common.TxCustomTokenType:
			{
				tx := tx.(*transaction.TxCustomToken)
				for _, outcoin := range tx.Proof.OutputCoins {
					temp1, temp2, err := view.processFetchTxViewPoint(block, db, outcoin)
					acceptedNullifiers = append(acceptedNullifiers, temp1...)
					acceptedCommitments = append(acceptedCommitments, temp2...)
					if err != nil {
						return NewBlockChainError(UnExpectedError, err)
					}
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
		for _, item := range acceptedCommitments {
			view.listCommitments = append(view.listCommitments, item)
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
		customTokenTxs:  make(map[int32]*transaction.TxCustomToken, 0),
	}
}
