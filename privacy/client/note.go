package client

import ("github.com/ninjadotorg/cash-prototype/privacy/client/crypto/sha256"
		"math/rand"
		"fmt"
		"golang.org/x/crypto/openpgp/elgamal"
		"math/big"
		cryptorand "crypto/rand"
		"encoding/binary"
		"unsafe"
		// "golang.org/x/crypto/curve25519"
)

const CMPreImageLength = 105 // bytes

type Note struct {
	Value          uint64
	Apk            SpendingAddress
	Rho, R, Nf, Cm []byte
}

type NotePlain struct {
	Value			uint64
	Rho, R, Memo	[]byte
}

type Ciphertext struct {
	c1 *big.Int
	c2 *big.Int
}

var generator *big.Int = FromHexToBigInt(generatorHex)
var prime *big.Int = FromHexToBigInt(primeHex)

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


/* Uint64: generate random uint64 
 * Input: none
 * Output: uint64 
 */
func Uint64() uint64 {
    return uint64(rand.Uint32())<<32 + uint64(rand.Uint32())
}

/* GenNote: generate random note to test encryption note
 * Input: none
 * Output: *Note 
 */
func GenNote() *NotePlain{
	var note NotePlain

	note.Value = Uint64()
	note.Rho = RandBits(256)
	note.R = RandBits(256)
	note.Memo = []byte{}

	return &note
}

// func ParseNoteToJson(note *NotePlain) []byte {
// 	noteJson, err := json.Marshal(note)
// 	if err != nil {
// 		return []byte{}
// 	}
// 	fmt.Printf("%s", noteJson)
// 	return noteJson
// }

// func ParseJsonToNote (str []byte) note *Note{

// }


/* FromHexToBigInt: parse hex number to big int 
 * Input: hex
 * Output: *big.Int 
 */
func FromHexToBigInt(hex string) *big.Int {
	n, ok := new(big.Int).SetString(hex, 16)
	if !ok {
		panic("Failed to parse hex number")
	}
	return n
}

/* RandBigInt: generate random big int number with given max value 
 * Input: *big.Int
 * Output: *big.Int 
 */
func RandBigInt(max *big.Int)  *big.Int {
	n, err := cryptorand.Int(cryptorand.Reader, max)
	if err != nil {
		return nil
	}
	return n
}


/* EncryptNote: encrypt note plaintext  
 * Input: note *Note, pub * elgamal.PublicKey
 * Output: [6]Ciphertext
 * Each note's attribute is encrypted and is stored in Ciphertext  
 */


func EncryptNote(note *NotePlain, pkenc * TransmissionKey  ) [4]Ciphertext {
	// Each note's attribute is converted to byte sequence 
	var note_ptxt [4][]byte
	var ciphers [4]Ciphertext

	buf := make([]byte, binary.MaxVarintLen64)
	value := binary.PutVarint(buf, (int64)(note.Value))
	note_ptxt[0] = buf[:value]
	note_ptxt[1] = note.Rho
	note_ptxt[2] = note.R
	note_ptxt[3] = note.Memo

	for i:=0; i<4; i++ {
		c1, c2, err := elgamal.Encrypt(cryptorand.Reader, (*elgamal.PublicKey)(pkenc), note_ptxt[i])
		if err != nil {
			fmt.Printf("\n%s\n", "Can not encrypt message")
			fmt.Println(err)
		} else {
			ciphers[i].c1 = c1
			ciphers[i].c2 = c2
		}
	}
	return ciphers
}

/* DecryptNote: decrypt note ciphertext  
 * Input: ciphers [6]Ciphertext, priv * elgamal.PrivateKey
 * Output: Note
 * Each note's ciphertext is decrypted
 */

func DecryptNote(ciphers [4]Ciphertext, skenc * ReceivingKey) NotePlain{
	var note NotePlain
	var msgs [4][]byte

	for i:=0; i < 4; i++ {
		msg, err := elgamal.Decrypt((*elgamal.PrivateKey)(skenc), ciphers[i].c1, ciphers[i].c2)
		
		if err != nil{
			fmt.Printf("\n%s\n", "Can not decrypt ciphertext")
			panic(err)
		} else {
			msgs[i] = msg
		}
	} 
	
	value, _ := binary.Varint(msgs[0])
	note.Value = (uint64)(value)
	note.Rho = msgs[1]
	note.R = msgs[2]
	note.Memo = msgs[3]
	  
	return note
}


func TestEncryption(){
	//Test setstring
	fmt.Printf("Test set string function\n")
	fmt.Printf("\n From big int: %x \n", generator)
	fmt.Printf("\n From hexa: %s\n", generatorHex)

	//Generate note
	note := GenNote()
	fmt.Printf("\nPlain note: %+v\n", note)

	// Generate key pair
	skenc := GenReceivingKey()
	pkenc := GenTransmissionKey(skenc)

	pksize := unsafe.Sizeof(*pkenc.G) + unsafe.Sizeof(*pkenc.P) + unsafe.Sizeof(*pkenc.Y)

	fmt.Printf("\nReceiving key: %+v\n", skenc)
	fmt.Printf("\nTransmission key: %+v\n", pkenc)

	fmt.Printf("\nReceiving key size: %d \n", unsafe.Sizeof(*skenc.X)+ pksize)
	fmt.Printf("\nTransmission key size: %d \n", pksize)
	
	noteCipher := EncryptNote(note, &pkenc)
	fmt.Printf("\nCiphertext: %+v\n", noteCipher)
	fmt.Printf("\nCiphertext size: %d\n", unsafe.Sizeof(*noteCipher[0].c1) + unsafe.Sizeof(*noteCipher[0].c2) + unsafe.Sizeof(*noteCipher[1].c1) + unsafe.Sizeof(*noteCipher[1].c2) + unsafe.Sizeof(*noteCipher[2].c1) + unsafe.Sizeof(*noteCipher[2].c2) )

	noteDecrypted := DecryptNote(noteCipher, &skenc)
	fmt.Printf("\nDecrypted note: %+v\n", noteDecrypted)
}