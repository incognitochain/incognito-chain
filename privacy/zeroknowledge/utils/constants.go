package utils

const (
	// size of zero knowledge proof corresponding one input
	OneOfManyProofSize   = 716
	SnPrivacyProofSize   = 326
	SnNoPrivacyProofSize = 196

	inputCoinsPrivacySize    = 40  // serial number + 7 for flag
	outputCoinsPrivacySize   = 223 // PublicKey + coin commitment + SND + Ciphertext (122 bytes) + 9 bytes flag
	inputCoinsNoPrivacySize  = 178 // PublicKey + coin commitment + SND + Serial number + Randomness + Value + 7 flag
	outputCoinsNoPrivacySize = 147 // PublicKey + coin commitment + SND + Randomness + Value + 9 flag
)
