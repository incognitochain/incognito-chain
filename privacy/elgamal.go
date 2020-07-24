package privacy

// elGamalPublicKeyOld represents to public key in ElGamal encryption
// H = G^X, X is private key
type elGamalPublicKey struct {
	h *Point
}

// elGamalPrivateKeyOld represents to private key in ElGamal encryption
type elGamalPrivateKey struct {
	x *Scalar
}

// elGamalCipherTextOld represents to ciphertext in ElGamal encryption
// in which C1 = G^k and C2 = H^k * message
// k is a random number (32 bytes), message is an elliptic point
type elGamalCipherText struct {
	c1, c2 *Point
}

func (ciphertext *elGamalCipherText) set(c1, c2 *Point) {
	ciphertext.c1 = c1
	ciphertext.c2 = c2
}

func (pub *elGamalPublicKey) set(H *Point) {
	pub.h = H
}

func (pub elGamalPublicKey) GetH() *Point {
	return pub.h
}

func (priv *elGamalPrivateKey) set(x *Scalar) {
	priv.x = x
}

func (priv elGamalPrivateKey) GetX() *Scalar {
	return priv.x
}

// Bytes converts ciphertext to 66-byte array
func (ciphertext elGamalCipherText) Bytes() []byte {
	if ciphertext.c1.IsIdentity() {
		return []byte{}
	}
	b1 := ciphertext.c1.ToBytes()
	b2 := ciphertext.c2.ToBytes()
	res := append(b1[:], b2[:]...)
	return res
}

// SetBytes reverts 66-byte array to ciphertext
func (ciphertext *elGamalCipherText) SetBytes(bytes []byte) error {
	if len(bytes) == 0 {
		return NewPrivacyErr(InvalidInputToSetBytesErr, nil)
	}

	if ciphertext == nil {
		ciphertext = new(elGamalCipherText)
	}

	var err error

	var tmp [Ed25519KeySize]byte
	copy(tmp[:], bytes[:Ed25519KeySize])
	ciphertext.c1, err = new(Point).FromBytes(tmp)
	if err != nil {
		return err
	}
	copy(tmp[:], bytes[Ed25519KeySize:])
	ciphertext.c2, err = new(Point).FromBytes(tmp)
	if err != nil {
		return err
	}

	return nil
}

// encrypt encrypts plaintext (is an elliptic point) using public key ElGamal
// returns ElGamal ciphertext
func (pub elGamalPublicKey) encrypt(plaintext *Point) *elGamalCipherText {
	// r random, S:= h^r where h = g^x
	r := RandomScalar()
	S := new(Point).ScalarMult(pub.h, r)

	//return ciphertext (c1, c2) = (g^r, m.s=m.h^r)
	ciphertext := new(elGamalCipherText)
	ciphertext.c1 = new(Point).ScalarMultBase(r)
	ciphertext.c2 = new(Point).Add(plaintext, S)

	return ciphertext
}

// decrypt receives a ciphertext and
// decrypts it using private key ElGamal
// and returns plain text in elliptic point
func (priv elGamalPrivateKey) decrypt(ciphertext *elGamalCipherText) (*Point, error) {
	S := new(Point).ScalarMult(ciphertext.c1, priv.x)
	plaintext := new(Point).Sub(ciphertext.c2, S)
	return plaintext, nil
}
