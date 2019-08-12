package wallet

import (
	"github.com/stretchr/testify/assert"
	"log"
	"strings"
	"testing"
)

/*
	Unit test for NewEntropy function
*/

func TestMnemonicNewEntropy(t *testing.T) {
	data := []int{128, 256}

	mnemonic := MnemonicGenerator{}
	for _, item := range data {
		entropy, err := mnemonic.newEntropy(item)

		assert.Equal(t, nil, err)
		assert.Equal(t, item/8, len(entropy))
	}
}

func TestMnemonicNewEntropyWithInvalidBitSize(t *testing.T) {
	data := []int{33, 257}

	mnemonic := MnemonicGenerator{}
	for _, item := range data {
		_, err := mnemonic.newEntropy(item)

		assert.NotEqual(t, nil, err)
	}
}

/*
	Unit test for NewMnemonic function
*/
func TestMnemonicNewMnemonic(t *testing.T) {
	data := []int{128, 256}
	mnemonic := MnemonicGenerator{}

	for _, item := range data {
		entropy, err := mnemonic.newEntropy(item)
		mnemonic, err := mnemonic.newMnemonic(entropy)

		words := strings.Split(mnemonic, " ")

		log.Print("mnemonic ", mnemonic)

		assert.Equal(t, nil, err)
		assert.Greater(t, len(mnemonic), 0)
		assert.Equal(t, (item+item/32)/11, len(words))
	}
}

func TestMnemonicNewMnemonicWithInvalidEntropy(t *testing.T) {
	data := []int{128, 256}
	mnemonic := MnemonicGenerator{}

	for _, item := range data {
		entropy, _ := mnemonic.newEntropy(item)
		// edit entropy
		entropy = append(entropy, byte(12))
		_, err := mnemonic.newMnemonic(entropy)

		assert.NotEqual(t, nil, err)
	}
}

/*
	Unit test for MnemonicToByteArray function
*/
func TestMnemonicMnemonicToByteArray(t *testing.T) {
	data := []struct {
		len   int
		words string
	}{
		{12, "comic best traffic surround pool want vicious grape october shift scrap stadium"},
		{24, "organ local excess argue economy item surge unfair there knee tongue tree labor divert hockey mountain update differ trial buzz tomato ball farm seven"},
	}
	mnemonic := MnemonicGenerator{}

	for _, item := range data {
		mnemonicBytes, err := mnemonic.mnemonicToByteArray(item.words)

		log.Print("mnemonicBytes: ", mnemonicBytes)

		lenMnemonicBytes := (item.len*11-(item.len*11)%32)/8 + 1

		assert.Equal(t, nil, err)
		assert.Equal(t, lenMnemonicBytes, len(mnemonicBytes))
	}
}

func TestMnemonicMnemonicToByteArrayWithRaw(t *testing.T) {
	data := []struct {
		len   int
		words string
	}{
		{12, "comic best traffic surround pool want vicious grape october shift scrap stadium"},
		{24, "organ local excess argue economy item surge unfair there knee tongue tree labor divert hockey mountain update differ trial buzz tomato ball farm seven"},
	}
	mnemonic := MnemonicGenerator{}

	for _, item := range data {
		mnemonicBytes, err := mnemonic.mnemonicToByteArray(item.words, true)

		log.Print("mnemonicBytes: ", mnemonicBytes)

		lenMnemonicBytes := (item.len*11-(item.len*11)%32)/8 + 1
		lenRawMnemonicBytes := lenMnemonicBytes - lenMnemonicBytes%4

		assert.Equal(t, nil, err)
		assert.Equal(t, lenRawMnemonicBytes, len(mnemonicBytes))
	}
}

func TestMnemonicMnemonicToByteArrayWithInvalidMnemonic(t *testing.T) {
	data := []struct {
		len   int
		words string
	}{
		{1, "abc"},
		{13, "comic best traffic surround pool want vicious grape october shift scrap stadium abc"},
		{25, "organ local excess argue economy item surge unfair there knee tongue tree labor divert hockey mountain update differ trial buzz tomato ball farm seven abc"},
	}
	mnemonic := MnemonicGenerator{}

	for _, item := range data {
		_, err := mnemonic.mnemonicToByteArray(item.words)

		assert.NotEqual(t, nil, err)
	}
}

/*
	Unit test for NewSeed function
*/

func TestMnemonicNewSeed(t *testing.T) {
	data := []struct {
		len   int
		words string
	}{
		{12, "comic best traffic surround pool want vicious grape october shift scrap stadium"},
		{24, "organ local excess argue economy item surge unfair there knee tongue tree labor divert hockey mountain update differ trial buzz tomato ball farm seven"},
	}
	password := "123"
	mnemonic := MnemonicGenerator{}

	for _, item := range data {
		seedBytes := mnemonic.NewSeed(item.words, password)

		log.Print("seedBytes: ", seedBytes)
		assert.Equal(t, seedKeyLen, len(seedBytes))
	}
}

func TestMnemonicNewSeedWithInvalidMnemonic(t *testing.T) {
	data := []struct {
		len   int
		words string
	}{
		{1, "abc"},
		{13, "comic best traffic surround pool want vicious grape october shift scrap stadium abc"},
		{25, "organ local excess argue economy item surge unfair there knee tongue tree labor divert hockey mountain update differ trial buzz tomato ball farm seven abc"},
	}
	password := "123"
	mnemonic := MnemonicGenerator{}

	for _, item := range data {
		seedBytes := mnemonic.NewSeed(item.words, password)
		log.Print("seedBytes: ", seedBytes)
		assert.Equal(t, seedKeyLen, len(seedBytes))
	}
}

/*
	Unit test for IsMnemonicValid function
*/

func TestMnemonicIsMnemonicValid(t *testing.T) {
	data := []struct {
		len     int
		words   string
		isValid bool
	}{
		{12, "comic best traffic surround pool want vicious grape october shift scrap stadium", true},
		{15, "comic best traffic surround pool want vicious grape october shift scrap stadium ball farm seven", true},
		{18, "comic best traffic surround pool want vicious grape october shift scrap stadium trial buzz tomato ball farm seven", true},
		{21, "comic best traffic surround pool want vicious grape october shift scrap stadium mountain update differ trial buzz tomato ball farm seven", true},
		{24, "organ local excess argue economy item surge unfair there knee tongue tree labor divert hockey mountain update differ trial buzz tomato ball farm seven", true},

		{11, "comic best traffic surround pool want vicious grape october shift scrap", false},
		{16, "comic best traffic surround pool want vicious grape october shift scrap stadium ball farm seven buzz", false},
		{0, "", false},
	}
	mnemonic := MnemonicGenerator{}

	for _, item := range data {
		isValid := mnemonic.isMnemonicValid(item.words)
		assert.Equal(t, item.isValid, isValid)
	}
}
