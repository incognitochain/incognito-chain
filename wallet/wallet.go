package wallet

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
)

type AccountWallet struct {
	Name       string
	Key        KeyWallet
	Child      []AccountWallet
	IsImported bool
}
type Wallet struct {
	Seed          []byte
	Entropy       []byte
	PassPhrase    string
	Mnemonic      string
	MasterAccount AccountWallet
	Name          string
	Config        *WalletConfig
}

type WalletConfig struct {
	DataDir        string
	DataFile       string
	DataPath       string
	IncrementalFee uint64
}

func (wallet *Wallet) Init(passPhrase string, numOfAccount uint32, name string) (error) {
	mnemonicGen := MnemonicGenerator{}
	wallet.Name = name
	wallet.Entropy, _ = mnemonicGen.NewEntropy(128)
	wallet.Mnemonic, _ = mnemonicGen.NewMnemonic(wallet.Entropy)
	wallet.Seed = mnemonicGen.NewSeed(wallet.Mnemonic, passPhrase)
	wallet.PassPhrase = passPhrase

	masterKey, err := NewMasterKey(wallet.Seed)
	if err != nil {
		return err
	}
	wallet.MasterAccount = AccountWallet{
		Key:   *masterKey,
		Child: make([]AccountWallet, 0),
		Name:  "master",
	}

	if numOfAccount == 0 {
		numOfAccount = 1
	}

	for i := uint32(0); i < numOfAccount; i++ {
		childKey, _ := wallet.MasterAccount.Key.NewChildKey(i)
		account := AccountWallet{
			Key:   *childKey,
			Child: make([]AccountWallet, 0),
			Name:  fmt.Sprintf("AccountWallet %d", i),
		}
		wallet.MasterAccount.Child = append(wallet.MasterAccount.Child, account)
	}

	return nil
}

func (wallet *Wallet) CreateNewAccount(accountName string, shardID byte) *AccountWallet {
	newIndex := uint32(len(wallet.MasterAccount.Child))
	var childKey *KeyWallet
	for {
		childKey, _ = wallet.MasterAccount.Key.NewChildKey(newIndex)
		lastByte := childKey.KeySet.PaymentAddress.Pk[len(childKey.KeySet.PaymentAddress.Pk)-1]
		if lastByte == shardID {
			break
		}
		newIndex += 1
	}

	if accountName == "" {
		accountName = fmt.Sprintf("AccountWallet %d", len(wallet.MasterAccount.Child))
	}
	account := AccountWallet{
		Key:   *childKey,
		Child: make([]AccountWallet, 0),
		Name:  accountName,
	}
	wallet.MasterAccount.Child = append(wallet.MasterAccount.Child, account)
	wallet.Save(wallet.PassPhrase)
	return &account
}

func (wallet *Wallet) ExportAccount(childIndex uint32) string {
	return wallet.MasterAccount.Child[childIndex].Key.Base58CheckSerialize(PriKeyType)
}

func (wallet *Wallet) RemoveAccount(privateKeyStr string, accountName string, passPhrase string) error {
	if passPhrase != wallet.PassPhrase {
		return NewWalletError(WrongPassphraseErr, nil)
	}
	for i, account := range wallet.MasterAccount.Child {
		if account.Key.Base58CheckSerialize(PriKeyType) == privateKeyStr {
			wallet.MasterAccount.Child = append(wallet.MasterAccount.Child[:i], wallet.MasterAccount.Child[i+1:]...)
			wallet.Save(passPhrase)
			return nil
		}
	}
	return NewWalletError(UnexpectedErr, errors.New("Not found"))
}

func (wallet *Wallet) ImportAccount(privateKeyStr string, accountName string, passPhrase string) (*AccountWallet, error) {
	if passPhrase != wallet.PassPhrase {
		return nil, NewWalletError(WrongPassphraseErr, nil)
	}

	for _, account := range wallet.MasterAccount.Child {
		if account.Key.Base58CheckSerialize(PriKeyType) == privateKeyStr {
			return nil, NewWalletError(ExistedAccountErr, nil)
		}
		if account.Name == accountName {
			return nil, NewWalletError(ExistedAccountNameErr, nil)
		}
	}

	keyWallet, err := Base58CheckDeserialize(privateKeyStr)
	if err != nil {
		return nil, err
	}
	keyWallet.KeySet.ImportFromPrivateKey(&keyWallet.KeySet.PrivateKey)

	Logger.log.Infof("Pub-key : %s", keyWallet.Base58CheckSerialize(PaymentAddressType))
	Logger.log.Infof("Readonly-key : %s", keyWallet.Base58CheckSerialize(ReadonlyKeyType))

	account := AccountWallet{
		Key:        *keyWallet,
		Child:      make([]AccountWallet, 0),
		IsImported: true,
		Name:       accountName,
	}
	wallet.MasterAccount.Child = append(wallet.MasterAccount.Child, account)
	err = wallet.Save(wallet.PassPhrase)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (wallet *Wallet) Save(password string) error {
	if password == "" {
		password = wallet.PassPhrase
	}

	// parse to byte[]
	data, err := json.Marshal(*wallet)
	if err != nil {
		Logger.log.Error(err)
		return NewWalletError(UnexpectedErr, err)
	}

	// encrypt
	cipherText, err := AES{}.Encrypt(password, data)
	if err != nil {
		Logger.log.Error(err)
		return NewWalletError(UnexpectedErr, err)
	}
	// and
	// save file
	err = ioutil.WriteFile(wallet.Config.DataPath, []byte(cipherText), 0644)
	if err != nil {
		return NewWalletError(UnexpectedErr, err)
	}
	return nil
}

func (wallet *Wallet) LoadWallet(password string) error {
	// read file and decrypt
	bytesData, err := ioutil.ReadFile(wallet.Config.DataPath)
	if err != nil {
		return NewWalletError(UnexpectedErr, err)
	}
	bufBytes, err := AES{}.Decrypt(password, string(bytesData))
	if err != nil {
		return NewWalletError(UnexpectedErr, err)
	}

	// read to struct
	err = json.Unmarshal(bufBytes, &wallet)
	if err != nil {
		return NewWalletError(UnexpectedErr, err)
	}
	return nil
}

func (wallet *Wallet) DumpPrivkey(addressP string) (KeySerializedData) {
	for _, account := range wallet.MasterAccount.Child {
		address := account.Key.Base58CheckSerialize(PaymentAddressType)
		if address == addressP {
			key := KeySerializedData{
				PrivateKey: account.Key.Base58CheckSerialize(PriKeyType),
			}
			return key
		}
	}
	return KeySerializedData{}
}

func (wallet *Wallet) GetAccountAddress(accountParam string, shardID byte) (KeySerializedData) {
	for _, account := range wallet.MasterAccount.Child {
		if account.Name == accountParam {
			key := KeySerializedData{
				PaymentAddress: account.Key.Base58CheckSerialize(PaymentAddressType),
				Pubkey:         hex.EncodeToString(account.Key.KeySet.PaymentAddress.Pk),
				ReadonlyKey:    account.Key.Base58CheckSerialize(ReadonlyKeyType),
			}
			return key
		}
	}
	newAccount := wallet.CreateNewAccount(accountParam, shardID)
	key := KeySerializedData{
		PaymentAddress: newAccount.Key.Base58CheckSerialize(PaymentAddressType),
		Pubkey:         hex.EncodeToString(newAccount.Key.KeySet.PaymentAddress.Pk),
		ReadonlyKey:    newAccount.Key.Base58CheckSerialize(ReadonlyKeyType),
	}
	return key
}

func (wallet *Wallet) GetAddressesByAccount(accountParam string) ([]KeySerializedData) {
	result := make([]KeySerializedData, 0)
	for _, account := range wallet.MasterAccount.Child {
		if account.Name == accountParam {
			item := KeySerializedData{
				PaymentAddress: account.Key.Base58CheckSerialize(PaymentAddressType),
				Pubkey:         hex.EncodeToString(account.Key.KeySet.PaymentAddress.Pk),
				ReadonlyKey:    account.Key.Base58CheckSerialize(ReadonlyKeyType),
			}
			result = append(result, item)
		}
	}
	return result
}

func (wallet *Wallet) ListAccounts() map[string]AccountWallet {
	result := make(map[string]AccountWallet)
	for _, account := range wallet.MasterAccount.Child {
		result[account.Name] = account
	}
	return result
}

func (wallet *Wallet) ContainPubKey(pubKey []byte) bool {
	for _, account := range wallet.MasterAccount.Child {
		if bytes.Equal(account.Key.KeySet.PaymentAddress.Pk[:], pubKey) {
			return true
		}
	}
	return false
}
