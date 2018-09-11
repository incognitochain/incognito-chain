package wallet

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

	Config *WalletConfig
}

type WalletConfig struct {
	DataDir  string
	DataFile string
	DataPath string
}

func (self *Wallet) Init(passPhrase string, numOfAccount uint32) {
	mnemonicGen := MnemonicGenerator{}
	self.Entropy, _ = mnemonicGen.NewEntropy(128)
	self.Mnemonic, _ = mnemonicGen.NewMnemonic(self.Entropy)
	self.Seed = mnemonicGen.NewSeed(self.Mnemonic, passPhrase)
	self.PassPhrase = passPhrase

	masterKey, _ := NewMasterKey(self.Seed)
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

func (self *Wallet) ImportAccount(privateKey string) {
	key, _ := Base58CheckDeserialize(privateKey)
	account := Account{
		Key:        *key,
		Child:      make([]Account, 0),
		IsImported: true,
	}
	self.MasterAccount.Child = append(self.MasterAccount.Child, account)
}

func (self *Wallet) Save(password string) error {
	if password == "" {
		password = self.PassPhrase
	}

	// parse to byte[]
	data, err := json.Marshal(*self)
	if err != nil {
		Logger.log.Error(err)
		return err
	}

	// encrypt
	cipherText, err := AES{}.Encrypt(password, data)
	if err != nil {
		Logger.log.Error(err)
		return err
	}
	// and
	// save file
	err = ioutil.WriteFile(self.Config.DataPath, []byte(cipherText), 0644)
	return err
}

func (self *Wallet) LoadWallet(password string) error {
	// read file and decrypt
	bytesData, err := ioutil.ReadFile(self.Config.DataPath)
	if err != nil {
		Logger.log.Error(err)
		return err
	}
	bufBytes, err := AES{}.Decrypt(password, string(bytesData))
	if err != nil {
		Logger.log.Error(err)
		return err
	}

	// read to struct
	err = json.Unmarshal(bufBytes, &self)
	return err
}

func (self *Wallet) DumpPrivkey(addressP string) (KeySerializedData, error) {
	for _, account := range self.MasterAccount.Child {
		address := account.Key.Base58CheckSerialize(PubKeyType)
		if address == addressP {
			key := KeySerializedData{
				PrivateKey: account.Key.Base58CheckSerialize(PriKeyType),
			}
			return key, nil
		}
	}
	return KeySerializedData{}, nil
}

func (self *Wallet) GetAccountAddress(accountParam string) (KeySerializedData, error) {
	for _, account := range self.MasterAccount.Child {
		if account.Name == accountParam {
			key := KeySerializedData{
				PublicKey:   account.Key.Base58CheckSerialize(PubKeyType),
				ReadonlyKey: account.Key.Base58CheckSerialize(ReadonlyKeyType),
			}
			return key, nil
		}
	}
	newAccount := self.CreateNewAccount(accountParam)
	key := KeySerializedData{
		PublicKey:   newAccount.Key.Base58CheckSerialize(PubKeyType),
		ReadonlyKey: newAccount.Key.Base58CheckSerialize(ReadonlyKeyType),
	}
	return key, nil
}

func (self *Wallet) GetAddressesByAccount(accountParam string) ([]KeySerializedData, error) {
	result := make([]KeySerializedData, 0)
	for _, account := range self.MasterAccount.Child {
		if account.Name == accountParam {
			item := KeySerializedData{
				PublicKey:   account.Key.Base58CheckSerialize(PubKeyType),
				ReadonlyKey: account.Key.Base58CheckSerialize(ReadonlyKeyType),
			}
			result = append(result, item)
		}
	}
	return result, nil
}

func (self *Wallet) ListAccounts() map[string]float64 {
	result := make(map[string]float64)
	for _, account := range self.MasterAccount.Child {
		result[account.Name] = 0.0
	}
	return result
}
