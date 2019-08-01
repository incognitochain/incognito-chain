package blsmultisig

import "math/big"

/** Acronym
 * pn     : point
 * cmpr   : compress
 * sz     : size
 * decmpr : de-compress
 * cmpt	  : compute
 * <x>2<y>: <x> to <y>
 * <x>4<y>: <x> for <y>
 */

const (
	// CCmprPnSz Compress point size
	CCmprPnSz = 32
	// CBigIntSz Big Int Byte array size
	CBigIntSz = 32
	// CMaskByte 0b10000000
	CMaskByte = 0x80
	// CNotMaskB 0b01111111
	CNotMaskB = 0x7F
	// CPKSz Public key size
	CPKSz = 32
	// CSKSz Secret key size
	CSKSz = 32
)

const (
	CErr = " Details error: "
	// CErrInLn Error input length
	CErrInLn = "Wrong input length"
)

var (
	// pAdd1Div4 = (p + 1)/4
	pAdd1Div4, _ = new(big.Int).SetString("c19139cb84c680a6e14116da060561765e05aa45a1c72a34f082305b61f3f52", 16) //
)

// PublicKey is bytes of PublicKey point compressed
type PublicKey []byte

// SecretKey is bytes of SecretKey big Int in Fp
type SecretKey []byte
