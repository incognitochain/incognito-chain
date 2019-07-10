package transaction

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestTxCustomToken(t *testing.T) {
	key, err := wallet.Base58CheckDeserialize("112t8rnXCqbbNYBquntyd6EvDT4WiDDQw84ZSRDKmazkqrzi6w8rWyCVt7QEZgAiYAV4vhJiX7V9MCfuj4hGLoDN7wdU1LoWGEFpLs59X7K3")
	assert.Equal(t, nil, err)
	err = key.KeySet.ImportFromPrivateKey(&key.KeySet.PrivateKey)
	assert.Equal(t, nil, err)
	paymentAddress := key.KeySet.PaymentAddress

	tx2, err := BuildCoinbaseTxByCoinID(&paymentAddress, 1000, &key.KeySet.PrivateKey, db, nil, common.Hash{}, NormalCoinType, "PRV", 0)

	valid, err := tx2.ValidateSanityData(nil)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, valid)

	in1 := ConvertOutputCoinToInputCoin(tx2.(*Tx).Proof.OutputCoins)
	in1[0].CoinDetails.SerialNumber = privacy.PedCom.G[privacy.SK].Derive(new(big.Int).SetBytes(key.KeySet.PrivateKey),
		in1[0].CoinDetails.SNDerivator)
	tx := TxCustomToken{}
	err = tx.Init(&key.KeySet.PrivateKey,
		[]*privacy.PaymentInfo{{Amount: 10, PaymentAddress: paymentAddress}},
		in1,
		0,
		&CustomTokenParamTx{
			PropertyID:  common.PRVCoinID.String(),
			Amount:      10,
			TokenTxType: CustomTokenInit,
			Receiver:    []TxTokenVout{{PaymentAddress: paymentAddress, Value: 10}},
		},
		db,
		nil,
		false,
		6)
	if err.(*TransactionError) != nil {
		t.Error(err)
	}

	db.StoreCustomToken(common.PRVCoinID, tx.Hash()[:])

	err = tx.Init(&key.KeySet.PrivateKey,
		[]*privacy.PaymentInfo{{Amount: 10, PaymentAddress: paymentAddress}},
		in1,
		0,
		&CustomTokenParamTx{
			PropertyID:  common.PRVCoinID.String(),
			Amount:      10,
			TokenTxType: CustomTokenTransfer,
			Receiver:    []TxTokenVout{{PaymentAddress: paymentAddress, Value: 10}},
			vins:        []TxTokenVin{{PaymentAddress: paymentAddress, VoutIndex: 0, TxCustomTokenID: *tx.Hash()}},
		},
		db,
		nil,
		false,
		6)
	if err.(*TransactionError) != nil {
		t.Error(err)
	}

	assert.Equal(t, string(tx.TxTokenData.Vins[0].PaymentAddress.Pk[:]), string(tx.GetSender()))

	paymentAddress2, _ := wallet.Base58CheckDeserialize("1Uv3BkYiWy9Mjt1yBa4dXBYKo3az22TeCVEpeXN93ieJ8qhrTDuUZBzsPZWjjP2AeRQnjw1y18iFPHTRuAqqufwVC1vNUAWs4wHFbbWC2")
	err = tx.Init(&key.KeySet.PrivateKey,
		[]*privacy.PaymentInfo{{Amount: 10, PaymentAddress: paymentAddress}},
		in1,
		0,
		&CustomTokenParamTx{
			PropertyID:  common.PRVCoinID.String(),
			Amount:      10,
			TokenTxType: CustomTokenCrossShard,
			Receiver:    []TxTokenVout{{PaymentAddress: paymentAddress2.KeySet.PaymentAddress, Value: 10}},
			vins:        []TxTokenVin{{PaymentAddress: paymentAddress, VoutIndex: 0, TxCustomTokenID: *tx.Hash()}},
		},
		db,
		nil,
		false,
		6)
	if err.(*TransactionError) != nil {
		t.Error(err)
	}

	sign, err := tx.GetTxCustomTokenSignature(key.KeySet)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, privacy.SigNoPrivacySize, len(sign))

	isP := tx.IsPrivacy()
	assert.Equal(t, false, isP)

	typeTx := tx.ValidateType()
	assert.Equal(t, true, typeTx)

	assert.Equal(t, tx.Proof, tx.GetProof())

	_, amount := tx.GetTokenReceivers()
	assert.Equal(t, uint64(10), amount[0])

	unique, _, _ := tx.GetTokenUniqueReceiver()
	assert.Equal(t, true, unique)
}
