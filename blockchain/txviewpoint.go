package blockchain

import (
	"github.com/ninjadotorg/cash/common"
	"github.com/ninjadotorg/cash/database"
	"github.com/ninjadotorg/cash/transaction"
)

type TxViewPoint struct {
	chainId         byte
	listNullifiers  map[string]([][]byte)
	listCommitments map[string]([][]byte)

	// hash of best block in current
	currentBestBlockHash common.Hash
}

/*
ListNullifiers returns list nullifers which is contained in TxViewPoint
*/
// #1: joinSplitDescType is "Coin" Or "Bond"
func (view *TxViewPoint) ListNullifiers(joinSplitDescType string) [][]byte {
	return view.listNullifiers[joinSplitDescType]
}

/*
ListNullifiers returns list nullifers which is contained in TxViewPoint
*/
// #1: joinSplitDescType is "Coin" Or "Bond"
func (view *TxViewPoint) ListCommitments(joinSplitDescType string) [][]byte {
	return view.listCommitments[joinSplitDescType]
}

/*
CurrentBestBlockHash returns the hash of the best block in the chain the view currently
represents.
*/
func (view *TxViewPoint) CurrentBestBlockHash() *common.Hash {
	return &view.currentBestBlockHash
}

/*
SetBestHash sets the hash of the best block in the chain the view currently
represents.
*/
// func (view *TxViewPoint) SetBestHash(hash *common.Hash) {
// 	view.currentBestBlockHash = *hash
// }

/*
fetchTxViewPoint get list nullifiers and commitments from txs in block and check if they are not in Main chain db
return a tx view point which contains list new nullifiers and new commitments from block
*/
func (view *TxViewPoint) fetchTxViewPoint(db database.DatabaseInterface, block *Block) error {
	transactions := block.Transactions
	// Loop through all of the transaction descs (except for the salary tx)
	acceptedNullifiers := make(map[string][][]byte)
	acceptedCommitments := make(map[string][][]byte)
	for _, tx := range transactions {
		if tx.GetType() == common.TxNormalType || tx.GetType() == common.TxSalaryType {
			for _, desc := range tx.(*transaction.Tx).Descs {
				for _, item := range desc.Nullifiers {
					temp, err := db.HasNullifier(item, desc.Type, block.Header.ChainID)
					if err != nil {
						return err
					}
					if !temp {
						acceptedNullifiers[desc.Type] = append(acceptedNullifiers[desc.Type], item)
					}
				}
				for _, item := range desc.Commitments {
					temp, err := db.HasCommitment(item, desc.Type, block.Header.ChainID)
					if err != nil {
						return err
					}
					if !temp {
						acceptedCommitments[desc.Type] = append(acceptedCommitments[desc.Type], item)
					}
				}
			}
		} else if tx.GetType() == common.TxVotingType {
			for _, desc := range tx.(*transaction.TxVoting).Descs {
				for _, item := range desc.Nullifiers {
					temp, err := db.HasNullifier(item, desc.Type, block.Header.ChainID)
					if err != nil {
						return err
					}
					if !temp {
						acceptedNullifiers[desc.Type] = append(acceptedNullifiers[desc.Type], item)
					}
				}
				for _, item := range desc.Commitments {
					temp, err := db.HasCommitment(item, desc.Type, block.Header.ChainID)
					if err != nil {
						return err
					}
					if !temp {
						acceptedCommitments[desc.Type] = append(acceptedCommitments[desc.Type], item)
					}
				}
			}
		}

	}

	if len(acceptedNullifiers) > 0 {
		for key, item := range acceptedNullifiers {
			view.listNullifiers[key] = append(view.listNullifiers[key], item...)
		}
		for key, item := range acceptedCommitments {
			view.listCommitments[key] = append(view.listCommitments[key], item...)
		}
	}
	return nil
}

/*
Create a Tx view point, which contains data about nullifiers and commitments
*/
func NewTxViewPoint(chainId byte) *TxViewPoint {
	return &TxViewPoint{
		chainId:         chainId,
		listNullifiers:  make(map[string]([][]byte)),
		listCommitments: make(map[string]([][]byte)),
	}
}
