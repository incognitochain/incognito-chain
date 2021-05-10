package txpool

import (
	"os"
	"testing"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/metadata/mocks"
	"github.com/incognitochain/incognito-chain/transaction"
	pool "github.com/incognitochain/incognito-chain/txpool"
	pMocks "github.com/incognitochain/incognito-chain/txpool/mocks"
	"github.com/stretchr/testify/mock"
)

type logWriter struct{}

func (logWriter) Write(p []byte) (n int, err error) {
	os.Stdout.Write(p)

	return len(p), nil
}

var mapCommitment map[string]interface{}

const (
	INVALID_NOTFORUSER = iota
	INVALID_CANNOTLOADCOMMIT
	INVALID_WITHOUTCHAINSTATE
	INVALID_TIMEOUT
	VALID
)

func init() {
	bLog := common.NewBackend(logWriter{})
	txPoolLogger := bLog.Logger("Txpool log ", false)
	pool.Logger.Init(txPoolLogger)
	mapCommitment = map[string]interface{}{}
}

func setupPool(testType int) *pool.TxsPool {
	res := pool.NewTxsPool(
		&pMocks.TxVerifier{},
		make(chan metadata.Transaction),
	)
	switch testType {
	case INVALID_CANNOTLOADCOMMIT:
		txVer := &pMocks.TxVerifier{}
		txVer.On("LoadCommitment", mock.AnythingOfType("")).Return(false)
		res.UpdateTxVerifier(txVer)
	}
	return res
}

func setupNormalTx() []metadata.Transaction {
	res := []metadata.Transaction{}
	txNormal := &mocks.Transaction{}
	txNormal.On("GetType").Return(common.TxNormalType)
	h1 := common.HashH([]byte{1, 2, 34})
	txNormal.On("Hash").Return(&h1)
	res = append(res, txNormal)
	return res
}

func setupTxNotForUser() []metadata.Transaction {
	res := []metadata.Transaction{}
	txReturnStake := &mocks.Transaction{}
	txReturnStake.On("GetType").Return(common.TxReturnStakingType)
	h1 := common.HashH([]byte{1, 2, 34})
	txReturnStake.On("Hash").Return(&h1)
	res = append(res, txReturnStake)
	txReward := &mocks.Transaction{}
	txReward.On("GetType").Return(common.TxRewardType)
	h2 := common.HashH([]byte{1, 2, 35})
	txReward.On("Hash").Return(&h2)
	res = append(res, txReward)
	txEmbed := &mocks.Transaction{}
	txEmbed.On("GetType").Return(common.TxCustomTokenPrivacyType)
	h3 := common.HashH([]byte{1, 2, 35})
	txEmbed.On("Hash").Return(&h3)
	txCustom := transaction.TxCustomTokenPrivacy{
		Tx: transaction.Tx{
			Type: common.TxCustomTokenPrivacyType,
		},
		TxPrivacyTokenData: transaction.TxPrivacyTokenData{
			Mintable: true,
		},
	}
	res = append(res, &txCustom)
	return res
}

func NormalPoolAndInvalidTx(testType int) (*pool.TxsPool, []metadata.Transaction) {
	switch testType {
	case INVALID_NOTFORUSER:
		return setupPool(testType), setupTxNotForUser()
	case INVALID_CANNOTLOADCOMMIT:
		return setupPool(testType), setupNormalTx()
	}
	return setupPool(testType), setupTxNotForUser()
}

func TestTxsPool_ValidateNewTx(t *testing.T) {
	type args struct {
		tx *mocks.Transaction
	}
	tests := []struct {
		name     string
		testType int
		txPool   *pool.TxsPool
		tx       []metadata.Transaction
		want     bool
		want1    time.Duration
		wantErr  bool
		setup    func(testType int) (txP *pool.TxsPool, tx []metadata.Transaction)
	}{
		{
			name:     "Tx not for user",
			testType: INVALID_NOTFORUSER,
			txPool:   nil,
			tx:       make([]metadata.Transaction, 0),
			want:     false,
			want1:    2 * time.Second,
			wantErr:  true,
			setup:    NormalPoolAndInvalidTx,
		},
		{
			name:     "Tx Can not load commitment",
			testType: INVALID_CANNOTLOADCOMMIT,
			txPool:   nil,
			tx:       make([]metadata.Transaction, 0),
			want:     false,
			want1:    2 * time.Second,
			wantErr:  true,
			setup:    NormalPoolAndInvalidTx,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.txPool, tt.tx = tt.setup(tt.testType)
			for _, tx := range tt.tx {
				got, err, got1 := tt.txPool.ValidateNewTx(tx)
				if (err != nil) != tt.wantErr {
					t.Errorf("pool.TxsPool.ValidateNewTx() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("pool.TxsPool.ValidateNewTx() got = %v, want %v", got, tt.want)
				}
				if got1 > tt.want1 {
					t.Errorf("pool.TxsPool.ValidateNewTx() got1 = %v, want %v", got1, tt.want1)
				}
			}
		})
	}
}
