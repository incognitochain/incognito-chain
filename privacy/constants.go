package privacy

const (
	pointCompressed byte = 0x2
)

const (
	CompressedPointSize = 33

	CMRingSize    = 8 // 2^3
	CMRingSizeExp = 3

	// it is used for both privacy and no privacy
	SigPubKeySize    = 33
	SigNoPrivacySize = 64
	SigPrivacySize   = 96

	BigIntSize          = 32 // bytes
	Uint64Size          = 8  // bytes
	PrivateKeySize      = 32 // bytes
	PublicKeySize       = 33 // bytes
	TransmissionKeySize = 33 //bytes
	ReceivingKeySize    = 32 // bytes
	PaymentAddressSize  = 66 // bytes

	ElGamalCiphertextSize = 66 // bytes
	SchnMultiSigSize      = 65 // bytes

)
