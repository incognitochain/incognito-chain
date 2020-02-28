package syncker

import (
	"context"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/peerv2/proto"
	"github.com/incognitochain/incognito-chain/wire"
	"time"
)

//TODO: Request sync must include all block that in pool
type S2BPeerState struct {
	Timestamp int64
	Height    map[int]uint64 //shardid -> height
	processed bool
}

type S2BSyncProcess struct {
	Status            string                  //stop, running
	S2BPeerState      map[string]S2BPeerState //sender -> state
	S2BPeerStateCh    chan *wire.MessagePeerState
	Server            Server
	BeaconSyncProcess *BeaconSyncProcess
	BeaconChain       BeaconChainInterface
	S2BPool           *BlkPool
	actionCh          chan func()
}

func NewS2BSyncProcess(server Server, beaconSyncProc *BeaconSyncProcess, beaconChain BeaconChainInterface) *S2BSyncProcess {
	s := &S2BSyncProcess{
		Status:            STOP_SYNC,
		Server:            server,
		BeaconChain:       beaconChain,
		S2BPool:           NewBlkPool("ShardToBeaconPool"),
		BeaconSyncProcess: beaconSyncProc,
		S2BPeerState:      make(map[string]S2BPeerState),
		S2BPeerStateCh:    make(chan *wire.MessagePeerState),
		actionCh:          make(chan func()),
	}

	go s.syncS2BPoolProcess()
	return s
}

func (s *S2BSyncProcess) Start() {
	if s.Status == RUNNING_SYNC {
		return
	}
	s.Status = RUNNING_SYNC
	go func() {
		ticker := time.NewTicker(time.Millisecond * 500)
		for {
			if s.Status != RUNNING_SYNC {
				time.Sleep(time.Second)
				continue
			}
			select {
			case f := <-s.actionCh:
				f()
			case s2bPeerState := <-s.S2BPeerStateCh:
				s2bState := make(map[int]uint64)
				for sid, v := range s2bPeerState.ShardToBeaconPool {
					s2bState[int(sid)] = v[len(v)-1]
				}
				s.S2BPeerState[s2bPeerState.SenderID] = S2BPeerState{
					Timestamp: s2bPeerState.Timestamp,
					Height:    s2bState,
				}
			case <-ticker.C:

			}
		}
	}()

}

func (s *S2BSyncProcess) Stop() {
	s.Status = STOP_SYNC
}

func (s *S2BSyncProcess) GetS2BPeerState() map[string]S2BPeerState {
	res := make(chan map[string]S2BPeerState)
	s.actionCh <- func() {
		ps := make(map[string]S2BPeerState)
		for k, v := range s.S2BPeerState {
			ps[k] = v
		}
		res <- ps
	}
	return <-res
}

func (s *S2BSyncProcess) syncS2BPoolProcess() {
	for {
		requestCnt := 0
		if !s.BeaconSyncProcess.FewBlockBehind || s.Status != RUNNING_SYNC {
			time.Sleep(time.Second)
			continue
		}
		for peerID, pState := range s.GetS2BPeerState() {
			requestCnt += s.streamFromPeer(peerID, pState)
		}

		//last check, if we still need to sync more
		if requestCnt == 0 {
			time.Sleep(time.Second * 5)
		}

	}

}

func (s *S2BSyncProcess) streamFromPeer(peerID string, pState S2BPeerState) (requestCnt int) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer func() {
		if requestCnt == 0 {
			pState.processed = true
		}
		cancel()
	}()

	if pState.processed {
		return
	}

	for fromSID, toHeight := range pState.Height {
		if time.Now().Unix()-pState.Timestamp > 30 {
			return
		}

		//retrieve information from pool -> request missing block
		//not retrieve genesis block (if height = 0, we get block shard height = 1)
		sID := byte(fromSID)
		viewHash := s.BeaconChain.GetShardBestViewHash()[sID]
		viewHeight := s.BeaconChain.GetShardBestViewHeight()[sID]
		if viewHeight == 0 {
			blk := *s.Server.GetChainParam().GenesisShardBlock
			blk.Header.ShardID = sID
			viewHash = *blk.Hash()
			viewHeight = 1
		}

		reqFromHeight := viewHeight + 1
		if viewHeight < toHeight {
			validS2BBlock := s.S2BPool.GetLongestChain(viewHash.String())
			if len(validS2BBlock) > 100 {
				return
			}
			if len(validS2BBlock) > 0 {
				reqFromHeight = validS2BBlock[len(validS2BBlock)-1].GetHeight() + 1
			}
		}
		if viewHeight+100 > toHeight {
			toHeight = viewHeight + 100
		}

		//start request
		requestCnt++
		ch, err := s.Server.RequestBlocksViaStream(ctx, peerID, int(sID), proto.BlkType_BlkS2B, reqFromHeight, reqFromHeight, toHeight, "")
		if err != nil {
			fmt.Println("Syncker: create channel fail")
			return
		}

		//start receive
		blkCnt := int(0)
		for {
			blkCnt++
			select {
			case blk := <-ch:
				if !isNil(blk) {
					//fmt.Println("Syncker: Insert shard2beacon block", blk.GetHeight(), blk.Hash().String(), blk.(common.BlockPoolInterface).GetPrevHash())
					s.S2BPool.AddBlock(blk.(common.BlockPoolInterface))
				}
			}
			if blkCnt > 100 {
				break
			}
		}
	}
	return
}
