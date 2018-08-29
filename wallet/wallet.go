package wallet

import (
	"io/ioutil"
	"encoding/json"
)

type Account struct {
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

	masterKey, _ := NewMasterKey(self.Seed)
	self.MasterAccount = Account{
		Key:   *masterKey,
		Child: make([]Account, 0),
	}

	if numOfAccount == 0 {
		numOfAccount = 1
	}

	for i := uint32(0); i < numOfAccount; i++ {
		childKey, _ := self.MasterAccount.Key.NewChildKey(i)
		account := Account{
			Key:   *childKey,
			Child: make([]Account, 0),
		}
		self.MasterAccount.Child = append(self.MasterAccount.Child, account)
	}
}

func (self *Wallet) CreateNewAccount() {
	newIndex := uint32(len(self.MasterAccount.Child))
	childKey, _ := self.MasterAccount.Key.NewChildKey(newIndex)
	account := Account{
		Key:   *childKey,
		Child: make([]Account, 0),
	}
	self.MasterAccount.Child = append(self.MasterAccount.Child, account)
}

func (self *Wallet) ExportAccount(childIndex uint32) (string) {
	return self.MasterAccount.Child[childIndex].Key.Base58CheckSerialize(true)
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

func (self *Wallet) ListAccounts() ([]Account) {
	return self.MasterAccount.Child
}

func (self *Wallet) Save(password string) (error) {
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

func (self *Wallet) LoadWallet(password string) (error) {
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
