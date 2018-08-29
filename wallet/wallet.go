package wallet

type Account struct {
	key   Key
	child []Account
}
type Wallet struct {
	seed          []byte
	entropy       []byte
	passpharse    string
	mnemonic      string
	masterAccount Account
}

func (self *Wallet) Init(passphase string, numOfAccount int) {
	mnemonicGen := MnemonicGenerator{}
	self.entropy, _ = mnemonicGen.NewEntropy(128)
	self.mnemonic, _ = mnemonicGen.NewMnemonic(self.entropy)
	self.seed = mnemonicGen.NewSeed(self.mnemonic, passphase)

	masterKey, _ := NewMasterKey(self.seed)
	self.masterAccount = Account{
		key:   *masterKey,
		child: make([]Account, 0),
	}

	if numOfAccount == 0 {
		numOfAccount = 1
	}

	for i := 0; i < numOfAccount; i++ {
		childKey := masterKey.NewChildKey(i)
		account := Account{
			key:   childKey,
			child: make([]Account, 0),
		}
	}

}
