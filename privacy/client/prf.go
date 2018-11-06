package client

import (
	"github.com/ninjadotorg/constant/privacy/client/crypto/sha256"
	"golang.org/x/crypto/blake2b"
)

func PRF_addr_x(x []byte, t byte) []byte {
	var y [32]byte
	y[0] = t
	return PRF(true, true, false, false, x, y[:])
}

func PRF_nf(x, y []byte) []byte {
	return PRF(true, true, true, false, x, y)
}

func PRF_rho(index uint64, phi, hsig []byte) []byte {
	bit := false
	if index > 0 {
		bit = true
	}
	return PRF(false, bit, true, false, phi, hsig)
}

func PRF_pk(index uint64, ask, hsig []byte) []byte {
	bit := false
	if index > 0 {
		bit = true
	}
	return PRF(false, bit, false, false, ask, hsig)
}

func PRF(a, b, c, d bool, x, y []byte) []byte {
	// TODO: check for x, y length
	xc := make([]byte, len(x))
	copy(xc, x)

	var h byte
	if a {
		h |= 1 << 7
	}
	if b {
		h |= 1 << 6
	}
	if c {
		h |= 1 << 5
	}
	if d {
		h |= 1 << 4
	}

	xc[0] &= 0x0F
	xc[0] |= h
	z := append(xc, y...)
	// fmt.Printf("PRF z: %x\n", z)

	r := sha256.Sum256NoPad(z)
	// fmt.Printf("PRF r: %x\n", r)
	return r[:]
}

func HSigCRH(seed, nf1, nf2, pubKey []byte) []byte {
	var data []byte
	data = append(seed, nf1...)
	data = append(data, nf2...)
	data = append(data, pubKey...)
	result := blake2b.Sum256(data)
	return result[:]
}
