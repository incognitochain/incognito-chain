package blockchain

import (
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/database"
	"github.com/ninjadotorg/cash-prototype/transaction"
)

type UsedTxViewPoint struct {
	listNullifiers [][]byte

	// hash of best block in current
	currentBestHash common.Hash
}

// BestHash returns the hash of the best block in the chain the view currently
// represents.
func (view *UsedTxViewPoint) BestHash() *common.Hash {
	return &view.currentBestHash
}

// SetBestHash sets the hash of the best block in the chain the view currently
// represents.
func (view *UsedTxViewPoint) SetBestHash(hash *common.Hash) {
	view.currentBestHash = *hash
}

/**
// fetchUsedTx get list nullifiers from txs in block and check if they are not in Main chain db
return list new nullifiers
*/
func (view *UsedTxViewPoint) fetchUsedTx(db database.DB, block *Block) (error) {
	transactions := block.Transactions
	// Loop through all of the transaction descs (except for the coinbase tx)
	acceptedNullifiers := make([][]byte, 0)
	for _, tx := range transactions[1:] {
		for _, desc := range tx.(*transaction.Tx).Desc {
			for _, nullifier := range desc.Nullifiers {
				temp, err := db.HasNullifier(nullifier)
				if err != nil {
					return err
				}
				if !temp {
					acceptedNullifiers = append(acceptedNullifiers, nullifier)
				}
			}
		}
	}

	if len(acceptedNullifiers) > 0 {
		for _, nullifier := range acceptedNullifiers {
			view.listNullifiers = append(view.listNullifiers, nullifier)
		}
	}
	return nil
}

func NewUsedTxViewPoint() *UsedTxViewPoint {
	return &UsedTxViewPoint{
		listNullifiers: make([][]byte, 0),
	}
}
