package consensus_v2

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/wallet"
	"testing"
)

func TestGetMiningKeyFromPrivateSeed(t *testing.T){
	incKey1 := "112t8rnXCerQX2RRd8KhPfsFCj2rrBYUx42FZJKgRFcdBfg36Mid3ygKyMn5LSc5LBHsxqapRaN6xMav7bGhA6VtGUzNNYuA9Y78CB5oGkti"
	incKey2 := "112t8rnXYgxipKvTJJfHg7tQhcdmA2R1jPpCPmXg37Xi1VfgrFzWFuNy4U6828q1yfbD7VEdutD63HfVYAqL6U32joXVjqdkfUP52LnNGXda"
	incKey3 := "112t8rnXe3Jxg5d1Rejg2fB1NwnqNsr94RCT3PX14h5NNDjrdgLeEWFkqcMNamKCHask1Gx46g5WYZDKHKx7kzLVD7h1cgvU6NxNijkyGmA9"
	incKey4 := "112t8rnY2gqonwhnhGD6rKeEXkbJDB7DHUtZQKC8SfLci6ABb5eCEj4o7ezWBZWaGbu7CJ1R1mrADGqmRjugg42GeA6jhaXbNDeP2HUr8udw"
	incKeys := []string{incKey1, incKey2, incKey3, incKey4}

	seeds := [][]byte{}
	for _, key := range incKeys {
		keyWallet ,  _ := wallet.Base58CheckDeserialize(key)
		privateSeedBytes := common.HashB(common.HashB(keyWallet.KeySet.PrivateKey))
		privateSeed := base58.Base58Check{}.Encode(privateSeedBytes, common.Base58Version)
		miningKey, _ := GetMiningKeyFromPrivateSeed(privateSeed)
		seeds = append(seeds, miningKey.PriKey[common.BridgeConsensus])
		fmt.Printf("%#v\n", miningKey.PriKey[common.BridgeConsensus])
	}
}