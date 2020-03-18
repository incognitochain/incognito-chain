package bnb

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"

	//"github.com/stretchr/testify/assert"

	"github.com/tendermint/tendermint/types"
	"testing"
)

var _ = func() (_ struct{}) {
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	Logger.log.Info("This runs before init()!")
	return
}()

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	m.Run()
}

func TestHeaderChain_ReceiveNewHeader(t *testing.T) {
	header1Json := `
		{
			"version": {
			  "block": 10,
			  "app": 0
			},
			"chain_id": "Binance-Chain-Tigris",
			"height": 1,
			"time": "2019-04-18T05:59:26.228734998Z",
			"num_txs": 0,
			"total_txs": 0,
			"last_block_id": {
			  "hash": "",
			  "parts": {
				"total": 0,
				"hash": ""
			  }
			},
			"last_commit_hash": "",
			"data_hash": "",
			"validators_hash": "43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49",
			"next_validators_hash": "43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49",
			"consensus_hash": "294D8FBD0B94B767A7EBA9840F299A3586DA7FE6B5DEAD3B7EECBA193C400F93",
			"app_hash": "",
			"last_results_hash": "",
			"evidence_hash": "",
			"proposer_address": "14CFCE69B645F3F88BAF08EA5B77FA521E4480F9"
		}
	`

	header2Json := `
		{
			"version": {
			  "block": 10,
			  "app": 0
			},
			"chain_id": "Binance-Chain-Tigris",
			"height": 2,
			"time": "2019-04-18T06:07:02.154340159Z",
			"num_txs": 0,
			"total_txs": 0,
			"last_block_id": {
			  "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
			  "parts": {
				"total": 1,
				"hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
			  }
			},
			"last_commit_hash": "DBDCD4FB32B47D560A7FB045D254C997EDB456FB98F49495B9C1808981DB57AE",
			"data_hash": "",
			"validators_hash": "43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49",
			"next_validators_hash": "43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49",
			"consensus_hash": "294D8FBD0B94B767A7EBA9840F299A3586DA7FE6B5DEAD3B7EECBA193C400F93",
			"app_hash": "B6E4D322EDA04F783EBDCD3D59AA908D211675F890EBC05A893C2093D101D009",
			"last_results_hash": "",
			"evidence_hash": "",
			"proposer_address": "14CFCE69B645F3F88BAF08EA5B77FA521E4480F9"
		  }
		`
	lastCommitJson1 := `
			{
				"block_id": {
				  "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
				  "parts": {
					"total": 1,
					"hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
				  }
				},
				"precommits": [
				  null,
				  {
					"type": 2,
					"height": 1,
					"round": 1,
					"block_id": {
					  "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
					  "parts": {
						"total": 1,
						"hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
					  }
					},
					"timestamp": "2019-04-18T06:07:02.233083179Z",
					"validator_address": "14CFCE69B645F3F88BAF08EA5B77FA521E4480F9",
					"validator_index": 1,
					"signature": "jbuHaCfT/uZDdbLcEHGWZHuqc1pkV6hlgaqq7wfrPf5NNnBAGpqcD5R9V8dGq+avAcg7G/7fxwXK6Pt+Q2d/Cw=="
				  },
				  {
					"type": 2,
					"height": 1,
					"round": 1,
					"block_id": {
					  "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
					  "parts": {
						"total": 1,
						"hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
					  }
					},
					"timestamp": "2019-04-18T06:07:02.153180042Z",
					"validator_address": "17B42E8F284D3CA0E420262F89CD76C749BB12C9",
					"validator_index": 2,
					"signature": "w4N2FT1fHEX8HrHY5jiq8ApNsuLBKK8r7KxwwLCcItydBME97IIK0SDcJOuNCuRd7zMfgp5JRGiJuSJSB8L6Aw=="
				  },
				  {
					"type": 2,
					"height": 1,
					"round": 1,
					"block_id": {
					  "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
					  "parts": {
						"total": 1,
						"hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
					  }
					},
					"timestamp": "2019-04-18T06:07:02.23279118Z",
					"validator_address": "3CD4AABABDDEB7ABFEA9618732E331077A861D2B",
					"validator_index": 3,
					"signature": "mzZuMGM/q+92GiY4ZL2Qv4WvRcy7Vh91UwK5uhXbSA5j+R1RabmTjZpFBthQsOc2TGTyCvsE5en5F1pwfiWtBg=="
				  },
				  {
					"type": 2,
					"height": 1,
					"round": 1,
					"block_id": {
					  "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
					  "parts": {
						"total": 1,
						"hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
					  }
					},
					"timestamp": "2019-04-18T06:07:02.154340159Z",
					"validator_address": "414FB3BBA216AF84C47E07D6EBAA2DCFC3563A2F",
					"validator_index": 4,
					"signature": "JETscHroKuMBkntQ7ALoWfTwptq5yyODmiWtA7Ac9v1//EsQJ/0jVE/YHmU1LYFXuJ68U3I7l6NH7jMEgxtnAA=="
				  },
				  {
					"type": 2,
					"height": 1,
					"round": 1,
					"block_id": {
					  "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
					  "parts": {
						"total": 1,
						"hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
					  }
					},
					"timestamp": "2019-04-18T06:07:02.23639005Z",
					"validator_address": "71F253E6FEA9EDD4B4753F5483549FE4F0F3A21C",
					"validator_index": 5,
					"signature": "n8oPNTYLjPIGzrxaWkaAXZQa6iqcGXsVfXjp279slYgEqVAcGuHwMubzPPFQG1KWTuvpfwAhKBPNq4nqs0gLBg=="
				  },
				  null,
				  {
					"type": 2,
					"height": 1,
					"round": 1,
					"block_id": {
					  "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
					  "parts": {
						"total": 1,
						"hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
					  }
					},
					"timestamp": "2019-04-18T06:07:02.232155249Z",
					"validator_address": "A71E5CD078B8C5C7B1AF88BCE84DD70B0557D93E",
					"validator_index": 7,
					"signature": "0nuhfUXcH6Xfcx3KIhCIpbBZfjEU3FQ3O7yGBHqnwwljCR/FG06+7wTIO2lfPkAnakXwspj+S36dBTkngQeADw=="
				  },
				  {
					"type": 2,
					"height": 1,
					"round": 1,
					"block_id": {
					  "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
					  "parts": {
						"total": 1,
						"hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
					  }
					},
					"timestamp": "2019-04-18T06:07:02.153428422Z",
					"validator_address": "A9157B3FA6EB4C1E396B9B746E95327A07DC42E5",
					"validator_index": 8,
					"signature": "q/Yw0yasJttINNpUCQRf9YwhqLoVcdEihuRK8vfakhhJ6MFonX3Ynt5ZGoLlyIbVyaenVmP7lcefheFknc7NAQ=="
				  },
				  {
					"type": 2,
					"height": 1,
					"round": 1,
					"block_id": {
					  "hash": "494A674FCDADB0F67BC99CBC33B6F8061E1B76A786485B662EE422C6918C86C8",
					  "parts": {
						"total": 1,
						"hash": "AF4CEDFAB7FB7E5263D9C5A72F5ECD3F3A4C186D8D3C10D80C95FCB2BCEE0E11"
					  }
					},
					"timestamp": "2019-04-18T06:07:02.153867648Z",
					"validator_address": "B0FBB52FF7EE93CC476DFE6B74FA1FC88584F30D",
					"validator_index": 9,
					"signature": "TV2VzCqxuAWQaDFHm4dXpXtEcYw/EdMxMtD5wSMoxLef2SPvnQlkG/sY2zOb3KNtgOmjelkdkzqeA0kAPIlLBw=="
				  },
				  null
				]
			  }
		`

	headerJson3 := `
		{
			"version": {
			  "block": 10,
			  "app": 0
			},
			"chain_id": "Binance-Chain-Tigris",
			"height": 3,
			"time": "2019-04-18T06:07:07.739441221Z",
			"num_txs": 0,
			"total_txs": 0,
			"last_block_id": {
			  "hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
			  "parts": {
				"total": 1,
				"hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
			  }
			},
			"last_commit_hash": "E7BDF3448318E45BA69BAE0BC3B94F5005279B4FFA1FEDF51AAF0F4CF1F6D529",
			"data_hash": "",
			"validators_hash": "43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49",
			"next_validators_hash": "43C53A50D8653EF8CF1E5716DA68120FB51B636DC6D111EC3277B098ECD42D49",
			"consensus_hash": "294D8FBD0B94B767A7EBA9840F299A3586DA7FE6B5DEAD3B7EECBA193C400F93",
			"app_hash": "04C4F444F7101F26AAFC4F80507BEB61C066354BBB4B94502E6AF7E7DBE69CA1",
			"last_results_hash": "",
			"evidence_hash": "",
			"proposer_address": "14CFCE69B645F3F88BAF08EA5B77FA521E4480F9"
		  }
		`
	lastCommitJson2 := `
		{
			"block_id": {
			  "hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
			  "parts": {
				"total": 1,
				"hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
			  }
			},
			"precommits": [
			  null,
			  {
				"type": 2,
				"height": 2,
				"round": 1,
				"block_id": {
				  "hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
				  "parts": {
					"total": 1,
					"hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
				  }
				},
				"timestamp": "2019-04-18T06:07:07.81740164Z",
				"validator_address": "14CFCE69B645F3F88BAF08EA5B77FA521E4480F9",
				"validator_index": 1,
				"signature": "VyHgok5hmgzjUJGPUPwAquXtmuLaJu4p2B5kzCvY/oI06LC2bbUPLUrkNFAtmz7XkRhVqUHMwgODSCsxpqpeAA=="
			  },
			  {
				"type": 2,
				"height": 2,
				"round": 1,
				"block_id": {
				  "hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
				  "parts": {
					"total": 1,
					"hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
				  }
				},
				"timestamp": "2019-04-18T06:07:07.7393153Z",
				"validator_address": "17B42E8F284D3CA0E420262F89CD76C749BB12C9",
				"validator_index": 2,
				"signature": "kv6bsey/YxsRHuxLo2juzDHPrD/ZkxY432ii1gHDnuFXhbqwZM0xrKb+BU7dz7DLhCYJ8wuw/yo43vYP9/WCAQ=="
			  },
			  {
				"type": 2,
				"height": 2,
				"round": 1,
				"block_id": {
				  "hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
				  "parts": {
					"total": 1,
					"hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
				  }
				},
				"timestamp": "2019-04-18T06:07:07.817334959Z",
				"validator_address": "3CD4AABABDDEB7ABFEA9618732E331077A861D2B",
				"validator_index": 3,
				"signature": "vS7ASlEZsg+5PFAJ8YYDozCBswsbG0VrdQKBh6R6JhhJLt84KSqSaVkAGQtzFH397kAxDJpDEtULNkzd6v6NAw=="
			  },
			  {
				"type": 2,
				"height": 2,
				"round": 1,
				"block_id": {
				  "hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
				  "parts": {
					"total": 1,
					"hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
				  }
				},
				"timestamp": "2019-04-18T06:07:07.735960102Z",
				"validator_address": "414FB3BBA216AF84C47E07D6EBAA2DCFC3563A2F",
				"validator_index": 4,
				"signature": "RUxRSxOA1pqqh8LvG2V7BRBanhcD51rC36gQNdztASlx4DVON7UhUPMaDYXKTjm/DQsMMPz8363HHflWz4RPAg=="
			  },
			  {
				"type": 2,
				"height": 2,
				"round": 1,
				"block_id": {
				  "hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
				  "parts": {
					"total": 1,
					"hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
				  }
				},
				"timestamp": "2019-04-18T06:07:07.818207384Z",
				"validator_address": "71F253E6FEA9EDD4B4753F5483549FE4F0F3A21C",
				"validator_index": 5,
				"signature": "zASakyX56qEkiAE+rXFXCjVch2Fr8CcOmUAkoQNvIu4efkK9wAkdTlfwSmDtxN+fa398MkadRGxkLWUHNS0BBg=="
			  },
			  null,
			  {
				"type": 2,
				"height": 2,
				"round": 1,
				"block_id": {
				  "hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
				  "parts": {
					"total": 1,
					"hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
				  }
				},
				"timestamp": "2019-04-18T06:07:07.816888459Z",
				"validator_address": "A71E5CD078B8C5C7B1AF88BCE84DD70B0557D93E",
				"validator_index": 7,
				"signature": "0tW24yXhTOo5eW8hAIwhFZUumdsHB9fdMtJgAKbKcGP79nOoCBYAoEObYITVL/XpfXjHglmDcTHUmTnpqQDCAw=="
			  },
			  {
				"type": 2,
				"height": 2,
				"round": 1,
				"block_id": {
				  "hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
				  "parts": {
					"total": 1,
					"hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
				  }
				},
				"timestamp": "2019-04-18T06:07:07.739441221Z",
				"validator_address": "A9157B3FA6EB4C1E396B9B746E95327A07DC42E5",
				"validator_index": 8,
				"signature": "V+dwIBkgrvZrDP659JP6sW3YejrStULTf2MROly+M8EZdRYfQLt56GcwcvricEl9VcHAHAJILgkENTmZG6WuDw=="
			  },
			  {
				"type": 2,
				"height": 2,
				"round": 1,
				"block_id": {
				  "hash": "B7FB22262EE63921EB260E10A285AC14A96CC794F6298EABB7155DD1B4EF22CF",
				  "parts": {
					"total": 1,
					"hash": "F801520DB85754F18002773752C0020605B3C25A357339826C12D71A1522475D"
				  }
				},
				"timestamp": "2019-04-18T06:07:07.738348307Z",
				"validator_address": "B0FBB52FF7EE93CC476DFE6B74FA1FC88584F30D",
				"validator_index": 9,
				"signature": "zGrjAbOy20MQXk00aj/w5AkG7Bkbl11PQlZjKmEwKlerUXmLXDWBQ8Yez/bVvYJdfZLUek6WGPwvq7bXu4o5AA=="
			  },
			  null
			]
		  }
		`

	header1, err := ParseHeaderFromJson(header1Json)
	assert.Nil(t, err)
	header2, err := ParseHeaderFromJson(header2Json)
	assert.Nil(t, err)
	lastCommit1, err := ParseCommitFromJson(lastCommitJson1)
	assert.Nil(t, err)

	header3, err := ParseHeaderFromJson(headerJson3)
	assert.Nil(t, err)
	lastCommit2, err := ParseCommitFromJson(lastCommitJson2)
	assert.Nil(t, err)

	testcases := []struct {
		header                          *types.Header
		lastCommit                      *types.Commit
		expectedBlockHeight             int
		expectedUnconfirmedHeaderNumber int
	}{
		{header1, nil, 0, 1},
		{header2, lastCommit1, 1, 1},
		{header3, lastCommit2, 2, 1},
	}

	// header chain
	headerChain := new(HeaderChain)
	headerChain.HeaderChain = []*types.Header{}
	headerChain.UnconfirmedHeaders = []*types.Header{}

	var isResult bool

	for _, tc := range testcases {
		fmt.Printf("Receive header with height %v\n", tc.header.Height)
		headerChain, isResult, err = headerChain.ReceiveNewHeader(tc.header, tc.lastCommit, MainnetBNBChainID)
		assert.Nil(t, err)
		assert.Equal(t, true, isResult)
		assert.Equal(t, tc.expectedUnconfirmedHeaderNumber, len(headerChain.UnconfirmedHeaders))
		assert.Equal(t, tc.expectedBlockHeight, len(headerChain.HeaderChain))
	}
}

// block height 66038977
func TestDataHash(t *testing.T) {
	dataHash, _ := hex.DecodeString("E9BDA8FD942F6908A7398B7D5FA1C46553240BC919B0836083878F2C7BA91709")

	data, _ := base64.StdEncoding.DecodeString("5AHwYl3uCmjObcBDChSx/h0kYscxmvL5L8/H9OBFRG7rJxIuQjFGRTFEMjQ2MkM3MzE5QUYyRjkyRkNGQzdGNEUwNDU0NDZFRUIyNy0zNzU4NRoLUFlOLUMzN19CTkIgAigCMKewATiAqJXprU1AARJyCibrWumHIQOg/PhFdQ/yHskF8lUWm2lrY1DCXPaiJgiPQ7T6BMlJSBJAaMkWNjobugia6AVueclUOBJvcWGGfNS0TthGsd9onZhhBcPN42gsUbtUniqB6Rg7DhHcTTktwmu2tAm8tiHgUxigqRIg0KUCIAE=")

	fmt.Printf("Data tx: %v\n", data)
	txs := types.Txs{data}
	//txs[0] = data
	root := txs.Hash()
	isEqual := bytes.Equal(dataHash, root)
	fmt.Println(isEqual)

}
