package client

import "github.com/ninjadotorg/cash-prototype/privacy/client/crypto/sha256"

const CMPreImageLength = 105 // bytes

type Note struct {
	Value          uint64
	Apk            SpendingAddress
	Rho, R, Nf, Cm []byte
}

func GetCommitment(note *Note) []byte {
	var data [CMPreImageLength]byte
	data[0] = 0xB0
	copy(data[1:], note.Apk[:])
	for i := 0; i < 8; i++ {
		data[i+33] = byte(note.Value >> uint(i*8))
	}
	copy(data[41:], note.Rho)
	copy(data[73:], note.R)

	result := sha256.Sum256(data[:])
	return result[:]
}

func GetNullifier(ask SpendingKey, Rho [32]byte) []byte {
	return PRF_nf(ask[:], Rho[:])
}