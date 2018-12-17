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

	InfoLength 										 	 = 512

	// size of zero knowledge proof corresponding one input
	ComInputOpeningsProofSize = 198
	OneOfManyProof    				= 781
	EqualityOfCommittedValProof = 230
	ProductCommitmentProof      = 197

	// size of zero knowledge proof corresponding one output
	ComOutputOpeningsProof = 198
	ComOutputMultiRangeProof = 1174
	SumOutRangeProof = 99
	ComZeroProof = 99


	InputCoinsPrivacy = 33  // serial number
	OutputCoinsPrivacy = 237 // PK + coin commitment + SND + Encrypted (138 bytes)

	InputCoinsNoPrivacy = 171
	OutputCoinsNoPrivacy = 138 // except serial number


	BigIntSize     									= 32

	// it is used for both privacy and no privacy
	SigPubKeySize = 33
	SigSize = 64




)
