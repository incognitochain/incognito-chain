package wallet

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"github.com/pkg/errors"
)

type Account struct {
	Name       string
	Key        Key
	Child      []Account
	IsImported bool
}
type Wallet struct {
	Seed          []byte
	Entropy       []byte
	PassPhrase    string
	Mnemonic      string
	MasterAccount Account
	Name          string
	Config        *WalletConfig
}

type WalletConfig struct {
	DataDir        string
	DataFile       string
	DataPath       string
	IncrementalFee uint64
}

func (self *Wallet) Init(passPhrase string, numOfAccount uint32, name string) (error) {
	mnemonicGen := MnemonicGenerator{}
	self.Name = name
	self.Entropy, _ = mnemonicGen.NewEntropy(128)
	self.Mnemonic, _ = mnemonicGen.NewMnemonic(self.Entropy)
	self.Seed = mnemonicGen.NewSeed(self.Mnemonic, passPhrase)
	self.PassPhrase = passPhrase

	masterKey, err := NewMasterKey(self.Seed)
	if err != nil {
		return err
	}
	self.MasterAccount = Account{
		Key:   *masterKey,
		Child: make([]Account, 0),
		Name:  "master",
	}

	if numOfAccount == 0 {
		numOfAccount = 1
	}

	for i := uint32(0); i < numOfAccount; i++ {
		childKey, _ := self.MasterAccount.Key.NewChildKey(i)
		account := Account{
			Key:   *childKey,
			Child: make([]Account, 0),
			Name:  fmt.Sprintf("Account %d", i),
		}
		self.MasterAccount.Child = append(self.MasterAccount.Child, account)
	}

	return nil
}

func (self *Wallet) CreateNewAccount(accountName string) *Account {
	newIndex := uint32(len(self.MasterAccount.Child))
	childKey, _ := self.MasterAccount.Key.NewChildKey(newIndex)
	if accountName == "" {
		accountName = fmt.Sprintf("Account %d", len(self.MasterAccount.Child))
	}
	account := Account{
		Key:   *childKey,
		Child: make([]Account, 0),
		Name:  accountName,
	}
	self.MasterAccount.Child = append(self.MasterAccount.Child, account)
	self.Save(self.PassPhrase)
	return &account
}

func (self *Wallet) ExportAccount(childIndex uint32) string {
	return self.MasterAccount.Child[childIndex].Key.Base58CheckSerialize(PriKeyType)
}

func (self *Wallet) RemoveAccount(privateKeyStr string, accountName string, passPhrase string) error {
	if passPhrase != self.PassPhrase {
		return NewWalletError(WrongPassphraseErr, nil)
	}
	for i, account := range self.MasterAccount.Child {
		if account.Key.Base58CheckSerialize(PriKeyType) == privateKeyStr {
			self.MasterAccount.Child = append(self.MasterAccount.Child[:i], self.MasterAccount.Child[i+1:]...)
			self.Save(passPhrase)
			return nil
		}
	}
	return NewWalletError(UnexpectedErr, errors.New("Not found"))
}

func (self *Wallet) ImportAccount(privateKeyStr string, accountName string, passPhrase string) (*Account, error) {
	if passPhrase != self.PassPhrase {
		return nil, NewWalletError(WrongPassphraseErr, nil)
	}

	for _, account := range self.MasterAccount.Child {
		if account.Key.Base58CheckSerialize(PriKeyType) == privateKeyStr {
			return nil, NewWalletError(ExistedAccountErr, nil)
		}
		if account.Name == accountName {
			return nil, NewWalletError(ExistedAccountNameErr, nil)
		}
	}

	priKey, err := Base58CheckDeserialize(privateKeyStr)
	if err != nil {
		return nil, err
	}
	priKey.KeySet.ImportFromPrivateKey(&priKey.KeySet.PrivateKey)

	Logger.log.Infof("Pub-key : %s", priKey.Base58CheckSerialize(PaymentAddressType))
	Logger.log.Infof("Readonly-key : %s", priKey.Base58CheckSerialize(ReadonlyKeyType))

	account := Account{
		Key:        *priKey,
		Child:      make([]Account, 0),
		IsImported: true,
		Name:       accountName,
	}
	self.MasterAccount.Child = append(self.MasterAccount.Child, account)
	err = self.Save(self.PassPhrase)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (self *Wallet) Save(password string) error {
	if password == "" {
		password = self.PassPhrase
	}

	// parse to byte[]
	data, err := json.Marshal(*self)
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
	err = ioutil.WriteFile(self.Config.DataPath, []byte(cipherText), 0644)
	if err != nil {
		return NewWalletError(UnexpectedErr, err)
	}
	return nil
}

func (self *Wallet) LoadWallet(password string) error {
	// read file and decrypt
	bytesData, err := ioutil.ReadFile(self.Config.DataPath)
	if err != nil {
		return NewWalletError(UnexpectedErr, err)
	}
	bufBytes, err := AES{}.Decrypt(password, string(bytesData))
	if err != nil {
		return NewWalletError(UnexpectedErr, err)
	}

	// read to struct
	err = json.Unmarshal(bufBytes, &self)
	if err != nil {
		return NewWalletError(UnexpectedErr, err)
	}
	return nil
}

func (self *Wallet) DumpPrivkey(addressP string) (KeySerializedData) {
	for _, account := range self.MasterAccount.Child {
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

func (self *Wallet) GetAccountAddress(accountParam string) (KeySerializedData) {
	for _, account := range self.MasterAccount.Child {
		if account.Name == accountParam {
			key := KeySerializedData{
				PaymentAddress: account.Key.Base58CheckSerialize(PaymentAddressType),
				ReadonlyKey:    account.Key.Base58CheckSerialize(ReadonlyKeyType),
			}
			return key
		}
	}
	newAccount := self.CreateNewAccount(accountParam)
	key := KeySerializedData{
		PaymentAddress: newAccount.Key.Base58CheckSerialize(PaymentAddressType),
		ReadonlyKey:    newAccount.Key.Base58CheckSerialize(ReadonlyKeyType),
	}
	return key
}

func (self *Wallet) GetAddressesByAccount(accountParam string) ([]KeySerializedData) {
	result := make([]KeySerializedData, 0)
	for _, account := range self.MasterAccount.Child {
		if account.Name == accountParam {
			item := KeySerializedData{
				PaymentAddress: account.Key.Base58CheckSerialize(PaymentAddressType),
				ReadonlyKey:    account.Key.Base58CheckSerialize(ReadonlyKeyType),
			}
			result = append(result, item)
		}
	}
	return result
}

func (self *Wallet) ListAccounts() map[string]Account {
	result := make(map[string]Account)
	for _, account := range self.MasterAccount.Child {
		result[account.Name] = account
	}
	return result
}
