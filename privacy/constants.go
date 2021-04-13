package privacy

import "math/big"

const (
	pointCompressed       byte = 0x2
	elGamalCiphertextSize      = 64 // bytes
	schnMultiSigSize           = 65 // bytes
)

const (
	Ed25519KeySize        = 32
	AESKeySize            = 32
	CommitmentRingSize    = 8
	CommitmentRingSizeExp = 3
	CStringBulletProof    = "bulletproof"
	CStringBurnAddress    = "burningaddress"
	FixedRandomnessString = "fixedrandomness"
)

const (
	MaxSizeInfoCoin = 255 // byte
)

var LInt = new(big.Int).SetBytes([]byte{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x14, 0xde, 0xf9, 0xde, 0xa2, 0xf7, 0x9c, 0xd6, 0x58, 0x12, 0x63, 0x1a, 0x5c, 0xf5, 0xd3, 0xed})

// FixedRandomnessShardID is fixed randomness for shardID commitment from param.BCHeightBreakPointFixRandShardCM
// is result from HashToScalar([]byte(FixedRandomnessString))
var FixedRandomnessShardID = new(Scalar).FromBytesS([]byte{0x60, 0xa2, 0xab, 0x35, 0x26, 0x9, 0x97, 0x7c, 0x6b, 0xe1, 0xba, 0xec, 0xbf, 0x64, 0x27, 0x2, 0x6a, 0x9c, 0xe8, 0x10, 0x9e, 0x93, 0x4a, 0x0, 0x47, 0x83, 0x15, 0x48, 0x63, 0xeb, 0xda, 0x6})

// type ValidationEnviroment interface {
// 	IsPrivacy() bool
// 	IsConfimed() bool
// 	TxType() string
// 	ShardID() int
// 	ShardHeight() uint64
// 	BeaconHeight() uint64
// 	ConfimedTime() int64
// 	Version() int
// }
