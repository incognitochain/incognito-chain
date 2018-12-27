package privacy

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)


type AES struct{
	key []byte
}

//func (self *AES) GenKey(keyLength byte) error{
//	if keyLength != 16 || keyLength != 24 || keyLength != 32{
//		return NewPrivacyErr(UnexpectedErr, errors.New("privacy/aes: invalid key size " + strconv.Itoa(int(keyLength))))
//	}
//	self.key = RandBytes(int(keyLength))
//	return nil
//}


func (self *AES) SetKey(key []byte) {
	self.key = key
}

func (self *AES) Encrypt(plaintext []byte) ([]byte, error){
	block, err := aes.NewCipher(self.key)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))

	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	return ciphertext, nil
}


func (self *AES) Decrypt(ciphertext []byte) ([]byte, error) {
	plaintext := make([]byte, len(ciphertext[aes.BlockSize:]))

	block, err := aes.NewCipher(self.key)
	if err != nil {
		return nil, err
	}

	iv := ciphertext[:aes.BlockSize]
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(plaintext, ciphertext[aes.BlockSize:])

	return plaintext, nil
}

