package bridgeagg

import (
	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type ShieldData struct {
	BlockHash  []byte      `json:"BlockHash"`
	TxIndex    uint        `json:"TxIndex"`
	Proof      []byte      `json:"Proof"`
	IncTokenID common.Hash `json:"IncTokenID"`
	NetworkID  uint        `json:"NetworkID"`
}

type ShieldRequest struct {
	ShieldDatas []ShieldData `json:"ShieldDatas"`
	metadataCommon.MetadataBase
}
