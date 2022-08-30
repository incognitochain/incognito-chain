package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/incognitochain/incognito-chain/blockchain/bridgeagg"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/metadata/rpccaller"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

func TestUnifiedTokenFile(t *testing.T) {
	// set network to testnet2
	os.Setenv(config.NetworkKey, config.TestNetNetwork)
	os.Setenv(config.NetworkVersionKey, config.TestNetVersion2)

	// read file default path config/mainnet/unified_token.json
	config.LoadUnifiedToken([]byte{})

	// call API get list privacy tokens:
	incClient := rpccaller.NewRPCClient()
	// host := "https://mainnet.incognito.org/fullnode"
	host := "https://testnet.incognito.org/fullnode"
	method := "listprivacycustomtoken"
	method2 := "getallbridgetokens"

	params := []string{}

	type Resp struct {
		Result jsonresult.ListCustomToken
		Error  error
	}
	resp := Resp{}
	err := incClient.RPCCall("", host, "", method, params, &resp)
	fmt.Printf("Err call get list pTokens: %v\n", err)
	// fmt.Printf("List ptoken: %+v\n", resp.Result)

	pTokens := resp.Result.ListCustomToken

	type Resp2 struct {
		Result []*rawdbv2.BridgeTokenInfo
		Error  error
	}
	resp2 := Resp2{}
	err = incClient.RPCCall("", host, "", method2, params, &resp2)
	fmt.Printf("Err call get list bridge Tokens: %v\n", err)
	fmt.Printf("List bridge token: %+v\n", resp2.Result)

	bridgeTokens := resp2.Result

	for _, unifiedTokens := range config.UnifiedToken() {
		for unifiedTokenID, vaults := range unifiedTokens {
			// unifiedTokenID must not exist
			for _, pToken := range pTokens {
				if pToken.ID == unifiedTokenID.String() {
					fmt.Errorf("unifiedTokenID is exist: %v\n", unifiedTokenID)
					break
				}
			}

			// vaults: incTokenID and externalTokenID must matching
			for incTokenID, v := range vaults {
				// get expected external token ID from list pTokens
				expectedExtTokenID := []byte{}
				isExistIncTokenID := false
				for _, pToken := range bridgeTokens {
					if pToken.TokenID.String() == incTokenID.String() {
						expectedExtTokenID = pToken.ExternalTokenID
						isExistIncTokenID = true
						break
					}
				}

				actualExtTokenID, _ := bridgeagg.GetExternalTokenIDByNetworkID(v.ExternalTokenID, v.NetworkID)
				if !isExistIncTokenID {
					fmt.Printf("IncTokenID is not exist: %v\n", incTokenID.String())

					// TODO: must check external token is not exist
					for _, pToken := range bridgeTokens {
						if bytes.Equal(pToken.ExternalTokenID, actualExtTokenID) {
							fmt.Errorf("IncTokenID is not exist, but externalTokenID is exist: %v - %v", incTokenID.String(), v.ExternalTokenID)
							break
						}
					}
				} else {
					if !bytes.Equal(expectedExtTokenID, actualExtTokenID) {
						fmt.Errorf("invalid external token id: unifiedTokenID %v - expectedExtTokenID %v - actualExtTokenID %v\n", unifiedTokenID, expectedExtTokenID, actualExtTokenID)
					}
				}
			}
		}
	}
}
