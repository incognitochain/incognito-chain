package privacy

const (
	pointBytesLenCompressed      = 33
	pointCompressed         byte = 0x2
	SK                           = byte(0x00)
	VALUE                        = byte(0x01)
	SND                          = byte(0x02)
	RAND                         = byte(0x03)
	FULL                         = byte(0x04)

	CMRingSize    = 8 // 2^3
	CMRingSizeExp = 3

	ComInputOpeningsProofSize       = 0
	ComOutputOpeningsProofSize      = 0
	OneOfManyProofSize              = 0
	EqualityOfCommittedValProofSize = 0
	ComMultiRangeProofSize          = 0
	ComZeroProofSize                = 0
	ComZeroOneProofSize             = 0
	EllipticPointCompressSize       = 33
	CommitmentSize                  = 0
	BigIntSize                      = 32
)
