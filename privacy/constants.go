package privacy

const (
	pointCompressed       byte = 0x2
	elGamalCiphertextSize      = 64 // bytes
	schnMultiSigSize           = 65 // bytes
)

const (
	CompressedEllipticPointSize = 33 // EllipticPoint compress size
	Ed25519KeySize = 32
	AESKeySize = 32
	CommitmentRingSize    = 8
	CommitmentRingSizeExp = 3
	CStringBulletProof = "bulletproof"
)
