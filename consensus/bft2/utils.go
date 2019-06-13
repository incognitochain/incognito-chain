package bft2

import (
	"fmt"
	"time"
)

func (e *BFTEngine) getTimeSinceLastBlock() time.Duration {
	return time.Since(time.Unix(e.Chain.GetLastBlockTimeStamp(), 0))
}

func (e *BFTEngine) waitForNextRound() {
	timeSinceLastBlk := e.getTimeSinceLastBlock()
	if timeSinceLastBlk > e.Chain.GetBlkMinTime() {
		return
	}
	time.Sleep(timeSinceLastBlk)
}

/*
Return the round is calculated since the latest block time
*/
func (e *BFTEngine) getCurrentRound() int {
	return int(e.getTimeSinceLastBlock().Seconds() / TIMEOUT.Seconds())
}

func (e *BFTEngine) validateBlockWithCurrentView(b BlockInterface, v View) bool {
	return true
}

func (e *BFTEngine) createNewBlockFromCurrentView(v View) BlockInterface {
	return "sd"
}

func (e *BFTEngine) getValidatorMaxHeight() uint64 {
}

func (e *BFTEngine) createCurrentView() View {
	view := View{}
	view.Round = e.getCurrentRound()
	view.Role = e.Chain.GetRole()
	view.Height = e.Chain.GetHeight()
	view.CommitteeSize = e.Chain.GetCommitteeSize()
	view.PeerID = e.PeerID
	//nodeType, shardID := blockchain.GetBestStateBeacon().GetPubkeyNodeRole(e.UserKeySet.GetPublicKeyB58())
	//round := e.predictCurrentRound(nodeType, shardID)
	//role := blockchain.GetBestStateBeacon().GetPubkeyChainRole(e.UserKeySet.GetPublicKeyB58(), round)
	//view.Role = Role{nodeType, role, shardID}
	//view.Round = round
	//view.PubKey = e.UserKeySet.GetPublicKeyB58()
	//view.ShardHeight[shardID] = blockchain.GetBestStateShard(shardID).ShardHeight
	//view.BeaconHeight = blockchain.GetBestStateBeacon().BeaconHeight
	//view.CommitteeSize.Beacon = len(blockchain.GetBestStateBeacon().BeaconCommittee)
	//view.CommitteeSize.Shard[shardID] = len(blockchain.GetBestStateShard(shardID).ShardCommittee)

	return view
}

//
//func (e *BFTEngine) predictCurrentRound(nodeType string, shardID byte) int {
//	timeSinceLastBlk := time.Since(time.Unix(0, 0))
//	if nodeType == common.BEACON_ROLE {
//		timeSinceLastBlk = time.Since(time.Unix(blockchain.GetBestStateBeacon().BestBlock.Header.Timestamp, 0))
//	} else if nodeType == common.SHARD_ROLE {
//		timeSinceLastBlk = time.Since(time.Unix(blockchain.GetBestStateShard(shardID).BestBlock.Header.Timestamp, 0))
//	} else {
//		return 0
//	}
//
//	if timeSinceLastBlk > TIMEOUT {
//		return int(math.Floor(timeSinceLastBlk.Seconds() / TIMEOUT.Seconds()))
//	}
//
//	return 0
//}

//func (e *BFTEngine) getMajorityView() View {
//view := View{}
//
//for k, v := range e.ValidatorsView {
//
//}
//}

//curView := e.createCurrentView()
//if curView.Role.nodeType == common.BEACON_ROLE {
////if beacon chain -> beacon must be max
//maxHeight := uint64(0)
//for _, v := range e.ValidatorsView {
//nodeType, shardID := blockchain.GetBestStateBeacon().GetPubkeyNodeRole(v.PubKey)
//if nodeType != curView.Role.nodeType && shardID != curView.Role.shardID {
//continue
//}
//if v.BeaconHeight > maxHeight {
//maxHeight = v.BeaconHeight
//}
//}
////check if get max
//if curView.BeaconHeight == maxHeight && curView.Role.role != common.PENDING_ROLE {
//e.IsReady = true
//} else {
//e.IsReady = false
//}
//}
//
//if curView.Role.nodeType == common.SHARD_ROLE {
////if shard chain -> shard height must be max
//maxHeight := uint64(0)
//for _, v := range e.ValidatorsView {
//nodeType, shardID := blockchain.GetBestStateBeacon().GetPubkeyNodeRole(v.PubKey)
//if nodeType != curView.Role.nodeType && shardID != curView.Role.shardID {
//continue
//}
//if v.ShardHeight[curView.Role.shardID] > maxHeight {
//maxHeight = v.ShardHeight[shardID]
//}
//}
//
////check if get max
//if curView.ShardHeight[curView.Role.shardID] == maxHeight && curView.Role.role != common.PENDING_ROLE {
//e.IsReady = true
//} else {
//e.IsReady = false
//}
//}

func (e *BFTEngine) debug(s ...interface{}) {
	if e.PeerID == "1" {
		fmt.Println(s...)
	}

}
