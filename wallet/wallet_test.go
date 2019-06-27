package wallet

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"errors"
)

var dataDir string
var wallet *Wallet
var _ = func() (_ struct{}) {
	fmt.Println("This runs before init()!")
	dataDir, _ = os.Getwd()
	wallet = new(Wallet)

	// set config wallet
	dataDir := filepath.Join(common.AppDataDir("incognito", false), "data")
	dataFile := "wallet"
	walletConf := &WalletConfig{
		DataDir:        dataDir,
		DataFile:       dataFile,
		DataPath:       filepath.Join(dataDir, dataFile),
		IncrementalFee: 0, // 0 mili PRV
	}

	wallet.SetConfig(walletConf)

	Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

/*
		Unit test for Init function
 */
func TestInit(t *testing.T){
	data := []struct {
		passPhrase string
		numOfAccount uint32
		name string
	}{
		{"", uint32(2), "Wallet1"},
		{"12345678", uint32(3), "Wallet2"},
		{"12345678", uint32(10), "Wallet3"},
	}

	for _, item := range data {
		err := wallet.Init(item.passPhrase, item.numOfAccount, item.name)

		assert.Equal(t, nil, err)
		assert.Equal(t, int(item.numOfAccount), len(wallet.MasterAccount.Child))
		assert.Equal(t, item.name, wallet.Name)
		assert.Equal(t, item.passPhrase, wallet.PassPhrase)
		assert.Equal(t, SeedKeyLen, len(wallet.Seed))
		assert.Greater(t, len(wallet.Mnemonic), 0)
	}
}

func TestInitWithNumAccIsZero(t *testing.T){
	passPhrase :=  "12345678"
	numOfAccount :=  uint32(0)
	name := "Wallet 1"

	err := wallet.Init(passPhrase, numOfAccount, name)

	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(wallet.MasterAccount.Child))
}

func TestInitWithEmptyName(t *testing.T){
	passPhrase :=  "12345678"
	numOfAccount :=  uint32(3)
	name := ""

	err := wallet.Init(passPhrase, numOfAccount, name)

	assert.Equal(t, NewWalletError(EmptyWalletNameErr, nil), err)
}

/*
		Unit test for CreateNewAccount function
 */

func TestCreateNewAccount(t *testing.T){
	data := []struct {
		accountName string
		shardID byte
	}{
		{"Acc A", byte(0)},
		{"Acc B", byte(1)},
		{"Acc C", byte(2)},
		{"Acc D", byte(3)},
	}

	wallet.Init("", 0, "Wallet")

	numAccount := len(wallet.MasterAccount.Child)

	for _, item := range data {
		newAccount, err := wallet.CreateNewAccount(item.accountName, &item.shardID)
		actualShardID := common.GetShardIDFromLastByte(newAccount.Key.KeySet.PaymentAddress.Pk[len(newAccount.Key.KeySet.PaymentAddress.Pk) -1])

		assert.Equal(t, nil, err)
		assert.Equal(t, numAccount + 1, len(wallet.MasterAccount.Child))
		assert.Equal(t, item.accountName, newAccount.Name)
		assert.Equal(t, item.shardID, actualShardID)
		assert.Equal(t, false, newAccount.IsImported)
		assert.Equal(t, 0, len(newAccount.Child))
		assert.Equal(t, ChildNumberLen, len(newAccount.Key.ChildNumber))
		assert.Equal(t, ChainCodeLen, len(newAccount.Key.ChainCode))
		assert.Equal(t, privacy.PublicKeySize, len(newAccount.Key.KeySet.PaymentAddress.Pk))
		assert.Equal(t, privacy.TransmissionKeySize, len(newAccount.Key.KeySet.PaymentAddress.Tk))
		assert.Equal(t, privacy.PrivateKeySize, len(newAccount.Key.KeySet.PrivateKey))
		assert.Equal(t, privacy.ReceivingKeySize, len(newAccount.Key.KeySet.ReadonlyKey.Rk))

		numAccount++
	}
}

func TestCreateNewAccountWithEmptyName(t *testing.T){
	// init wallet
	wallet.Init("", 0, "Wallet")

	numAccount := len(wallet.MasterAccount.Child)

	// create new account with empty name
	accountName := ""
	shardID := byte(0)

	newAccount, err := wallet.CreateNewAccount(accountName, &shardID)
	actualShardID := common.GetShardIDFromLastByte(newAccount.Key.KeySet.PaymentAddress.Pk[len(newAccount.Key.KeySet.PaymentAddress.Pk) -1])

	assert.Equal(t, nil, err)
	assert.Equal(t, numAccount + 1, len(wallet.MasterAccount.Child))
	assert.Equal(t, "AccountWallet " + strconv.Itoa(numAccount), newAccount.Name)
	assert.Equal(t, shardID, actualShardID)
	assert.Equal(t, false, newAccount.IsImported)
	assert.Equal(t, 0, len(newAccount.Child))
	assert.Equal(t, ChildNumberLen, len(newAccount.Key.ChildNumber))
	assert.Equal(t, ChainCodeLen, len(newAccount.Key.ChainCode))
	assert.Equal(t, privacy.PublicKeySize, len(newAccount.Key.KeySet.PaymentAddress.Pk))
	assert.Equal(t, privacy.TransmissionKeySize, len(newAccount.Key.KeySet.PaymentAddress.Tk))
	assert.Equal(t, privacy.PrivateKeySize, len(newAccount.Key.KeySet.PrivateKey))
	assert.Equal(t, privacy.ReceivingKeySize, len(newAccount.Key.KeySet.ReadonlyKey.Rk))
}

func TestCreateNewAccountWithNilShardID(t *testing.T){
	// init wallet
	wallet.Init("", 0, "Wallet")
	numAccount := len(wallet.MasterAccount.Child)

	// create new account with empty name
	accountName := "Acc A"

	newAccount, err := wallet.CreateNewAccount(accountName, nil)
	actualShardID := common.GetShardIDFromLastByte(newAccount.Key.KeySet.PaymentAddress.Pk[len(newAccount.Key.KeySet.PaymentAddress.Pk) -1])

	assert.Equal(t, nil, err)
	assert.Equal(t, numAccount + 1, len(wallet.MasterAccount.Child))
	assert.Equal(t, accountName, newAccount.Name)
	assert.GreaterOrEqual(t, actualShardID, byte(0))
	assert.Equal(t, false, newAccount.IsImported)
	assert.Equal(t, 0, len(newAccount.Child))
	assert.Equal(t, ChildNumberLen, len(newAccount.Key.ChildNumber))
	assert.Equal(t, ChainCodeLen, len(newAccount.Key.ChainCode))
	assert.Equal(t, privacy.PublicKeySize, len(newAccount.Key.KeySet.PaymentAddress.Pk))
	assert.Equal(t, privacy.TransmissionKeySize, len(newAccount.Key.KeySet.PaymentAddress.Tk))
	assert.Equal(t, privacy.PrivateKeySize, len(newAccount.Key.KeySet.PrivateKey))
	assert.Equal(t, privacy.ReceivingKeySize, len(newAccount.Key.KeySet.ReadonlyKey.Rk))
}


func TestWalletCreateNewAccountDuplicateAccountName(t *testing.T) {
	wallet.Init("", 0, "Wallet")

	// create the first account with name = "Acc A"
	accountName := "Acc E"
	shardID := byte(0)

	wallet.CreateNewAccount(accountName, &shardID)

	// create new account with existed name
	_, err := wallet.CreateNewAccount(accountName, &shardID)

	assert.Equal(t,  NewWalletError(ExistedAccountNameErr, nil), err)
}

// max len of name account???

/*
		Unit test for ExportAccount function
 */

func TestWalletExportAccount(t *testing.T) {
	accountName := "Acc F"
	shardID := byte(0)
	wallet.CreateNewAccount(accountName, &shardID)

	for i := range wallet.MasterAccount.Child {
		res := wallet.ExportAccount(uint32(i))
		assert.Equal(t, PrivateKeySerializedLen, len(res))
	}
}

func TestWalletExportAccountWithWrongIndex(t *testing.T) {
	accountName := "Acc G"
	shardID := byte(0)
	wallet.CreateNewAccount(accountName, &shardID)

	res := wallet.ExportAccount(uint32(len(wallet.MasterAccount.Child)))
	assert.Equal(t,"", res)
}

/*
		Unit test for ImportAccount function
 */

func TestWalletImportAccount(t *testing.T){
	data := []struct {
		privateKeyStr string
		accountName string
		passPhrase string
	}{
		{"112t8rnY6orkxdArx6fH7xV8C3kiEAJMuDmf7ptrgQ3iqo6VKzSzippYzqT3kPqCXyVmb4iP5AnyTzD1thrhybntuWockJrtYHq6CeSWK5VZ", "Acc A", "123"},
		{"112t8rnYJncU5TRMexdSX2X9a58c9dKPfzWMEaS7AXY3WniXbVUXvDVmZaKms2QEXtviEUKPdrqq3auNqZB8wQPtuXv8JfzprtMtgdGRiFij", "Acc B", "123"},
		{"112t8rnYh9nB6vgnPrsnoMe5Sd39fGUTvyrBtKGN82LLXEcr2EJ2jR2c4rLtEHauCCcaXvwHtYem865L95jKBNUFGrd8mFaExvxtmjuZNqNF", "Acc C", "123"},
	}

	wallet.Init("123", 0, "Wallet")

	numAccount := len(wallet.MasterAccount.Child)

	for _, item := range data {
		newAccount, err := wallet.ImportAccount(item.privateKeyStr, item.accountName, item.passPhrase)
		keyWallet, _ := Base58CheckDeserialize(item.privateKeyStr)

		// TODO:
		assert.Equal(t, nil, err)
		assert.Equal(t, numAccount + 1, len(wallet.MasterAccount.Child))
		assert.Equal(t, item.accountName, newAccount.Name)
		assert.Equal(t, true, newAccount.IsImported)
		assert.Equal(t, 0, len(newAccount.Child))
		assert.Equal(t, ChildNumberLen, len(newAccount.Key.ChildNumber))
		assert.Equal(t, ChainCodeLen, len(newAccount.Key.ChainCode))
		assert.Equal(t,keyWallet.KeySet.PrivateKey, newAccount.Key.KeySet.PrivateKey)

		numAccount++
	}
}

func TestWalletImportAccountWithWrongPrivKeyStr(t *testing.T){
	privateKeyStr := "abc"
	accountName := "Acc A"
	passPhrase := "123"

	wallet.Init(passPhrase, 0, "Wallet")

	_, err := wallet.ImportAccount(privateKeyStr, accountName, passPhrase)
	assert.Equal(t, errors.New("invalid format: version and/or checksum bytes missing"), err)
}

func TestWalletImportAccountWithExistedPrivKeyStr(t *testing.T){
	privateKeyStr := "112t8rnY6orkxdArx6fH7xV8C3kiEAJMuDmf7ptrgQ3iqo6VKzSzippYzqT3kPqCXyVmb4iP5AnyTzD1thrhybntuWockJrtYHq6CeSWK5VZ"
	accountName := "Acc A"
	passPhrase := "123"

	wallet.Init(passPhrase, 0, "Wallet")
	wallet.ImportAccount(privateKeyStr, accountName, passPhrase)

	_, err := wallet.ImportAccount(privateKeyStr, "Acc B", passPhrase)
	assert.Equal(t, NewWalletError(ExistedAccountErr, nil), err)
}

func TestWalletImportAccountWithExistedAccountName(t *testing.T){
	privateKeyStr := "112t8rnY6orkxdArx6fH7xV8C3kiEAJMuDmf7ptrgQ3iqo6VKzSzippYzqT3kPqCXyVmb4iP5AnyTzD1thrhybntuWockJrtYHq6CeSWK5VZ"
	accountName := "Acc A"
	passPhrase := "123"

	wallet.Init(passPhrase, 0, "Wallet")
	wallet.CreateNewAccount(accountName, nil)

	_, err := wallet.ImportAccount(privateKeyStr, accountName, passPhrase)
	fmt.Printf("err: %v\n", err)
	assert.Equal(t, NewWalletError(ExistedAccountNameErr, nil), err)
}

func TestWalletImportAccountWithUnmatchedPassPhrase(t *testing.T){
	privateKeyStr := "112t8rnY6orkxdArx6fH7xV8C3kiEAJMuDmf7ptrgQ3iqo6VKzSzippYzqT3kPqCXyVmb4iP5AnyTzD1thrhybntuWockJrtYHq6CeSWK5VZ"
	accountName := "Acc A"
	passPhrase := "123"

	wallet.Init(passPhrase, 0, "Wallet")

	_, err := wallet.ImportAccount(privateKeyStr, accountName, "1234")
	fmt.Printf("err: %v\n", err)
	assert.Equal(t, NewWalletError(WrongPassphraseErr, nil), err)
}




