package relaying

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/types"
	"strings"
)

// Mainnet - Binance
var validatorMapMainnet, _ = NewFixedValidators()

// address = hexEncode(sha256(pubKey)[:20])
var ValidatorAddresses = []string{
	"1175946A48EAA473868A0A6F52E6C66CCAF472EA",
	"14CFCE69B645F3F88BAF08EA5B77FA521E4480F9",
	"17B42E8F284D3CA0E420262F89CD76C749BB12C9",
	"3CD4AABABDDEB7ABFEA9618732E331077A861D2B",
	"414FB3BBA216AF84C47E07D6EBAA2DCFC3563A2F",
	"71F253E6FEA9EDD4B4753F5483549FE4F0F3A21C",
	"7235EF143D20FC0ABC427615D83014BB02D7C06C",
	"A71E5CD078B8C5C7B1AF88BCE84DD70B0557D93E",
	"A9157B3FA6EB4C1E396B9B746E95327A07DC42E5",
	"B0FBB52FF7EE93CC476DFE6B74FA1FC88584F30D",
	"B7707D9F593C62E85BB9E1A2366D12A97CD5DFF2",
}

// public key on ed25519 curve (base64 encoded)
var ValidatorB64EncodePubKeys = []string{
	"03adih94tMF6ll96MNQYH6u9H5afRtPI6Dta1IRUIdg=",
	"K6ToFUL0N7euH4o13bIzx4mo3CJzQ3fZttY68cpAO2E=",
	"342oxav9s4WVORMIu3HloeCqvcHQzzgxXVDWvpObJgY=",
	"tmGe3KQUNISAAoHWmLcMk16RUq1Xsx2FwF8vefZLOfM=",
	"lEbRSthsjS10eAsIRxEAAaHC4lLu3+pHU+u7/OOiL1I=",
	"A1PGOfgMyAFZRENtqxAyJF1E+RLtwx72aP+fSkXNBZk=",
	"6B03l+BUTDpxjh8F8Pt4IhLiSOeEwahRvofneuDbIw4=",
	"Xj/NowvRnUXEtzaI2jXn2h/OfGhZssHyDtUgLSQUTj4=",
	"sGpZotdb9dAU/OfJmbXnHnqWCHD3JYR9S6MjW66qCO8=",
	"DJEOL+ZQ5OAUBrMxC0iftgqEvD/1xb7jpW1YmLaorzI=",
	"cfLXuOwci5mmU0KbARjNIB95T0CdD+pNZbG2YvKwAGM=",
}

// ValidatorPubKeyBytes are results from base-64 decoding ValidatorB64EncodePubKeys
var ValidatorPubKeyBytes = [][]byte{
	{211, 118, 157, 138, 31, 120, 180, 193, 122, 150, 95, 122, 48, 212, 24, 31, 171, 189, 31, 150, 159, 70, 211, 200, 232, 59, 90, 212, 132, 84, 33, 216},
	{43, 164, 232, 21, 66, 244, 55, 183, 174, 31, 138, 53, 221, 178, 51, 199, 137, 168, 220, 34, 115, 67, 119, 217, 182, 214, 58, 241, 202, 64, 59, 97},
	{223, 141, 168, 197, 171, 253, 179, 133, 149, 57, 19, 8, 187, 113, 229, 161, 224, 170, 189, 193, 208, 207, 56, 49, 93, 80, 214, 190, 147, 155, 38, 6},
	{182, 97, 158, 220, 164, 20, 52, 132, 128, 2, 129, 214, 152, 183, 12, 147, 94, 145, 82, 173, 87, 179, 29, 133, 192, 95, 47, 121, 246, 75, 57, 243},
	{148, 70, 209, 74, 216, 108, 141, 45, 116, 120, 11, 8, 71, 17, 0, 1, 161, 194, 226, 82, 238, 223, 234, 71, 83, 235, 187, 252, 227, 162, 47, 82},
	{3, 83, 198, 57, 248, 12, 200, 1, 89, 68, 67, 109, 171, 16, 50, 36, 93, 68, 249, 18, 237, 195, 30, 246, 104, 255, 159, 74, 69, 205, 5, 153},
	{232, 29, 55, 151, 224, 84, 76, 58, 113, 142, 31, 5, 240, 251, 120, 34, 18, 226, 72, 231, 132, 193, 168, 81, 190, 135, 231, 122, 224, 219, 35, 14},
	{94, 63, 205, 163, 11, 209, 157, 69, 196, 183, 54, 136, 218, 53, 231, 218, 31, 206, 124, 104, 89, 178, 193, 242, 14, 213, 32, 45, 36, 20, 78, 62},
	{176, 106, 89, 162, 215, 91, 245, 208, 20, 252, 231, 201, 153, 181, 231, 30, 122, 150, 8, 112, 247, 37, 132, 125, 75, 163, 35, 91, 174, 170, 8, 239},
	{12, 145, 14, 47, 230, 80, 228, 224, 20, 6, 179, 49, 11, 72, 159, 182, 10, 132, 188, 63, 245, 197, 190, 227, 165, 109, 88, 152, 182, 168, 175, 50},
	{113, 242, 215, 184, 236, 28, 139, 153, 166, 83, 66, 155, 1, 24, 205, 32, 31, 121, 79, 64, 157, 15, 234, 77, 101, 177, 182, 98, 242, 176, 0, 99},
}

var ValidatorVotingPowers = []int64{
	1000000000000,
	1000000000000,
	1000000000000,
	1000000000000,
	1000000000000,
	1000000000000,
	1000000000000,
	1000000000000,
	1000000000000,
	1000000000000,
	1000000000000,
}

// SHA256 returns the SHA256 of the bz.
// todo: need to be moved to common package
func SHA256(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

// DecodePublicKeyValidator decodes encoded public key to public key in bytes array
func DecodePublicKeyValidator() error {
	ValidatorPubKeyBytes = make([][]byte, len(ValidatorB64EncodePubKeys))
	for i, item := range ValidatorB64EncodePubKeys {
		bytes, err := base64.StdEncoding.DecodeString(item)
		if err != nil {
			return err
		}

		// check public key bytes to address
		pubKeyHash := SHA256(bytes)
		addTmpStr := strings.ToUpper(hex.EncodeToString(pubKeyHash[0:20]))

		if addTmpStr == ValidatorAddresses[i] {
			ValidatorPubKeyBytes[i] = bytes
		} else{
			fmt.Printf("Public key is wrong %v\n", i)
		}
	}

	fmt.Printf("ValidatorPubKeyBytes %#v\n", ValidatorPubKeyBytes)
	return nil
}

func NewFixedValidators() (map[string]*types.Validator, error) {
	if len(ValidatorAddresses) != len(ValidatorPubKeyBytes) || len(ValidatorAddresses) != len(ValidatorVotingPowers) {
		return nil, errors.New("invalid validator set data")
	}
	validators := make(map[string]*types.Validator, len(ValidatorAddresses))
	for i, addressStr := range ValidatorAddresses {
		var pubKey ed25519.PubKeyEd25519
		copy(pubKey[:], ValidatorPubKeyBytes[i])
		validators[addressStr] = &types.Validator{
			PubKey:      pubKey,
			VotingPower: ValidatorVotingPowers[i],
		}
	}
	return validators, nil
}


// Testnet

// address = hexEncode(sha256(pubKey)[:20])
var ValidatorAddressesTestnet = []string{
	"75D14638A3DD5779D886646B640BD3BBA729902E",
}

// public key on ed25519 curve (base64 encoded)
var ValidatorB64EncodePubKeysTestnet = []string{
	"tVNtS1VjHDoC7zho5WkEdzQXiOe4407LvE+FuBsxgJM=",
}

// ValidatorPubKeyBytes are results from base-64 decoding ValidatorB64EncodePubKeys
var ValidatorPubKeyBytesTestnet = [][]byte{
	[]byte{0xb5, 0x53, 0x6d, 0x4b, 0x55, 0x63, 0x1c, 0x3a, 0x2, 0xef, 0x38, 0x68, 0xe5, 0x69, 0x4, 0x77, 0x34, 0x17, 0x88, 0xe7, 0xb8, 0xe3, 0x4e, 0xcb, 0xbc, 0x4f, 0x85, 0xb8, 0x1b, 0x31, 0x80, 0x93},
}

var ValidatorVotingPowersTestnet = []int64{
	1000000000000,
}

var validatorMapTestnet, _ = NewFixedValidatorsTestnet()

func NewFixedValidatorsTestnet() (map[string]*types.Validator, error) {
	if len(ValidatorAddressesTestnet) != len(ValidatorPubKeyBytesTestnet) || len(ValidatorAddressesTestnet) != len(ValidatorVotingPowersTestnet) {
		return nil, errors.New("invalid validator set data")
	}
	validators := make(map[string]*types.Validator, len(ValidatorAddressesTestnet))
	for i, addressStr := range ValidatorAddressesTestnet {
		var pubKey ed25519.PubKeyEd25519
		copy(pubKey[:], ValidatorPubKeyBytesTestnet[i])
		validators[addressStr] = &types.Validator{
			PubKey:      pubKey,
			VotingPower: ValidatorVotingPowersTestnet[i],
		}
	}
	return validators, nil
}

