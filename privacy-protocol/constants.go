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
	ComInputOpeningsProofSize  			 =	226
	EqualityOfCommittedValProofSize  = 230
	ProductCommitmentProofSize 			= 197
	ComOutputOpeningsProofSize 			= 226
	ComZeroProofSize                = 99
	CommitmentSize 									= 0
	BigIntSize     									= 32

)
