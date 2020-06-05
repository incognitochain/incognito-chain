package transaction

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/stretchr/testify/assert"
)

func TestEstimateTxSize(t *testing.T) {
	key, err := wallet.Base58CheckDeserialize("112t8rnXCqbbNYBquntyd6EvDT4WiDDQw84ZSRDKmazkqrzi6w8rWyCVt7QEZgAiYAV4vhJiX7V9MCfuj4hGLoDN7wdU1LoWGEFpLs59X7K3")
	assert.Equal(t, nil, err)
	err = key.KeySet.InitFromPrivateKey(&key.KeySet.PrivateKey)
	assert.Equal(t, nil, err)
	paymentAddress := key.KeySet.PaymentAddress
	tx := &Tx{}
	err = tx.InitTxSalary(10, &paymentAddress, &key.KeySet.PrivateKey, db, nil)
	if err != nil {
		t.Error(err)
	}

	payments := []*privacy.PaymentInfo{&privacy.PaymentInfo{
		PaymentAddress: paymentAddress,
		Amount:         5,
	}}

	outputCoins := tx.Proof.GetOutputCoins()
	size := EstimateTxSize(NewEstimateTxSizeParam(len(outputCoins), len(payments), true, nil, nil, 1))
	fmt.Println(size)
	assert.Greater(t, size, uint64(0))

	//customTokenParams := CustomTokenParamTx{
	//	Receiver: []TxTokenVout{{PaymentAddress: paymentAddress, Value: 5}},
	//	vins:     []TxTokenVin{{PaymentAddress: paymentAddress, VoutIndex: 1}},
	//}
	//size1 := EstimateTxSize(NewEstimateTxSizeParam(tx.Proof.GetOutputCoins(), payments, true, nil, &customTokenParams, nil, 1))
	//assert.Greater(t, size1, uint64(0))

	privacyCustomTokenParams := CustomTokenPrivacyParamTx{
		Receiver: []*privacy.PaymentInfo{{
			PaymentAddress: paymentAddress, Amount: 5,
		}},
	}
	size2 := EstimateTxSize(NewEstimateTxSizeParam(len(outputCoins), len(payments), true, nil, &privacyCustomTokenParams, 1))
	fmt.Println(size2)
	assert.Greater(t, size2, uint64(0))
}

func TestRandomCommitmentsProcess(t *testing.T) {
	key, _ := wallet.Base58CheckDeserialize("112t8rnXCqbbNYBquntyd6EvDT4WiDDQw84ZSRDKmazkqrzi6w8rWyCVt7QEZgAiYAV4vhJiX7V9MCfuj4hGLoDN7wdU1LoWGEFpLs59X7K3")
	_ = key.KeySet.InitFromPrivateKey(&key.KeySet.PrivateKey)
	paymentAddress := key.KeySet.PaymentAddress
	tx1 := &Tx{}
	err := tx1.InitTxSalary(10, &paymentAddress, &key.KeySet.PrivateKey, db, nil)
	if err != nil {
		t.Error(err)
	}
	statedb.StoreCommitments(db, common.Hash{}, paymentAddress.Pk, [][]byte{tx1.Proof.GetOutputCoins()[0].CoinDetails.GetCoinCommitment().ToBytesS()}, 0)

	//in1 := ConvertOutputCoinToInputCoin(tx1.Proof.GetOutputCoins())

	cmmIndexs, myIndexs, cmm := RandomCommitmentsProcess(NewRandomCommitmentsProcessParam(in1, 0, db, 0, &common.Hash{}))
	assert.Equal(t, 8, len(cmmIndexs))
	assert.Equal(t, 1, len(myIndexs))
	assert.Equal(t, 8, len(cmm))

	tx2 := &Tx{}
	err = tx2.InitTxSalary(5, &paymentAddress, &key.KeySet.PrivateKey, db, nil)
	if err != nil {
		t.Error(err)
	}
	statedb.StoreCommitments(db, common.Hash{}, paymentAddress.Pk, [][]byte{tx2.Proof.GetOutputCoins()[0].CoinDetails.GetCoinCommitment().ToBytesS()}, 0)
	tx3 := &Tx{}
	err = tx3.InitTxSalary(5, &paymentAddress, &key.KeySet.PrivateKey, db, nil)
	statedb.StoreCommitments(db, common.Hash{}, paymentAddress.Pk, [][]byte{tx3.Proof.GetOutputCoins()[0].CoinDetails.GetCoinCommitment().ToBytesS()}, 0)
	//in2 := ConvertOutputCoinToInputCoin(tx2.Proof.GetOutputCoins())
	in := append(in1, in2...)

	cmmIndexs, myIndexs, cmm = RandomCommitmentsProcess(NewRandomCommitmentsProcessParam(in, 0, db, 0, &common.Hash{}))
	assert.Equal(t, 16, len(cmmIndexs))
	assert.Equal(t, 16, len(cmm))
	assert.Equal(t, 2, len(myIndexs))

	//statedb.CleanCommitments(db)
	//cmmIndexs1, myCommIndex1, cmm1 := RandomCommitmentsProcess(NewRandomCommitmentsProcessParam(in, 0, db, 0, &common.Hash{}))
	//assert.Equal(t, 0, len(cmmIndexs1))
	//assert.Equal(t, 0, len(myCommIndex1))
	//assert.Equal(t, 0, len(cmm1))
}

var db incdb.Database
var _ = func() (_ struct{}) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_")
	if err != nil {
		log.Fatalf("failed to create temp dir: %+v", err)
	}
	log.Println(dbPath)
	db, err = incdb.Open("leveldb", dbPath)
	if err != nil {
		log.Fatalf("could not open db path: %s, %+v", dbPath, err)
	}
	incdb.Logger.Init(common.NewBackend(nil).Logger("db", true))
	Logger.Init(common.NewBackend(nil).Logger("tx", true))
	//privacy.Logger.Init(common.NewBackend(nil).Logger("privacy", true))
	return
}()

func TestBuildCoinbaseTxByCoinID(t *testing.T) {
	key, err := wallet.Base58CheckDeserialize("112t8rnXCqbbNYBquntyd6EvDT4WiDDQw84ZSRDKmazkqrzi6w8rWyCVt7QEZgAiYAV4vhJiX7V9MCfuj4hGLoDN7wdU1LoWGEFpLs59X7K3")
	assert.Equal(t, nil, err)
	err = key.KeySet.InitFromPrivateKey(&key.KeySet.PrivateKey)
	assert.Equal(t, nil, err)
	paymentAddress := key.KeySet.PaymentAddress

	txBuild := NewBuildCoinBaseTxByCoinIDParams(
		&paymentAddress,
		10,
		&key.KeySet.PrivateKey,
		db,
		nil,
		common.Hash{},
		NormalCoinType,
		"PRV",
		0,
	)
	tx, err := BuildCoinBaseTxByCoinID(txBuild)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, tx)
	//assert.Equal(t, uint64(10), tx.(*Tx).Proof.GetOutputCoins()[0].CoinDetails.GetValue())
	assert.Equal(t, common.PRVCoinID.String(), tx.GetTokenID().String())

	txCustomTokenPrivacy, err := BuildCoinBaseTxByCoinID(NewBuildCoinBaseTxByCoinIDParams(&paymentAddress, 10, &key.KeySet.PrivateKey, db, nil, common.Hash{2}, CustomTokenPrivacyType, "Custom Token", 0))
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, tx)
	//assert.Equal(t, uint64(10), txCustomTokenPrivacy.(*TxCustomTokenPrivacy).TxPrivacyTokenData.TxNormal.Proof.GetOutputCoins()[0].CoinDetails.GetValue())
	assert.Equal(t, common.Hash{2}.String(), txCustomTokenPrivacy.GetTokenID().String())
}
