package bridgehub

import (
	"fmt"
	"strconv"

	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type BridgeHubStartNetworkInstruction struct {
	BridgePoolPubKey string `json:"BridgePoolPubKey"`
}

func (i *BridgeHubStartNetworkInstruction) FromStringSlice(str []string) error {
	if len(str) != 2 {
		return fmt.Errorf("Invalid length expect %v but get %v", 2, len(str))
	}
	if str[0] != strconv.Itoa(metadataCommon.BridgeHubStartNetwork) {
		return fmt.Errorf("Invalid type expect %v but get %s", metadataCommon.BridgeHubStartNetwork, str[0])
	}
	if str[1] == "" {
		return fmt.Errorf("BridgePoolPubKey cannot be empty")
	}
	i.BridgePoolPubKey = str[1]
	return nil
}

func (i *BridgeHubStartNetworkInstruction) ToStringSlice() []string {
	res := []string{}
	res = append(res, strconv.Itoa(metadataCommon.BridgeHubStartNetwork))
	res = append(res, i.BridgePoolPubKey)
	return res
}
