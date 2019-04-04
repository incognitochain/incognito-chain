package privacy

const (
	CompressedPointSize      = 33
	PointCompressed     byte = 0x2

	CMRingSize    = 8 // 2^3
	CMRingSizeExp = 3

	MaxExp = 64

	// size of zero knowledge proof corresponding one input
	OneOfManyProofSize = 716

	SNPrivacyProofSize   = 326
	SNNoPrivacyProofSize = 196

	InputCoinsPrivacySize  = 40  // serial number + 7 for flag
	OutputCoinsPrivacySize = 223 // PublicKey + coin commitment + SND + Ciphertext (122 bytes) + 9 bytes flag

	InputCoinsNoPrivacySize  = 178 // PublicKey + coin commitment + SND + Serial number + Randomness + Value + 7 flag
	OutputCoinsNoPrivacySize = 147 // PublicKey + coin commitment + SND + Randomness + Value + 9 flag

	// it is used for both privacy and no privacy
	SigPubKeySize    = 33
	SigNoPrivacySize = 64
	SigPrivacySize   = 96

	PrivateKeySize = 32

	BigIntSize    = 32 // bytes
	Uint64Size    = 8  // bytes
	PublicKeySize = 33
)
