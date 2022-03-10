package consensus_v2

import (
	"fmt"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/wallet"
)

func TestGetMiningKeyFromPrivateSeed(t *testing.T) {
	// incKey1 := "112t8rnXCerQX2RRd8KhPfsFCj2rrBYUx42FZJKgRFcdBfg36Mid3ygKyMn5LSc5LBHsxqapRaN6xMav7bGhA6VtGUzNNYuA9Y78CB5oGkti"
	// incKey2 := "112t8rnXYgxipKvTJJfHg7tQhcdmA2R1jPpCPmXg37Xi1VfgrFzWFuNy4U6828q1yfbD7VEdutD63HfVYAqL6U32joXVjqdkfUP52LnNGXda"
	// incKey3 := "112t8rnXe3Jxg5d1Rejg2fB1NwnqNsr94RCT3PX14h5NNDjrdgLeEWFkqcMNamKCHask1Gx46g5WYZDKHKx7kzLVD7h1cgvU6NxNijkyGmA9"
	// incKey4 := "112t8rnY2gqonwhnhGD6rKeEXkbJDB7DHUtZQKC8SfLci6ABb5eCEj4o7ezWBZWaGbu7CJ1R1mrADGqmRjugg42GeA6jhaXbNDeP2HUr8udw"

	incKeys := []string{
		"112t8rnX3Cz3ud5HG7EnM8U3apQqbtpmbAjbe5Uox3Lj7aJg85AAko91JVwXjC7wNHENWtMmFqPvQEJrYS8WhYYekDJmH1c5GBkL4YCHKV8o",
		"112t8rnXaRvy95YLYEt78ovWCY2Azi7pCrU4v7BCHm6AjfpUNYUDMbksf6WATFjY4tHUr4g6D5bmiKgMgmjB9ih1eNHifwqdRzC6Eqv23FHD",
		"112t8rnXbxgX9xpiJ8f4z7NhAbPY77XY9BxCEHegzhWeR2Vm19YURuxsTYqZDFkK9Nk16ERmtbXW4oGU2ww6P1WiDv2rvBwq9HgsUabhH7EB",
		"112t8rnY3iZhsADoE3EJMddgHEJCSCxrhqpixuqr7jwzLqsebMkK6sEuSBGWDav35tWfomGW5urs4rEoR9VrTNrmwmwFZkRvSQrSCYudiXLg",
	}

	seeds := [][]byte{}
	for _, key := range incKeys {
		keyWallet, _ := wallet.Base58CheckDeserialize(key)
		privateSeedBytes := common.HashB(common.HashB(keyWallet.KeySet.PrivateKey))
		privateSeed := base58.Base58Check{}.Encode(privateSeedBytes, common.Base58Version)
		miningKey, _ := GetMiningKeyFromPrivateSeed(privateSeed)
		seeds = append(seeds, miningKey.PriKey[common.BridgeConsensus])
		fmt.Printf("%#v\n", miningKey.PubKey[common.BridgeConsensus])
	}
}
