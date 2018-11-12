package client

import (
	"crypto/aes"
	"crypto/cipher"
	cryptorand "crypto/rand"
	"encoding/json"
	"fmt"

	"github.com/ninjadotorg/constant/privacy/client/crypto/sha256"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/curve25519"

	"bytes"
	"encoding/base64"
	"errors"
	"io"
	mathrand "math/rand"
	"strings"
)

const CMPreImageLength = 105 // bytes

type Note struct {
	Value                uint64
	Apk                  []byte
	Rho, R, Nf, Cm, Memo []byte
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

func ParseNoteToJson(note *Note) []byte {
	var tmpnote struct {
		Value        uint64
		Rho, R, Memo []byte
	}
	tmpnote.Value = note.Value
	tmpnote.Rho = note.Rho
	tmpnote.R = note.R
	tmpnote.Memo = note.Memo

	noteJson, err := json.Marshal(&tmpnote)
	if err != nil {
		return []byte{}
	}
	// fmt.Printf("%s", noteJson)
	return noteJson
}

func ParseJsonToNote(jsonnote []byte) (*Note, error) {
	var note Note
	err := json.Unmarshal(jsonnote, &note)
	if err != nil {
		return nil, err
	}
	// fmt.Println(note)
	return &note, nil
}

func EncryptNote(note [2]*Note, pkenc [2]TransmissionKey,
	esk EphemeralPrivKey, epk EphemeralPubKey, hSig []byte) [][]byte {

	noteJsons := [][]byte{ParseNoteToJson(note[0]), ParseNoteToJson(note[1])}

	var sk [32]byte
	copy(sk[:], esk[:])

	var epk1 [32]byte
	copy(epk1[:], epk[:])

	var pk [2][32]byte
	var sharedSecret [2][32]byte

	var symKey [2][]byte
	ciphernotes := make([][]byte, 2)

	// fmt.Printf("ciphernote[0] = %v", ciphernotes[0][:])

	//Create symmetric key 256-bit
	for i, _ := range pkenc {
		copy(pk[i][:], pkenc[i][:])
		sharedSecret[i] = KeyAgree(&pk[i], &sk)
		symKey[i] = KDF(sharedSecret[i], epk, pk[i], hSig)
		ciphernotes[i] = Encrypt(symKey[i], noteJsons[i][:])
		// fmt.Printf("\nShare secret key: %v\n", sharedSecret[i])
	}
	// fmt.Printf("\nCiphernote 1: %+v\n", ciphernotes[0])
	// fmt.Printf("\nCiphernote 2: %+v\n", ciphernotes[1])
	return ciphernotes
}

func DecryptNote(ciphernote []byte, skenc ReceivingKey,
	pkenc TransmissionKey, epk EphemeralPubKey, hSig []byte) (*Note, error) {

	var epk1 [32]byte
	copy(epk1[:], epk[:])

	var sharedSecret [32]byte
	var symKey []byte
	var plaintext []byte

	var sk, pk [32]byte

	copy(sk[:], skenc[:])
	copy(pk[:], pkenc[:])
	sharedSecret = KeyAgree(&epk1, &sk)
	symKey = KDF(sharedSecret, epk, pk, hSig)
	plaintext, err := Decrypt(symKey, ciphernote)
	if err != nil {
		return nil, err
	}

	note, err := ParseJsonToNote(plaintext)
	return note, err
}

// Create share secret key
func KeyAgree(pk *[32]byte, sk *[32]byte) [32]byte {
	var result [32]byte
	curve25519.ScalarMult(&result, sk, pk)
	return result
}

// Create symmetric key 256-bit
func KDF(sharedSecret [32]byte, epk [32]byte, pkenc [32]byte, hSig []byte) []byte {
	var data []byte

	//data = append(hSig[:], sharedSecret[:]...)
	data = append(data[:], sharedSecret[:]...)
	data = append(data[:], epk[:]...)
	data = append(data[:], pkenc[:]...)
	data = append(data[:], hSig[:]...)
	result := blake2b.Sum256(data)
	return result[:]
}

// AES
func addBase64Padding(value string) string {
	m := len(value) % 4
	if m != 0 {
		value += strings.Repeat("=", 4-m)
	}

	return value
}

func removeBase64Padding(value string) string {
	return strings.Replace(value, "=", "", -1)
}

func Pad(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func Unpad(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])

	if unpadding > length {
		return nil, errors.New("unpad error. This could happen when incorrect encryption key is used")
	}

	return src[:(length - unpadding)], nil
}

func Encrypt(key []byte, text []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	msg := Pad([]byte(text))
	ciphertext := make([]byte, aes.BlockSize+len(msg))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(cryptorand.Reader, iv); err != nil {
		panic(err)
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(msg))
	finalMsg := removeBase64Padding(base64.URLEncoding.EncodeToString(ciphertext))

	return []byte(finalMsg)
}

func Decrypt(key []byte, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	decodedMsg, err := base64.URLEncoding.DecodeString(addBase64Padding(string(text[:])))
	if err != nil {
		return nil, err
	}

	if (len(decodedMsg) % aes.BlockSize) != 0 {
		panic(errors.New("blocksize must be multipe of decoded message length"))
	}

	iv := decodedMsg[:aes.BlockSize]
	msg := decodedMsg[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(msg, msg)

	unpadMsg, err := Unpad(msg)
	if err != nil {
		return nil, err
	}

	return unpadMsg, nil
}

func Uint64() uint64 {
	return uint64(mathrand.Uint32())<<32 + uint64(mathrand.Uint32())
}

func GenNote() *Note {
	var note Note

	note.Value = Uint64()
	note.Rho = RandBits(256)
	note.R = RandBits(256)
	note.Memo = []byte{}

	return &note
}

func TestEncrypt() {
	var hSig [32]byte
	copy(hSig[:], []byte("the-key-has-to-be-32-bytes-long!"))
	tmp := RandBits(256)
	copy(hSig[:], tmp[:])
	//Generate note
	note_temp := GenNote()
	notes := [2]*Note{note_temp, note_temp}

	fmt.Printf("\nPlain note: %+v\n", notes)

	// Generate key pair
	ask := RandSpendingKey()
	skenc := GenReceivingKey(ask)
	pkenc := GenTransmissionKey(skenc)

	pkencs := [2]TransmissionKey{pkenc, pkenc}
	// skencs := [2]Rk{skenc, skenc}

	//Gen ephemeral key
	epk, esk := GenEphemeralKey()

	ciphernotes := EncryptNote(notes, pkencs, esk, epk, hSig[:])
	fmt.Printf("\nCiphernotes: %s\n", ciphernotes)

	fmt.Printf("\nReceiving key: %+v\n", skenc)
	fmt.Printf("\nTransmission key: %+v\n", pkenc)

	decrypted_note0, _ := DecryptNote(ciphernotes[0], skenc, pkenc, epk, hSig[:])
	decrypted_note1, _ := DecryptNote(ciphernotes[1], skenc, pkenc, epk, hSig[:])
	fmt.Printf("\nPlaintext: %+v\n", decrypted_note0)
	fmt.Printf("\nPlaintext: %+v\n", decrypted_note1)

}

func TestEncrypt1() {
	text := []byte("My name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is AstaxieMy name is Astaxie")
	key := []byte("the-key-has-to-be-32-bytes-long!")

	ciphertext := Encrypt(key, text)
	fmt.Printf("%s => %x\n", text, ciphertext)

	plaintext, _ := Decrypt(key, ciphertext)
	fmt.Printf("%x => %s\n", ciphertext, plaintext)
}
