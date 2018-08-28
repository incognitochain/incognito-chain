package blockchain

import (
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/transaction"
)

type txoFlags uint8

const (
	// tfCoinBase indicates that a txout was contained in a coinbase tx.
	tfCoinBase txoFlags = 1 << iota

	// tfSpent indicates that a txout is spent.
	tfSpent

	// tfModified indicates that a txout has been modified since it was
	// loaded.
	tfModified
)

type UtxoEntry struct {
	amount      int64
	pkScript    []byte // The public key script for the output.
	blockHeight int32  // Height of block containing tx.

	packedFlags txoFlags
}

func (entry *UtxoEntry) IsSpent() bool {
	return entry.packedFlags&tfSpent == tfSpent
}

type UtxoViewpoint struct {
	entries  map[transaction.OutPoint]*UtxoEntry
	bestHash common.Hash
}

func (view *UtxoViewpoint) LookupEntry(outpoint transaction.OutPoint) *UtxoEntry {
	return view.entries[outpoint]
}

// func (b *BlockChain) FetchUtxoView(tx transaction.Transaction) (*UtxoViewpoint, error) {

// 	neededSet := make(map[transaction.OutPoint]struct{})
// 	prevOut := transaction.OutPoint{Hash: *tx.Hash()}
// 	for txOutIdx := range tx.TxOut {
// 		prevOut.Index = uint32(txOutIdx)
// 		neededSet[prevOut] = struct{}{}
// 	}
// 	if !IsCoinBaseTx(tx) {
// 		for _, txIn := range tx.TxIn {
// 			neededSet[txIn.PreviousOutPoint] = struct{}{}
// 		}
// 	}

// 	// Request the utxos from the point of view of the end of the main
// 	// chain.
// 	view := NewUtxoViewpoint()
// 	b.chainLock.RLock()
// 	//@todo will implement late
// 	//err := view.fetchUtxosMain(b.d, neededSet)
// 	b.chainLock.RUnlock()
// 	return view, nil
// }

// func (view *UtxoViewpoint) fetchUtxosMain(db database.DB, outpoints map[transaction.OutPoint]struct{}) error {
//         // Nothing to do if there are no requested outputs.
//         if len(outpoints) == 0 {
//                 return nil
//         }
//
//         [>return db.View(func(dbTx database.Tx) error {
//                 for outpoint := range outpoints {
//                         //entry, err :=  nil, nil ///dbFetchUtxoEntry(dbTx, outpoint)
//                         //if err != nil {
//                         //	return err
//                         //}
//                         var entry UtxoEntry
//                         view.entries[outpoint] = &entry
//                 }
//
//                 return nil
//         })*/
//         return nil
// }

func NewUtxoViewpoint() *UtxoViewpoint {
	return &UtxoViewpoint{
		entries: make(map[transaction.OutPoint]*UtxoEntry),
	}
}
