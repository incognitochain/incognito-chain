package transaction

import (
	"github.com/incognitochain/incognito-chain/common"
	_ "github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConvertOutputCoinToInputCoin(t *testing.T) {
	key, err := wallet.Base58CheckDeserialize("112t8rnXCqbbNYBquntyd6EvDT4WiDDQw84ZSRDKmazkqrzi6w8rWyCVt7QEZgAiYAV4vhJiX7V9MCfuj4hGLoDN7wdU1LoWGEFpLs59X7K3")
	assert.Equal(t, nil, err)
	err = key.KeySet.ImportFromPrivateKey(&key.KeySet.PrivateKey)
	assert.Equal(t, nil, err)
	paymentAddress := key.KeySet.PaymentAddress
	tx, err := BuildCoinbaseTx(&paymentAddress, 10, &key.KeySet.PrivateKey, db, nil)

	in := ConvertOutputCoinToInputCoin(tx.Proof.OutputCoins)
	assert.Equal(t, 1, len(in))
	assert.Equal(t, tx.Proof.OutputCoins[0].CoinDetails.Value, in[0].CoinDetails.Value)
}

func TestEstimateTxSize(t *testing.T) {
	key, err := wallet.Base58CheckDeserialize("112t8rnXCqbbNYBquntyd6EvDT4WiDDQw84ZSRDKmazkqrzi6w8rWyCVt7QEZgAiYAV4vhJiX7V9MCfuj4hGLoDN7wdU1LoWGEFpLs59X7K3")
	assert.Equal(t, nil, err)
	err = key.KeySet.ImportFromPrivateKey(&key.KeySet.PrivateKey)
	assert.Equal(t, nil, err)
	paymentAddress := key.KeySet.PaymentAddress
	tx, err := BuildCoinbaseTx(&paymentAddress, 10, &key.KeySet.PrivateKey, db, nil)

	payments := []*privacy.PaymentInfo{&privacy.PaymentInfo{
		PaymentAddress: paymentAddress,
		Amount:         5,
	}}

	size := EstimateTxSize(tx.Proof.OutputCoins, payments, true, nil, nil, nil, 1)
	assert.Greater(t, size, uint64(0))

	customTokenParams := CustomTokenParamTx{
		Receiver: []TxTokenVout{{PaymentAddress: paymentAddress, Value: 5}},
		vins:     []TxTokenVin{{PaymentAddress: paymentAddress, VoutIndex: 1}},
	}
	size1 := EstimateTxSize(tx.Proof.OutputCoins, payments, true, nil, &customTokenParams, nil, 1)
	assert.Greater(t, size1, uint64(0))

	privacyCustomTokenParams := CustomTokenPrivacyParamTx{
		Receiver: []*privacy.PaymentInfo{{
			PaymentAddress: paymentAddress, Amount: 5,
		}},
	}
	size2 := EstimateTxSize(tx.Proof.OutputCoins, payments, true, nil, nil, &privacyCustomTokenParams, 1)
	assert.Greater(t, size2, uint64(0))
}

func TestRandomCommitmentsProcess(t *testing.T) {
	key, _ := wallet.Base58CheckDeserialize("112t8rnXCqbbNYBquntyd6EvDT4WiDDQw84ZSRDKmazkqrzi6w8rWyCVt7QEZgAiYAV4vhJiX7V9MCfuj4hGLoDN7wdU1LoWGEFpLs59X7K3")
	_ = key.KeySet.ImportFromPrivateKey(&key.KeySet.PrivateKey)
	paymentAddress := key.KeySet.PaymentAddress
	tx, _ := BuildCoinbaseTx(&paymentAddress, 10, &key.KeySet.PrivateKey, db, nil)
	db.StoreCommitments(common.Hash{}, paymentAddress.Pk, [][]byte{tx.Proof.OutputCoins[0].CoinDetails.CoinCommitment.Compress()}, 0)

	in := ConvertOutputCoinToInputCoin(tx.Proof.OutputCoins)
	cmmIndexs, myCommIndex, cmm := RandomCommitmentsProcess(in, 0, db, 0, &common.Hash{})
	assert.Equal(t, 8, len(cmmIndexs))
	assert.Equal(t, 1, len(myCommIndex))
	assert.Equal(t, 8, len(cmm))
}
