package mempool

import (
	"testing"

	"github.com/ninjadotorg/constant/transaction"
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

func (m *TxMockedObject) GetTransaction() transaction.TxNormal {

	//hash, error := common.Hash{}.NewHash([]byte("12345678901234567890123456789012"))
	//if error != nil {
	//
	//}
	return transaction.TxNormal{
		Version:  0,
		TxIn:     nil,
		TxOut:    nil,
		LockTime: 0,
	}

}

func TestTxPool_CanAcceptTransaction(t *testing.T) {
	mempool := New(&Config{
		Policy: Policy{},
	})
	tx := transaction.TxNormal{
		Version:  0,
		TxIn:     nil,
		TxOut:    nil,
		LockTime: 0,
	}
	txHash, txDesc, txError := mempool.CanAcceptTransaction(&tx)

	assert.Nil(t, txError)

	assert.NotNil(t, txHash, "hash should not nil")
	assert.NotEmpty(t, txDesc, "have txDesc")
}
