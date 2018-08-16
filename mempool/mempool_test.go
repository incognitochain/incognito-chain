package mempool

import (
	"testing"

	"github.com/internet-cash/prototype/common"
	"github.com/internet-cash/prototype/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TxMockedObject struct {
	mock.Mock
}

func TestNew(t *testing.T) {
	mempool := New(&Config{
		Policy: Policy{},
	})

	assert.Equal(t, mempool.Count(), 0, "mempool was started")
}

func (m *TxMockedObject) GetTransaction() transaction.Tx {

	hash, error := common.Hash{}.NewHash([]byte("12345678901234567890123456789012"))
	if error != nil {

	}
	return transaction.Tx{
		Version:  0,
		TxIn:     nil,
		TxOut:    nil,
		LockTime: 0,
		TxHash:   hash,
	}

}

func TestTxPool_CanAcceptTransaction(t *testing.T) {
	mempool := New(&Config{
		Policy: Policy{},
	})
	hash, error := common.Hash{}.NewHash([]byte("12345678901234567890123456789012"))
	if error != nil {

	}
	tx := transaction.Tx{
		Version:  0,
		TxIn:     nil,
		TxOut:    nil,
		LockTime: 0,
		TxHash:   hash,
	}
	txHash, txDesc, txError := mempool.CanAcceptTransaction(&tx)

	assert.Nil(t, txError)

	assert.Nil(t, txHash, "should not return hash")
	assert.NotEmpty(t, txDesc, "have txDesc")
}
