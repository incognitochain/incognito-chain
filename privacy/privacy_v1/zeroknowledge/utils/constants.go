package utils

const (
	// size of zero knowledge proof corresponding one input
	maxExp               = 64
	OneOfManyProofSize   = 704
	SnPrivacyProofSize   = 320
	SnNoPrivacyProofSize = 192

	inputCoinsPrivacySize    = 39  // serial number + 7 for flag
	outputCoinsPrivacySize   = 221 // PublicKey + coin commitment + SND + Ciphertext (122 bytes) + 9 bytes flag
	inputCoinsNoPrivacySize  = 175 // PublicKey + coin commitment + SND + Serial number + Randomness + Value + 7 flag
	OutputCoinsNoPrivacySize = 145 // PublicKey + coin commitment + SND + Randomness + Value + 9 flag
)
