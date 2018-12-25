package privacy

const (
	CompressedPointSize      = 33
	PointCompressed     byte = 0x2

	SK            = byte(0x00)
	VALUE         = byte(0x01)
	SND           = byte(0x02)
	SHARDID				= byte(0x03)
	RAND          = byte(0x04)
	FULL          = byte(0x05)

	CMRingSize      = 8 // 2^3
	CMRingSizeExp   = 3

	MaxExp = 64

	// size of zero knowledge proof corresponding one input
	//ComInputOpeningsProofSize       = 198
	OneOfManyProofSize              = 781	// corresponding to CMRingSize = 4: 521
	EqualityOfCommittedValProofSize = 230
	ProductCommitmentProofSize      = 197

	// size of zero knowledge proof corresponding one output
	ComOutputOpeningsProofSize   = 198
	SumOutRangeProofSize             = 99
	ComZeroProofSize                = 99


	InputCoinsPrivacySize = 33  // serial number
	OutputCoinsPrivacySize = 239 // PK + coin commitment + SND + Encrypted (138 bytes) + 2 bytes saving size



	// it is used for both privacy and no privacy
	SigPubKeySize = 33
	SigSize = 64

	SerialNumberSize   = 33 // bytes
	CoinCommitmentSize = 33 // bytes
	RandomSize         = 32 // bytes
	ValueSize          = 8  // bytes
	SNDerivatorSize    = 32 // bytes

	SpendingKeySize = 32

	BigIntSize     									= 32 // bytes
	Uint64Size		= 8 // bytes


)
