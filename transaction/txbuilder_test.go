package transaction

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	_ "github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

var db database.DatabaseInterface
var _ = func() (_ struct{}) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_")
	if err != nil {
		log.Fatalf("failed to create temp dir: %+v", err)
	}
	log.Println(dbPath)
	db, err = database.Open("leveldb", dbPath)
	if err != nil {
		log.Fatalf("could not open db path: %s, %+v", dbPath, err)
	}
	database.Logger.Init(common.NewBackend(nil).Logger("db", true))
	Logger.Init(common.NewBackend(nil).Logger("tx", true))
	return
}()

func TestBuildCoinbaseTx(t *testing.T) {
	key, err := wallet.Base58CheckDeserialize("112t8rnXCqbbNYBquntyd6EvDT4WiDDQw84ZSRDKmazkqrzi6w8rWyCVt7QEZgAiYAV4vhJiX7V9MCfuj4hGLoDN7wdU1LoWGEFpLs59X7K3")
	assert.Equal(t, nil, err)
	err = key.KeySet.ImportFromPrivateKey(&key.KeySet.PrivateKey)
	assert.Equal(t, nil, err)
	paymentAddress := key.KeySet.PaymentAddress
	tx, err := BuildCoinbaseTx(&paymentAddress, 10, &key.KeySet.PrivateKey, db, nil)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, tx)
	assert.Equal(t, uint64(10), tx.Proof.OutputCoins[0].CoinDetails.Value)
	assert.Equal(t, string(key.KeySet.PaymentAddress.Pk[:]), string(tx.Proof.OutputCoins[0].CoinDetails.PublicKey.Compress()[:]))

	paymentAddress.Pk[0] = 1
	_, err = BuildCoinbaseTx(&paymentAddress, 10, &key.KeySet.PrivateKey, db, nil)
	assert.NotEqual(t, nil, err)
}

func TestBuildCoinbaseTxByCoinID(t *testing.T) {
	key, err := wallet.Base58CheckDeserialize("112t8rnXCqbbNYBquntyd6EvDT4WiDDQw84ZSRDKmazkqrzi6w8rWyCVt7QEZgAiYAV4vhJiX7V9MCfuj4hGLoDN7wdU1LoWGEFpLs59X7K3")
	assert.Equal(t, nil, err)
	err = key.KeySet.ImportFromPrivateKey(&key.KeySet.PrivateKey)
	assert.Equal(t, nil, err)
	paymentAddress := key.KeySet.PaymentAddress

	tx, err := BuildCoinbaseTxByCoinID(&paymentAddress, 10, &key.KeySet.PrivateKey, db, nil, common.Hash{}, NormalCoinType, "PRV", 0)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, tx)
	assert.Equal(t, uint64(10), tx.(*Tx).Proof.OutputCoins[0].CoinDetails.Value)
	assert.Equal(t, common.PRVCoinID.String(), tx.GetTokenID().String())

	txCustomToken, err := BuildCoinbaseTxByCoinID(&paymentAddress, 10, &key.KeySet.PrivateKey, db, nil, common.Hash{1}, CustomTokenType, "Custom Token", 0)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, tx)
	assert.Equal(t, uint64(10), txCustomToken.(*TxCustomToken).TxTokenData.Vouts[0].Value)
	assert.Equal(t, common.Hash{1}.String(), txCustomToken.GetTokenID().String())

	txCustomTokenPrivacy, err := BuildCoinbaseTxByCoinID(&paymentAddress, 10, &key.KeySet.PrivateKey, db, nil, common.Hash{2}, CustomTokenPrivacyType, "Custom Token", 0)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, tx)
	assert.Equal(t, uint64(10), txCustomTokenPrivacy.(*TxCustomTokenPrivacy).TxTokenPrivacyData.TxNormal.Proof.OutputCoins[0].CoinDetails.Value)
	assert.Equal(t, common.Hash{2}.String(), txCustomTokenPrivacy.GetTokenID().String())
}
