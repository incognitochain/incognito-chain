package wallet

const (
	seedKeyLen     = 64 // bytes
	childNumberLen = 4  // bytes
	chainCodeLen   = 32 // bytes

	privateKeySerializedLen = 108 // len string

	privKeySerializedBytesLen     = 75 // bytes
	paymentAddrSerializedBytesLen = 73 // bytes
	readOnlyKeySerializedBytesLen = 72 // bytes

	privKeyBase58CheckSerializedBytesLen     = 107 // len string
	paymentAddrBase58CheckSerializedBytesLen = 105 // len string
	readOnlyKeyBase58CheckSerializedBytesLen = 104 // len string
)

const (
	PriKeyType         = byte(0x0) // Serialize wallet account key into string with only PRIVATE KEY of account keyset
	PaymentAddressType = byte(0x1) // Serialize wallet account key into string with only PAYMENT ADDRESS of account keyset
	ReadonlyKeyType    = byte(0x2) // Serialize wallet account key into string with only READONLY KEY of account keyset

)
