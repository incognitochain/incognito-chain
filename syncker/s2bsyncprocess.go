package syncker

import (
	"context"
	"fmt"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wire"
)

//TODO: Request sync must include all block that in pool
type S2BPeerState struct {
	Timestamp int64
	Height    uint64
	processed bool
}

type S2BSyncProcess struct {
	status            string                           //stop, running
	s2bPeerState      map[string]map[byte]S2BPeerState //sender -> state
	s2bPeerStateCh    chan *wire.MessagePeerState
	Server            Server
	beaconSyncProcess *BeaconSyncProcess
	beaconChain       BeaconChainInterface
	s2bPool           *BlkPool
	actionCh          chan func()
}

func NewS2BSyncProcess(server Server, beaconSyncProc *BeaconSyncProcess, beaconChain BeaconChainInterface) *S2BSyncProcess {

	var isOutdatedBlock = func(blk interface{}) bool {
		if blk.(*blockchain.ShardToBeaconBlock).GetHeight() < beaconChain.GetShardBestViewHeight()[byte(blk.(*blockchain.ShardToBeaconBlock).GetShardID())] {
			return true
		}
		return false
	}

	s := &S2BSyncProcess{
		status:            STOP_SYNC,
		Server:            server,
		beaconChain:       beaconChain,
		s2bPool:           NewBlkPool("ShardToBeaconPool", isOutdatedBlock),
		beaconSyncProcess: beaconSyncProc,
		s2bPeerState:      make(map[string]map[byte]S2BPeerState),
		s2bPeerStateCh:    make(chan *wire.MessagePeerState),
		actionCh:          make(chan func()),
	}

	go s.syncS2BPoolProcess()

	go func() {
		ticker := time.NewTicker(time.Millisecond * 500)
		for {
			select {
			case f := <-s.actionCh:
				f()
			case s2bPeerState := <-s.s2bPeerStateCh:
				if s.s2bPeerState[s2bPeerState.SenderID] == nil {
					s.s2bPeerState[s2bPeerState.SenderID] = make(map[byte]S2BPeerState)
				}
				for sid, v := range s2bPeerState.ShardToBeaconPool {
					s.s2bPeerState[s2bPeerState.SenderID][sid] = S2BPeerState{
						Timestamp: s2bPeerState.Timestamp,
						Height:    v[len(v)-1],
					}
				}
			case <-ticker.C:
				for sender, s2bPeerState := range s.s2bPeerState {
					for sid, ps := range s2bPeerState {
						if ps.Timestamp < time.Now().Unix()-10 {
							delete(s2bPeerState, sid)
						}
					}
					if len(s2bPeerState) == 0 {
						delete(s.s2bPeerState, sender)
					}
				}
			}
		}
	}()

	return s
}

func (s *S2BSyncProcess) start() {
	if s.status == RUNNING_SYNC {
		return
	}
	s.status = RUNNING_SYNC

}

func (s *S2BSyncProcess) stop() {
	s.status = STOP_SYNC
}

//helper function to access map in atomic way
func (s *S2BSyncProcess) getS2BPeerState() map[string]map[byte]S2BPeerState {
	res := make(chan map[string]map[byte]S2BPeerState)
	s.actionCh <- func() {
		ps := make(map[string]map[byte]S2BPeerState)
		psState := make(map[byte]S2BPeerState)
		for k, v := range s.s2bPeerState {
			for i, j := range v {
				psState[i] = j
			}
			ps[k] = psState
		}
		res <- ps
	}
	return <-res
}

func (s *S2BSyncProcess) syncS2BPoolProcess() {
	for {
		requestCnt := 0
		if !s.beaconSyncProcess.isCatchUp || s.status != RUNNING_SYNC {
			time.Sleep(time.Second)
			continue
		}
		for peerID, pState := range s.getS2BPeerState() {
			requestCnt += s.streamFromPeer(peerID, pState)
		}

		//last check, if we still need to sync more
		if requestCnt == 0 {
			//s.S2BPool.Print()
			time.Sleep(time.Second * 5)
		}

	}

}

func (s *S2BSyncProcess) streamFromPeer(peerID string, senderState map[byte]S2BPeerState) (requestCnt int) {
	var requestS2BBlockFromAShardPeer = func(fromSID byte, shardState S2BPeerState) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		if shardState.processed {
			return
		}
		toHeight := shardState.Height

		if time.Now().Unix()-senderState[fromSID].Timestamp > 30 {
			return
		}

		//retrieve information from pool -> request missing block
		//not retrieve genesis block (if height = 0, we get block shard height = 1)
		sID := byte(fromSID)
		viewHash := s.beaconChain.GetShardBestViewHash()[sID]
		viewHeight := s.beaconChain.GetShardBestViewHeight()[sID]
		if viewHeight == 0 {
			blk := *s.Server.GetChainParam().GenesisShardBlock
			blk.Header.ShardID = sID
			viewHash = *blk.Hash()
			viewHeight = 1
		}

		reqFromHeight := viewHeight + 1
		if viewHeight < toHeight {
			validS2BBlock := s.s2bPool.GetLongestChain(viewHash.String())
			if len(validS2BBlock) > 100 {
				return
			}
			if len(validS2BBlock) > 0 {
				reqFromHeight = validS2BBlock[len(validS2BBlock)-1].GetHeight() + 1
			}
		}
		if reqFromHeight+blockchain.DefaultMaxBlkReqPerPeer <= toHeight {
			toHeight = reqFromHeight + blockchain.DefaultMaxBlkReqPerPeer
		}
		if reqFromHeight > toHeight {
			return
		}
		//start request
		requestCnt++
		ch, err := s.Server.RequestShardBlocksViaStream(ctx, peerID, int(sID), reqFromHeight, toHeight)
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
					fmt.Println("Syncker Insert shard2beacon block", sID, blk.GetHeight(), blk.Hash().String(), blk.(common.BlockPoolInterface).GetPrevHash())
					s.s2bPool.AddBlock(blk.(common.BlockPoolInterface))
				} else {
					shardState.processed = true
					break
				}
			}
			if blkCnt > blockchain.DefaultMaxBlkReqPerPeer {
				break
			}
		}

	}
	//fmt.Printf("Syncker received S2BState %v, start sync\n", pState)
	for fromSID, shardState := range senderState {
		requestS2BBlockFromAShardPeer(fromSID, shardState)
	}
	return
}
