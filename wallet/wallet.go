package wallet

type Account struct {
	key        Key
	child      []Account
	isImported bool
}
type Wallet struct {
	seed          []byte
	entropy       []byte
	passpharse    string
	mnemonic      string
	masterAccount Account
}

func (self *Wallet) Init(passPhrase string, numOfAccount uint32) {
	mnemonicGen := MnemonicGenerator{}
	self.entropy, _ = mnemonicGen.NewEntropy(128)
	self.mnemonic, _ = mnemonicGen.NewMnemonic(self.entropy)
	self.seed = mnemonicGen.NewSeed(self.mnemonic, passPhrase)

	masterKey, _ := NewMasterKey(self.seed)
	self.masterAccount = Account{
		key:   *masterKey,
		child: make([]Account, 0),
	}

	if numOfAccount == 0 {
		numOfAccount = 1
	}

	for i := uint32(0); i < numOfAccount; i++ {
		childKey, _ := self.masterAccount.key.NewChildKey(i)
		account := Account{
			key:   *childKey,
			child: make([]Account, 0),
		}
		self.masterAccount.child = append(self.masterAccount.child, account)
	}
}

func (self *Wallet) CreateNewAccount() {
	newIndex := uint32(len(self.masterAccount.child))
	childKey, _ := self.masterAccount.key.NewChildKey(newIndex)
	account := Account{
		key:   *childKey,
		child: make([]Account, 0),
	}
	self.masterAccount.child = append(self.masterAccount.child, account)
}

func (self *Wallet) ExportAccount(childIndex uint32) (string) {
	return self.masterAccount.child[childIndex].key.Base58CheckSerialize(true)
}

func (self *Wallet) ImportAccount(privateKey string) {
	key, _ := Base58CheckDeserialize(privateKey)
	account := Account{
		key:        *key,
		child:      make([]Account, 0),
		isImported: true,
	}
	self.masterAccount.child = append(self.masterAccount.child, account)
}

func (self *Wallet) ListAccounts() ([]Account) {
	return self.masterAccount.child
}
