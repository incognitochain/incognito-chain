package wallet

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"golang.org/x/crypto/pbkdf2"
	"math/big"
	"strings"
)

var (
	// Some bitwise operands for working with big.Ints
	last11BitsMask          = big.NewInt(2047)
	rightShift11BitsDivider = big.NewInt(2048)
	bigOne                  = big.NewInt(1)
	bigTwo                  = big.NewInt(2)

	// used to isolate the checksum bits from the Entropy+checksum byte array
	wordLengthChecksumMasksMapping = map[int]*big.Int{
		12: big.NewInt(15),
		15: big.NewInt(31),
		18: big.NewInt(63),
		21: big.NewInt(127),
		24: big.NewInt(255),
	}
	// used to use only the desired x of 8 available checksum bits.
	// 256 bit (word length 24) requires all 8 bits of the checksum,
	// and thus no shifting is needed for it (we would get a divByZero crash if we did)
	wordLengthChecksumShiftMapping = map[int]*big.Int{
		12: big.NewInt(16),
		15: big.NewInt(8),
		18: big.NewInt(4),
		21: big.NewInt(2),
	}

	// wordList is the set of words to use
	wordList []string

	// wordMap is a reverse lookup map for wordList
	wordMap map[string]int
)

var (
	// ErrInvalidMnemonic is returned when trying to use a malformed Mnemonic.
	ErrInvalidMnemonic = errors.New("invalid menomic")

	// ErrEntropyLengthInvalid is returned when trying to use an Entropy set with
	// an invalid size.
	ErrEntropyLengthInvalid = errors.New("entropy length must be [128, 256] and a multiple of 32")

	// ErrValidatedSeedLengthMismatch is returned when a validated Seed is not the
	// same size as the given Seed. This should never happen is present only as a
	// sanity assertion.
	ErrValidatedSeedLengthMismatch = errors.New("seed length does not match validated Seed length")

	// ErrChecksumIncorrect is returned when Entropy has the incorrect checksum.
	ErrChecksumIncorrect = errors.New("checksum incorrect")
)

func init() {
	list := NewWordList("english")
	wordList = list
	wordMap = map[string]int{}
	for i, v := range wordList {
		wordMap[v] = i
	}
}

type MnemonicGenerator struct {
}

// NewEntropy will create random Entropy bytes
// so long as the requested size bitSize is an appropriate size.
// bitSize has to be a multiple 32 and be within the inclusive range of {128, 256}
func (mnemonicGen *MnemonicGenerator) NewEntropy(bitSize int) ([]byte, error) {
	err := validateEntropyBitSize(bitSize)
	if err != nil {
		return nil, err
	}

	// create bytes array for Entropy from bitSize
	entropy := make([]byte, bitSize/8)
	// random bytes array
	_, err = rand.Read(entropy)
	return entropy, err
}

// NewMnemonic will return a string consisting of the Mnemonic words for
// the given Entropy.
// If the provide Entropy is invalid, an error will be returned.
func (mnemonicGen *MnemonicGenerator) NewMnemonic(entropy []byte) (string, error) {
	// Compute some lengths for convenience
	entropyBitLength := len(entropy) * 8
	checksumBitLength := entropyBitLength / 32
	sentenceLength := (entropyBitLength + checksumBitLength) / 11

	err := validateEntropyBitSize(entropyBitLength)
	if err != nil {
		return "", err
	}

	// Add checksum to Entropy
	entropy = mnemonicGen.addChecksum(entropy)

	// Break Entropy up into sentenceLength chunks of 11 bits
	// For each word AND mask the rightmost 11 bits and find the word at that index
	// Then bitshift Entropy 11 bits right and repeat
	// Add to the last empty slot so we can work with LSBs instead of MSB

	// Entropy as an int so we can bitmask without worrying about bytes slices
	entropyInt := new(big.Int).SetBytes(entropy)

	// Slice to hold words in
	words := make([]string, sentenceLength)

	// Throw away big int for AND masking
	word := big.NewInt(0)

	for i := sentenceLength - 1; i >= 0; i-- {
		// Get 11 right most bits and bitshift 11 to the right for next time
		word.And(entropyInt, last11BitsMask)
		entropyInt.Div(entropyInt, rightShift11BitsDivider)

		// Get the bytes representing the 11 bits as a 2 byte slice
		wordBytes := mnemonicGen.padByteSlice(word.Bytes(), 2)

		// Convert bytes to an index and add that word to the list
		words[i] = wordList[binary.BigEndian.Uint16(wordBytes)]
	}

	return strings.Join(words, " "), nil
}

// MnemonicToByteArray takes a Mnemonic string and turns it into a byte array
// suitable for creating another Mnemonic.
// An error is returned if the Mnemonic is invalid.
func (mnemonicGen *MnemonicGenerator) MnemonicToByteArray(mnemonic string, raw ...bool) ([]byte, error) {
	var (
		mnemonicSlice    = strings.Split(mnemonic, " ")
		entropyBitSize   = len(mnemonicSlice) * 11
		checksumBitSize  = entropyBitSize % 32
		fullByteSize     = (entropyBitSize-checksumBitSize)/8 + 1
		checksumByteSize = fullByteSize - (fullByteSize % 4)
	)

	// Pre validate that the Mnemonic is well formed and only contains words that
	// are present in the word list
	if !mnemonicGen.IsMnemonicValid(mnemonic) {
		return nil, ErrInvalidMnemonic
	}

	// Convert word indices to a `big.Int` representing the Entropy
	checksummedEntropy := big.NewInt(0)
	modulo := big.NewInt(2048)
	for _, v := range mnemonicSlice {
		index := big.NewInt(int64(wordMap[v]))
		checksummedEntropy.Mul(checksummedEntropy, modulo)
		checksummedEntropy.Add(checksummedEntropy, index)
	}

	// Calculate the unchecksummed Entropy so we can validate that the checksum is
	// correct
	checksumModulo := big.NewInt(0).Exp(bigTwo, big.NewInt(int64(checksumBitSize)), nil)
	rawEntropy := big.NewInt(0).Div(checksummedEntropy, checksumModulo)

	// Convert `big.Int`s to byte padded byte slices
	rawEntropyBytes := mnemonicGen.padByteSlice(rawEntropy.Bytes(), checksumByteSize)
	checksummedEntropyBytes := mnemonicGen.padByteSlice(checksummedEntropy.Bytes(), fullByteSize)

	// ValidateTransaction that the checksum is correct
	newChecksummedEntropyBytes := mnemonicGen.padByteSlice(mnemonicGen.addChecksum(rawEntropyBytes), fullByteSize)
	if !mnemonicGen.compareByteSlices(checksummedEntropyBytes, newChecksummedEntropyBytes) {
		return nil, ErrChecksumIncorrect
	}

	if raw != nil && raw[0] {
		return rawEntropyBytes, nil
	}
	return checksummedEntropyBytes, nil
}

// NewSeed creates a hashed Seed output given a provided string and password.
// No checking is performed to validate that the string provided is a valid Mnemonic.
func (mnemonicGen *MnemonicGenerator) NewSeed(mnemonic string, password string) []byte {
	return pbkdf2.Key([]byte(mnemonic), []byte("Mnemonic"+password), 2048, SeedKeyLen, sha512.New)
}

// IsMnemonicValid attempts to verify that the provided Mnemonic is valid.
// Validity is determined by both the number of words being appropriate,
// and that all the words in the Mnemonic are present in the word list.
func (mnemonicGen *MnemonicGenerator) IsMnemonicValid(mnemonic string) bool {
	// Create a list of all the words in the Mnemonic sentence
	words := strings.Fields(mnemonic)

	// Get word count
	wordCount := len(words)

	// The number of words should be 12, 15, 18, 21 or 24
	if wordCount%3 != 0 || wordCount < 12 || wordCount > 24 {
		return false
	}

	// Check if all words belong in the wordlist
	for _, word := range words {
		if _, ok := wordMap[word]; !ok {
			return false
		}
	}

	return true
}

// Appends to data the first (len(data) / 32)bits of the result of sha256(data)
// Currently only supports data up to 32 bytes
func (mnemonicGen *MnemonicGenerator) addChecksum(data []byte) []byte {
	// Get first byte of sha256
	hash := mnemonicGen.computeChecksum(data)
	firstChecksumByte := hash[0]

	// len() is in bytes so we divide by 4
	checksumBitLength := uint(len(data) / 4)

	// For each bit of check sum we want we shift the data one the left
	// and then set the (new) right most bit equal to checksum bit at that index
	// staring from the left
	dataBigInt := new(big.Int).SetBytes(data)
	for i := uint(0); i < checksumBitLength; i++ {
		// Bitshift 1 left
		dataBigInt.Mul(dataBigInt, bigTwo)

		// Set rightmost bit if leftmost checksum bit is set
		if uint8(firstChecksumByte&(1<<(7-i))) > 0 {
			dataBigInt.Or(dataBigInt, bigOne)
		}
	}

	return dataBigInt.Bytes()
}

// computeChecksum returns hashing of data using SHA256
func (mnemonicGen *MnemonicGenerator) computeChecksum(data []byte) []byte {
	hasher := sha256.New()
	hasher.Write(data)
	return hasher.Sum(nil)
}

// validateEntropyBitSize ensures that Entropy is the correct size for being a
// Mnemonic.
func validateEntropyBitSize(bitSize int) error {
	if (bitSize%32) != 0 || bitSize < 128 || bitSize > 256 {
		return ErrEntropyLengthInvalid
	}
	return nil
}

// padByteSlice returns a byte slice of the given size with contents of the
// given slice left padded and any empty spaces filled with 0's.
func (mnemonic *MnemonicGenerator) padByteSlice(slice []byte, length int) []byte {
	offset := length - len(slice)
	if offset <= 0 {
		return slice
	}
	newSlice := make([]byte, length)
	copy(newSlice[offset:], slice)
	return newSlice
}

// compareByteSlices returns true of the byte slices have equal contents and
// returns false otherwise.
func (mnemonicGen *MnemonicGenerator) compareByteSlices(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// splitMnemonicWords splits mnemonic string into list of words in that mnemonic string
func (mnemonicGen *MnemonicGenerator) splitMnemonicWords(mnemonic string) ([]string, bool) {
	// Create a list of all the words in the Mnemonic sentence
	words := strings.Fields(mnemonic)

	//Get num of words
	numOfWords := len(words)

	// The number of words should be 12, 15, 18, 21 or 24
	if numOfWords%3 != 0 || numOfWords < 12 || numOfWords > 24 {
		return nil, false
	}
	return words, true
}
