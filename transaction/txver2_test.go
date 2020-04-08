package transaction

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/mlsag"
	"github.com/stretchr/testify/assert"
)

func TestTxSignatureVer2(t *testing.T) {
	var err error
	for i := 0; i < 100; i += 1 {
		n := 5
		m := 8

		keyInputs := []*operation.Scalar{}
		for i := 0; i < m; i += 1 {
			privateKey := operation.RandomScalar()
			keyInputs = append(keyInputs, privateKey)
		}
		pi := common.RandInt() % n
		ring := mlsag.NewRandomRing(keyInputs, n, pi)

		maxLen := new(big.Int)
		maxLen.SetString("1000000000000000000", 10)
		indexes := make([][]*big.Int, n)
		for i := 0; i < n; i += 1 {
			row := make([]*big.Int, m)
			for j := 0; j < m; j += 1 {
				row[j], err = common.RandBigIntMaxRange(maxLen)
				assert.Equal(t, nil, err, "Should not have any bug when Randomizing Int Max Range")
			}
			indexes[i] = row
		}

		txSig := new(TxSignatureVer2)
		txSig.SetRing(ring)
		txSig.SetIndexes(indexes)

		b, err := txSig.ToBytes()
		assert.Equal(t, nil, err, "Should not have any bug when txSig.ToBytes")

		txSig2 := new(TxSignatureVer2)
		err = txSig2.FromBytes(b)
		assert.Equal(t, nil, err, "Should not have any bug when txSig.FromBytes")

		b2, err := txSig2.ToBytes()
		assert.Equal(t, nil, err, "Should not have any bug when txSig2.ToBytes")

		assert.Equal(t, true, bytes.Equal(b, b2))

		b1, _ := txSig.GetRing().ToBytes()
		b2, err = txSig2.GetRing().ToBytes()
		assert.Equal(t, nil, err)
		assert.Equal(t, true, bytes.Equal(b1, b2))

		n1 := len(txSig.indexes)
		m1 := len(txSig.indexes[0])
		n2 := len(txSig2.indexes)
		m2 := len(txSig2.indexes[0])

		assert.Equal(t, n1, n2, "Two indexes length should be equal")
		assert.Equal(t, m1, m2, "Two indexes length should be equal")
		for i := 0; i < n; i += 1 {
			for j := 0; j < m; j += 1 {
				b1 := txSig.indexes[i][j].Bytes()
				b2 := txSig2.indexes[i][j].Bytes()
				assert.Equal(t, true, bytes.Equal(b1, b2), "Indexes[i][j] should be equal for every i j")
			}
		}
	}
}
