package bridge

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type AddToken struct {
	NewListTokens map[common.Hash]map[uint]config.Vault `json:"NewListTokens"`
}

func (a *AddToken) StringSlice() ([]string, error) {
	contentBytes, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return []string{
		strconv.Itoa(metadataCommon.BridgeAggAddTokenMeta),
		base64.StdEncoding.EncodeToString(contentBytes),
	}, nil
}

func (a *AddToken) FromStringSlice(source []string) error {
	if len(source) != 2 {
		return fmt.Errorf("len of instruction need to be 2 but get %d", len(source))
	}
	if strconv.Itoa(metadataCommon.BridgeAggAddTokenMeta) != source[0] {
		return fmt.Errorf("metaType need to be %d but get %s", metadataCommon.BridgeAggAddTokenMeta, source[0])
	}
	contentBytes, err := base64.StdEncoding.DecodeString(source[1])
	if err != nil {
		return err
	}
	err = json.Unmarshal(contentBytes, &a)
	if err != nil {
		return err
	}
	return nil
}
