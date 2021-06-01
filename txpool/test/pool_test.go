package txpool

import (
	"errors"
	"fmt"
	"math/big"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/metadata/mocks"
	"github.com/incognitochain/incognito-chain/privacy"
	zkp "github.com/incognitochain/incognito-chain/privacy/zeroknowledge"
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

const (
	REJECT_INVALID_WITHCHAINSTATE = iota
	REJECT_DOUBLESTAKE
	REJECT_DOUBLESPEND
	REJECT_LOADCOMMITMENT
)

func init() {
	bLog := common.NewBackend(logWriter{})
	txPoolLogger := bLog.Logger("Txpool log ", false)
	pool.Logger.Init(txPoolLogger)
	mapCommitment = map[string]interface{}{}
}

func setupPoolData(
	testType int,
) (
	*pool.TxsPool,
	[]metadata.Transaction,
	[]metadata.Transaction,
) {
	p := setupPool(VALID)
	txVer := &pMocks.TxVerifier{}
	listTxs := []metadata.Transaction{}
	listAccepted := []metadata.Transaction{}
	switch testType {
	case REJECT_DOUBLESPEND:
		// tx.On("LoadCommitment", mock.AnythingOfType("*mock.Transaction"), mock.AnythingOfType("metadata.ShardViewRetriever")).Return(true)
		txVer, listTxs, listAccepted = setupDoubleSpend(txVer, p, listTxs, listAccepted)
	case REJECT_DOUBLESTAKE:

	}
	p.UpdateTxVerifier(txVer)
	sort.Slice(listAccepted, func(i, j int) bool {
		return listAccepted[i].Hash().String() < listAccepted[j].Hash().String()
	})
	return p, listTxs, listAccepted
}

func setupDoubleSpend(
	txVer *pMocks.TxVerifier,
	p *pool.TxsPool,
	listTxs []metadata.Transaction,
	listAccepted []metadata.Transaction,
) (
	*pMocks.TxVerifier,
	[]metadata.Transaction,
	[]metadata.Transaction,
) {
	txVer.On(
		"ValidateWithChainState",
		mock.AnythingOfType("*mocks.Transaction"),
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(true, nil)
	txVer.On(
		"LoadCommitment",
		mock.AnythingOfType("*mocks.Transaction"),
		mock.Anything,
	).Return(true, nil)
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			tx := mocks.Transaction{}
			oCoins := []*privacy.OutputCoin{}
			for t := 0; t < 5; t++ {
				oCoin := &privacy.OutputCoin{
					CoinDetails: &privacy.Coin{},
				}
				oCoin.CoinDetails.SetSNDerivator(privacy.BigIntToScalar(big.NewInt(int64(10*i + j + t))))
				oCoins = append(oCoins, oCoin)
			}
			iCoins := []*privacy.InputCoin{}
			for t := 0; t < 5; t++ {
				iCoin := &privacy.InputCoin{
					CoinDetails: &privacy.Coin{},
				}
				iCoin.CoinDetails.SetSerialNumber(privacy.HashToPoint(big.NewInt(int64(10*i + j + t)).Bytes()))
				iCoins = append(iCoins, iCoin)
			}
			prf := &zkp.PaymentProof{}
			prf.SetInputCoins(iCoins)
			prf.SetOutputCoins(oCoins)
			tx.On("GetProof").Return(prf)
			h := common.HashH([]byte(fmt.Sprintf("Tx%v", i*3+j)))
			tx.On("Hash").Return(&h)
			tx.On("GetTxFee").Return(uint64(10 - j))
			tx.On("GetType").Return(common.TxNormalType)
			tx.On("GetMetadata").Return(nil)
			tx.On("GetMetadataType").Return(metadata.InvalidMeta)
			p.Data.TxByHash[tx.Hash().String()] = &tx
			p.Data.TxInfos[tx.Hash().String()] = pool.TxInfo{
				Fee:   uint64(10 - j),
				Size:  100 - uint64(j),
				VTime: 3 * time.Millisecond,
			}
			listTxs = append(listTxs, &tx)
			if j == 0 {
				listAccepted = append(listAccepted, &tx)
			}
		}
	}
	return txVer, listTxs, listAccepted
}

func setupDoubleStake(
	txVer *pMocks.TxVerifier,
	p *pool.TxsPool,
	listTxs []metadata.Transaction,
	listAccepted []metadata.Transaction,
) (
	*pMocks.TxVerifier,
	[]metadata.Transaction,
	[]metadata.Transaction,
) {
	txVer.On(
		"ValidateWithChainState",
		mock.AnythingOfType("*mocks.Transaction"),
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(true, nil)
	txVer.On(
		"LoadCommitment",
		mock.AnythingOfType("*mocks.Transaction"),
		mock.Anything,
	).Return(true, nil)
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			tx := mocks.Transaction{}
			oCoins := []*privacy.OutputCoin{}
			for t := 0; t < 5; t++ {
				oCoin := &privacy.OutputCoin{
					CoinDetails: &privacy.Coin{},
				}
				oCoin.CoinDetails.SetSNDerivator(privacy.BigIntToScalar(big.NewInt(int64(10*i + j + t))))
				oCoins = append(oCoins, oCoin)
			}
			iCoins := []*privacy.InputCoin{}
			for t := 0; t < 5; t++ {
				iCoin := &privacy.InputCoin{
					CoinDetails: &privacy.Coin{},
				}
				iCoin.CoinDetails.SetSerialNumber(privacy.HashToPoint(big.NewInt(int64(10*i + j + t)).Bytes()))
				iCoins = append(iCoins, iCoin)
			}
			prf := &zkp.PaymentProof{}
			prf.SetInputCoins(iCoins)
			prf.SetOutputCoins(oCoins)
			tx.On("GetProof").Return(prf)
			h := common.HashH([]byte(fmt.Sprintf("Tx%v", i*3+j)))
			tx.On("Hash").Return(&h)
			tx.On("GetTxFee").Return(uint64(10 - j))
			tx.On("GetType").Return(common.TxNormalType)
			tx.On("GetMetadata").Return(nil)
			tx.On("GetMetadataType").Return(metadata.InvalidMeta)
			p.Data.TxByHash[tx.Hash().String()] = &tx
			p.Data.TxInfos[tx.Hash().String()] = pool.TxInfo{
				Fee:   uint64(10 - j),
				Size:  100 - uint64(j),
				VTime: 3 * time.Millisecond,
			}
			listTxs = append(listTxs, &tx)
			if j == 0 {
				listAccepted = append(listAccepted, &tx)
			}
		}
	}
	return txVer, listTxs, listAccepted
}

func setupPool(testType int) *pool.TxsPool {
	res := pool.NewTxsPool(
		&pMocks.TxVerifier{},
		make(chan metadata.Transaction),
	)
	switch testType {
	case INVALID_CANNOTLOADCOMMIT:
		txVer := &pMocks.TxVerifier{}
		txVer.On("LoadCommitment", mock.AnythingOfType("*mocks.Transaction"), mock.Anything).Return(false, errors.New("Can not load commitment"))
		res.UpdateTxVerifier(txVer)
	case INVALID_WITHOUTCHAINSTATE:
		txVer := &pMocks.TxVerifier{}
		txVer.On("LoadCommitment", mock.AnythingOfType("*mocks.Transaction"), mock.Anything).Return(true, nil)
		txVer.On("ValidateWithoutChainstate", mock.AnythingOfType("*mocks.Transaction")).Return(false, errors.New("ValidateWithoutChainstate failed"))
		res.UpdateTxVerifier(txVer)
	case INVALID_TIMEOUT:
		txVer := &pMocks.TxVerifier{}
		txVer.On("LoadCommitment", mock.AnythingOfType("*mocks.Transaction"), mock.Anything).Run(func(args mock.Arguments) { time.Sleep(2100 * time.Millisecond) }).Return(true, nil)
		res.UpdateTxVerifier(txVer)
	case VALID:
		txVer := &pMocks.TxVerifier{}
		txVer.On("LoadCommitment", mock.AnythingOfType("*mocks.Transaction"), mock.Anything).Return(true, nil)
		txVer.On("ValidateWithoutChainstate", mock.AnythingOfType("*mocks.Transaction")).Return(true, nil)
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
	case INVALID_CANNOTLOADCOMMIT, INVALID_WITHOUTCHAINSTATE, INVALID_TIMEOUT, VALID:
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
		{
			name:     "Tx Can not validate without chainstate",
			testType: INVALID_WITHOUTCHAINSTATE,
			txPool:   nil,
			tx:       make([]metadata.Transaction, 0),
			want:     false,
			want1:    2 * time.Second,
			wantErr:  true,
			setup:    NormalPoolAndInvalidTx,
		},
		{
			name:     "Validate transaction timeout",
			testType: INVALID_TIMEOUT,
			txPool:   nil,
			tx:       make([]metadata.Transaction, 0),
			want:     false,
			want1:    2 * time.Second,
			wantErr:  true,
			setup:    NormalPoolAndInvalidTx,
		},
		{
			name:     "Valid transaction",
			testType: VALID,
			txPool:   nil,
			tx:       make([]metadata.Transaction, 0),
			want:     true,
			want1:    2 * time.Second,
			wantErr:  false,
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

func TestTxsPool_GetTxsTranferForNewBlock(t *testing.T) {
	type args struct {
		cView          metadata.ChainRetriever
		sView          metadata.ShardViewRetriever
		bcView         metadata.BeaconViewRetriever
		maxSize        uint64
		maxTime        time.Duration
		getTxsDuration time.Duration
	}
	tests := []struct {
		name     string
		testType int
		args     args
		setup    func(testType int) (*pool.TxsPool, []metadata.Transaction, []metadata.Transaction)
		want     []metadata.Transaction
	}{
		{
			name:     "Double spend",
			testType: REJECT_DOUBLESPEND,
			args: args{
				cView:          nil,
				sView:          &blockchain.ShardBestState{},
				bcView:         nil,
				maxSize:        4096,
				maxTime:        24 * time.Second,
				getTxsDuration: 5 * time.Second,
			},
			setup: setupPoolData,
			want:  []metadata.Transaction{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tp *pool.TxsPool
			tp, _, tt.want = tt.setup(tt.testType)
			go tp.Start()
			got := tp.GetTxsTranferForNewBlock(
				tt.args.cView,
				tt.args.sView,
				tt.args.bcView,
				tt.args.maxSize,
				tt.args.maxTime,
				tt.args.getTxsDuration,
			)
			sort.Slice(got, func(i, j int) bool {
				return got[i].Hash().String() < got[j].Hash().String()
			})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TxsPool.GetTxsTranferForNewBlock() = %v, want %v", got, tt.want)
			}
			go tp.Stop()
		})
	}
}

// func Test_setupPoolData(t *testing.T) {

// 	got, got1, got2 := setupPoolData(REJECT_DOUBLESPEND)
// 	fmt.Println(got)
// 	for _, v := range got1 {
// 		fmt.Printf("%v ", v.Hash().String())
// 	}
// 	fmt.Println()
// 	for _, v := range got2 {
// 		fmt.Printf("%v ", v.Hash().String())
// 	}
// }

func TestTxsPool_CheckDoubleSpend(t *testing.T) {
	dataHelper := map[[privacy.Ed25519KeySize]byte]struct {
		Index  uint
		Detail pool.TxInfoDetail
	}{}
	tp, txs, wanted := setupPoolData(REJECT_DOUBLESPEND)
	txsNew := []metadata.Transaction{}
	for i := 2; i >= 0; i-- {
		for j := 2; j >= 0; j-- {

			isDoubleSpend, needToReplace, _ := tp.CheckDoubleSpend(dataHelper, txs[j*3+i], &txsNew)
			if isDoubleSpend && !needToReplace {
				continue
			}
			txsNew = insertTxIntoList(dataHelper, pool.TxInfoDetail{
				Fee:   tp.Data.TxInfos[txs[j*3+i].Hash().String()].Fee,
				Size:  tp.Data.TxInfos[txs[j*3+i].Hash().String()].Size,
				VTime: tp.Data.TxInfos[txs[j*3+i].Hash().String()].VTime,
				Tx:    txs[j*3+i],
				Hash:  txs[j*3+i].Hash().String(),
			}, txsNew)
		}
	}
	removeNilTx(&txsNew)
	sort.Slice(txsNew, func(i, j int) bool {
		return txsNew[i].Hash().String() < txsNew[j].Hash().String()
	})
	if !reflect.DeepEqual(txsNew, wanted) {
		t.Errorf("TxsPool.GetTxsTranferForNewBlock() = %v, want %v", txsNew, wanted)
	}
}

func removeNilTx(txs *[]metadata.Transaction) {
	j := 0
	for _, tx := range *txs {
		if tx == nil {
			continue
		}
		(*txs)[j] = tx
		j++
	}
	*txs = (*txs)[:j]
}
func insertTxIntoList(
	dataHelper map[[privacy.Ed25519KeySize]byte]struct {
		Index  uint
		Detail pool.TxInfoDetail
	},
	txDetail pool.TxInfoDetail,
	txs []metadata.Transaction,
) []metadata.Transaction {
	tx := txDetail.Tx
	prf := tx.GetProof()
	if prf != nil {
		insertPrfForCheck(prf, dataHelper, txDetail, len(txs))
	}
	if tx.GetType() == common.TxCustomTokenPrivacyType {
		txNormal := tx.(*transaction.TxCustomTokenPrivacy).TxPrivacyTokenData.TxNormal
		normalPrf := txNormal.GetProof()
		if normalPrf != nil {
			insertPrfForCheck(normalPrf, dataHelper, txDetail, len(txs))
		}
	}
	return append(txs, tx)
}
func insertPrfForCheck(
	prf *zkp.PaymentProof,
	dataHelper map[[privacy.Ed25519KeySize]byte]struct {
		Index  uint
		Detail pool.TxInfoDetail
	},
	txDetail pool.TxInfoDetail,
	idx int,
) {
	iCoins := prf.GetInputCoins()
	oCoins := prf.GetOutputCoins()
	for _, iCoin := range iCoins {
		dataHelper[iCoin.CoinDetails.GetSerialNumber().ToBytes()] = struct {
			Index  uint
			Detail pool.TxInfoDetail
		}{
			Index:  uint(idx),
			Detail: txDetail,
		}
	}
	for _, oCoin := range oCoins {
		dataHelper[oCoin.CoinDetails.GetSNDerivator().ToBytes()] = struct {
			Index  uint
			Detail pool.TxInfoDetail
		}{
			Index:  uint(idx),
			Detail: txDetail,
		}
	}
}
