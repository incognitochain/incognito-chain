package blsbft

import (
	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain/types"
)

type actorV2 struct {
	actorBase
	currentTime     int64
	currentTimeSlot int64
	proposeHistory  *lru.Cache

	receiveBlockByHeight map[uint64][]*ProposeBlockInfo  //blockHeight -> blockInfo
	receiveBlockByHash   map[string]*ProposeBlockInfo    //blockHash -> blockInfo
	voteHistory          map[uint64]types.BlockInterface // bestview height (previsous height )-> block
}

func (actorV2 *actorV2) Destroy() {
	actorV2.actorBase.Destroy()
	close(actorV2.destroyCh)
}
