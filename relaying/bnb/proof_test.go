package relaying

import (
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/types"
	"testing"
)

func TestValidTxProof(t *testing.T) {
	cases := []struct {
		txs types.Txs
	}{
		{types.Txs{{1, 4, 34, 87, 163, 1}}},
		{types.Txs{{5, 56, 165, 2}, {4, 77}}},
		{types.Txs{types.Tx("foo"), types.Tx("bar"), types.Tx("baz")}},
		//{makeTxs(20, 5)},
		//{makeTxs(7, 81)},
		//{makeTxs(61, 15)},
	}
	blockHeight := int64(5)

	for h, tc := range cases {
		txs := tc.txs
		root := txs.Hash()
		// make sure valid proof for every tx
		for i := range txs {
			tx := []byte(txs[i])
			//proof := txs.Proof(i)
			proof, _ := buildProof(blockHeight, i)
			assert.Equal(t, i, proof.Proof.Index, "%d: %d", h, i)
			assert.Equal(t, len(txs), proof.Proof.Total, "%d: %d", h, i)
			assert.EqualValues(t, root, proof.RootHash, "%d: %d", h, i)
			assert.EqualValues(t, tx, proof.Data, "%d: %d", h, i)
			assert.EqualValues(t, txs[i].Hash(), proof.Leaf(), "%d: %d", h, i)
			isValid , err := verifyProof(proof, blockHeight, root)
			assert.Equal(t, true, isValid)
			assert.Nil(t, err, "%d: %d", h, i)
			//assert.NotNil(t, proof.Validate([]byte("foobar")), "%d: %d", h, i)

			// read-write must also work
			//var p2 types.TxProof
			//bin, err := types.cdc.MarshalBinaryLengthPrefixed(proof)
			//assert.Nil(t, err)
			//err = cdc.UnmarshalBinaryLengthPrefixed(bin, &p2)
			//if assert.Nil(t, err, "%d: %d: %+v", h, i, err) {
			//	assert.Nil(t, p2.Validate(root), "%d: %d", h, i)
			//}
		}
	}
}


func TestGetTxByHash(t *testing.T){
	getTxByHash("24B93E8B6C5817B159870E5C617597EBD0BDAE100430DB8242BFBA5DA37D70CE")
}
