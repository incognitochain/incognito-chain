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

	ElGamalCiphertextSize = 66 // bytes
	SchnMultiSigSize      = 65 // bytes

)
