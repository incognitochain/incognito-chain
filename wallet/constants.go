package wallet

const (
	PriKeyType         = byte(0x0) // Serialize wallet account key into string with only PRIVATE KEY of account keyset
	PaymentAddressType = byte(0x1) // Serialize wallet account key into string with only PAYMENT ADDRESS of account keyset
	ReadonlyKeyType    = byte(0x2) // Serialize wallet account key into string with only READONLY KEY of account keyset

	SeedKeyLen        = 64 //bytes
	ChildNumberLen = 4 // bytes
	ChainCodeLen = 32 // bytes

	PrivateKeySerializedLen = 108
)
