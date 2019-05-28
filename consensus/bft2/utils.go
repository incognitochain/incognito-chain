package bft2

import (
	"github.com/constant-money/constant-chain/blockchain"
	"github.com/constant-money/constant-chain/common"
	"math"
	"time"
)

func (e *BFTEngine) validateBlockWithCurrentView(b Block, v View) bool {
	return true
}

func (e *BFTEngine) createNewBlockFromCurrentView(v View) Block {
	return "sd"
}

func (e *BFTEngine) createCurrentView() View {
	view := View{}
	nodeType, shardID := blockchain.GetBestStateBeacon().GetPubkeyNodeRole(e.UserKeySet.GetPublicKeyB58())
	round := e.predictCurrentRound(nodeType, shardID)
	role := blockchain.GetBestStateBeacon().GetPubkeyChainRole(e.UserKeySet.GetPublicKeyB58(), round)
	view.Role = Role{nodeType, role, shardID}
	view.Round = round
	view.PubKey = e.UserKeySet.GetPublicKeyB58()
	view.ShardHeight[shardID] = blockchain.GetBestStateShard(shardID).ShardHeight
	view.BeaconHeight = blockchain.GetBestStateBeacon().BeaconHeight
	view.CommitteeSize.Beacon = len(blockchain.GetBestStateBeacon().BeaconCommittee)
	view.CommitteeSize.Shard[shardID] = len(blockchain.GetBestStateShard(shardID).ShardCommittee)

	return view
}

func (e *BFTEngine) predictCurrentRound(nodeType string, shardID byte) int {
	timeSinceLastBlk := time.Since(time.Unix(0, 0))
	if nodeType == common.BEACON_ROLE {
		timeSinceLastBlk = time.Since(time.Unix(blockchain.GetBestStateBeacon().BestBlock.Header.Timestamp, 0))
	} else if nodeType == common.SHARD_ROLE {
		timeSinceLastBlk = time.Since(time.Unix(blockchain.GetBestStateShard(shardID).BestBlock.Header.Timestamp, 0))
	} else {
		return 0
	}

	roundTime := LISTEN_TIMEOUT + PRE_PREPARE_TIMEOUT + PREPARE_TIMEOUT + COMMIT_TIMEOUT
	if timeSinceLastBlk > roundTime {
		return int(math.Floor(timeSinceLastBlk.Seconds() / roundTime.Seconds()))
	}

	return 0
}

func (e *BFTEngine) getMajorityView() View {
	view := View{}

	for k, v := range e.ValidatorsView {

	}
}
