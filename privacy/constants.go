package privacy

const (
	CompressedPointSize      = 33
	PointCompressed     byte = 0x2

	CMRingSize    = 8 // 2^3
	CMRingSizeExp = 3

	MaxExp = 64

	// size of zero knowledge proof corresponding one input
	OneOfManyProofSize = 716 // corresponding to CMRingSize = 4: 521

	SNPrivacyProofSize   = 326
	SNNoPrivacyProofSize = 196

	// size of zero knowledge proof corresponding one output
	SumOutRangeProofSize = 99
	ComZeroProofSize     = 66

	InputCoinsPrivacySize  = 33  // serial number
	OutputCoinsPrivacySize = 239 // vKey + coin commitment + input + Encrypted (138 bytes) + 2 bytes saving size

	// it is used for both privacy and no privacy
	SigPubKeySize = 33
	SigSize       = 64

	SpendingKeySize = 32

	BigIntSize = 32 // bytes
	Uint64Size = 8  // bytes

	EncryptedRandomnessSize = 48 //bytes
	EncryptedSymKeySize     = 66 //bytes

)
