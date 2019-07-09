package transaction

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestUnmarshalJSON(t *testing.T) {
	key, err := wallet.Base58CheckDeserialize("112t8rnXCqbbNYBquntyd6EvDT4WiDDQw84ZSRDKmazkqrzi6w8rWyCVt7QEZgAiYAV4vhJiX7V9MCfuj4hGLoDN7wdU1LoWGEFpLs59X7K3")
	assert.Equal(t, nil, err)
	err = key.KeySet.ImportFromPrivateKey(&key.KeySet.PrivateKey)
	assert.Equal(t, nil, err)
	paymentAddress := key.KeySet.PaymentAddress
	responseMeta, err := metadata.NewWithDrawRewardResponse(&common.Hash{})
	tx, err := BuildCoinbaseTxByCoinID(&paymentAddress, 10, &key.KeySet.PrivateKey, db, responseMeta, common.Hash{}, NormalCoinType, "PRV", 0)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, tx)
	assert.Equal(t, uint64(10), tx.(*Tx).Proof.OutputCoins[0].CoinDetails.Value)
	assert.Equal(t, common.PRVCoinID.String(), tx.GetTokenID().String())

	jsonStr, err := json.Marshal(tx)
	assert.Equal(t, nil, err)
	fmt.Println(string(jsonStr))

	tx1 := Tx{}
	err = json.Unmarshal(jsonStr, &tx1)
	assert.Equal(t, nil, err)
	assert.Equal(t, uint64(10), tx1.Proof.OutputCoins[0].CoinDetails.Value)
	assert.Equal(t, common.PRVCoinID.String(), tx1.GetTokenID().String())
}

func TestInitTx(t *testing.T) {
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

	tx1 := Tx{}
	in1[0].CoinDetails.SerialNumber = privacy.PedCom.G[privacy.SK].Derive(new(big.Int).SetBytes(key.KeySet.PrivateKey),
		in1[0].CoinDetails.SNDerivator)
	paymentAddress2, _ := wallet.Base58CheckDeserialize("1Uv3BkYiWy9Mjt1yBa4dXBYKo3az22TeCVEpeXN93ieJ8qhrTDuUZBzsPZWjjP2AeRQnjw1y18iFPHTRuAqqufwVC1vNUAWs4wHFbbWC2")
	err = tx1.Init(&key.KeySet.PrivateKey, []*privacy.PaymentInfo{{PaymentAddress: paymentAddress2.KeySet.PaymentAddress, Amount: 5}}, in1, 1, false, db, nil, nil)
	if err.(*TransactionError) != nil {
		t.Error(err)
	}
	unique, pubk, amount := tx1.GetUniqueReceiver()
	assert.Equal(t, true, unique)
	assert.Equal(t, string(pubk[:]), string(paymentAddress2.KeySet.PaymentAddress.Pk[:]))
	assert.Equal(t, uint64(5), amount)

	unique, pubk, amount, coinID := tx1.GetTransferData()
	assert.Equal(t, true, unique)
	assert.Equal(t, common.PRVCoinID.String(), coinID.String())
	assert.Equal(t, string(pubk[:]), string(paymentAddress2.KeySet.PaymentAddress.Pk[:]))

	a, b := tx1.GetTokenReceivers()
	assert.Equal(t, 0, len(a))
	assert.Equal(t, 0, len(b))

	e, d, c := tx1.GetTokenUniqueReceiver()
	assert.Equal(t, false, e)
	assert.Equal(t, 0, len(d))
	assert.Equal(t, uint64(0), c)

	tx3 := Tx{}
	db.StoreCommitments(common.PRVCoinID, paymentAddress.Pk, [][]byte{tx2.(*Tx).Proof.OutputCoins[0].CoinDetails.CoinCommitment.Compress()}, 6)
	err = tx3.Init(&key.KeySet.PrivateKey, []*privacy.PaymentInfo{{PaymentAddress: paymentAddress, Amount: 5}}, in1, 1, true, db, nil, nil)
	if err.(*TransactionError) != nil {
		t.Error(err)
	}

	valid, err = tx3.ValidateSanityData(nil)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, valid)

	verified, err := tx3.ValidateTransaction(true, db, 6, &common.PRVCoinID)
	assert.Equal(t, true, verified)

	tx3.ValidateConstDoubleSpendWithBlockchain(nil, 6, db)
	tx3.ValidateTxWithBlockChain(nil, 6, db)
}
