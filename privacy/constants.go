package privacy

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
)

const (
	MaxSizeInfoCoin = 255 // byte
)
