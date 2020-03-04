package syncker

import (
	"context"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wire"
	"time"
)

type BeaconPeerState struct {
	Timestamp      int64
	BestViewHash   string
	BestViewHeight uint64
	processed      bool
}

type BeaconSyncProcess struct {
	status                      string //stop, running
	isCommittee                 bool
	FewBlockBehind              bool
	beaconPeerStates            map[string]BeaconPeerState //sender -> state
	beaconPeerStateCh           chan *wire.MessagePeerState
	server                      Server
	chain                       Chain
	beaconPool                  *BlkPool
	s2bSyncProcess              *S2BSyncProcess
	actionCh                    chan func()
	lastUpdateConfirmCrossShard uint64
}

func NewBeaconSyncProcess(server Server, chain BeaconChainInterface) *BeaconSyncProcess {
	s := &BeaconSyncProcess{
		status:                      STOP_SYNC,
		server:                      server,
		chain:                       chain,
		beaconPool:                  NewBlkPool("BeaconPool"),
		beaconPeerStates:            make(map[string]BeaconPeerState),
		beaconPeerStateCh:           make(chan *wire.MessagePeerState),
		actionCh:                    make(chan func()),
		lastUpdateConfirmCrossShard: 1,
	}
	s.s2bSyncProcess = NewS2BSyncProcess(server, s, chain)
	go s.syncBeacon()
	go s.insertBeaconBlockFromPool()
	go s.updateConfirmCrossShard()
	return s
}

func (s *BeaconSyncProcess) start(isCommittee bool) {
	if s.status == RUNNING_SYNC {
		return
	}
	s.isCommittee = isCommittee
	s.status = RUNNING_SYNC

	go func() {
		ticker := time.NewTicker(time.Millisecond * 500)
		for {
			if s.isCommittee {
				s.s2bSyncProcess.start()
			} else {
				s.s2bSyncProcess.stop()
			}

			select {
			case f := <-s.actionCh:
				f()
			case beaconPeerState := <-s.beaconPeerStateCh:
				s.beaconPeerStates[beaconPeerState.SenderID] = BeaconPeerState{
					Timestamp:      beaconPeerState.Timestamp,
					BestViewHash:   beaconPeerState.Beacon.BlockHash.String(),
					BestViewHeight: beaconPeerState.Beacon.Height,
				}
			case <-ticker.C:
				s.chain.SetReady(s.FewBlockBehind)
			}
			if s.status != RUNNING_SYNC {
				time.Sleep(time.Second)
				continue
			}
		}
	}()

}

func (s *BeaconSyncProcess) stop() {
	s.status = STOP_SYNC
	s.s2bSyncProcess.stop()
}

func (s *BeaconSyncProcess) getBeaconPeerStates() map[string]BeaconPeerState {
	res := make(chan map[string]BeaconPeerState)
	s.actionCh <- func() {
		ps := make(map[string]BeaconPeerState)
		for k, v := range s.beaconPeerStates {
			ps[k] = v
		}
		res <- ps
	}
	return <-res
}

func (s *BeaconSyncProcess) updateConfirmCrossShard() {
	//TODO: update lastUpdateConfirmCrossShard using DB
	fmt.Println("crossdebug lastUpdateConfirmCrossShard ", s.lastUpdateConfirmCrossShard)
	for {
		if s.lastUpdateConfirmCrossShard > s.chain.GetBestViewHeight() { //TODO: get confirm height
			time.Sleep(time.Second * 5)
			continue
		}

		blk, err := s.server.FetchBeaconBlock(s.lastUpdateConfirmCrossShard)
		if err != nil {
			time.Sleep(time.Second * 5)
			continue
		}
		s.lastUpdateConfirmCrossShard++
		for fromSID, shardState := range blk.Body.ShardState {
			for _, blockState := range shardState {
				for _, toSID := range blockState.CrossShard {
					fmt.Printf("crossdebug: from %d to %d with crossshard height %d confirmed by beacon hash %s\n", int(fromSID), int(toSID), blockState.Height, blk.Hash().String())
					err := s.server.StoreBeaconHashConfirmCrossShardHeight(int(fromSID), int(toSID), blockState.Height, blk.Hash().String())
					if err != nil {
						panic(err)
					}
				}
			}
		}
	}
}

func (s *BeaconSyncProcess) insertBeaconBlockFromPool() {
	defer func() {
		if s.FewBlockBehind {
			time.AfterFunc(time.Millisecond*100, s.insertBeaconBlockFromPool)
		} else {
			time.AfterFunc(time.Second*1, s.insertBeaconBlockFromPool)
		}
	}()

	if !s.FewBlockBehind {
		return
	}
	var blk common.BlockPoolInterface
	if s.isCommittee {
		blk = s.beaconPool.GetNextBlock(s.chain.GetBestViewHash(), true)
	} else {
		blk = s.beaconPool.GetNextBlock(s.chain.GetBestViewHash(), false)
	}

	if isNil(blk) {
		return
	}

	fmt.Println("Syncker: Insert beacon from pool", blk.(common.BlockInterface).GetHeight())
	s.beaconPool.RemoveBlock(blk.Hash().String())
	if err := s.chain.ValidateBlockSignatures(blk.(common.BlockInterface), s.chain.GetCommittee()); err != nil {
		return
	}

	if err := s.chain.InsertBlk(blk.(common.BlockInterface)); err != nil {
	}
}

func (s *BeaconSyncProcess) syncBeacon() {
	for {
		requestCnt := 0
		if s.status != RUNNING_SYNC {
			s.FewBlockBehind = false
			time.Sleep(time.Second)
			continue
		}

		for peerID, pState := range s.getBeaconPeerStates() {
			requestCnt += s.streamFromPeer(peerID, pState)
		}

		//last check, if we still need to sync more
		if requestCnt > 0 {
			s.FewBlockBehind = false
		} else {
			if len(s.beaconPeerStates) > 0 {
				s.FewBlockBehind = true
			}
			time.Sleep(time.Second * 5)
		}
	}
}

func (s *BeaconSyncProcess) streamFromPeer(peerID string, pState BeaconPeerState) (requestCnt int) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	blockBuffer := []common.BlockInterface{}
	defer func() {
		if requestCnt == 0 {
			pState.processed = true
		}
		cancel()
	}()

	toHeight := pState.BestViewHeight
	//process param
	if !s.isCommittee { //if not beacon committee, not insert the newest block (incase we need to revert beacon block)
		toHeight -= 1
	}

	if toHeight <= s.chain.GetBestViewHeight() {
		return
	}

	//stream
	ch, err := s.server.RequestBeaconBlocksViaStream(ctx, peerID, s.chain.GetBestViewHeight()+1, toHeight)
	if err != nil {
		fmt.Println("Syncker: create channel fail")
		return
	}

	//receive
	requestCnt++
	insertTime := time.Now()
	for {
		select {
		case blk := <-ch:
			if !isNil(blk) {
				blockBuffer = append(blockBuffer, blk)
			}

			if len(blockBuffer) >= 350 || (len(blockBuffer) > 0 && (isNil(blk) || time.Since(insertTime) > time.Millisecond*1000)) {
				insertBlkCnt := 0
				for {
					if successBlk, err := InsertBatchBlock(s.chain, blockBuffer); err != nil {
						return
					} else {
						insertBlkCnt += successBlk
						if successBlk >= len(blockBuffer) {
							break
						}
						blockBuffer = blockBuffer[successBlk:]
					}
				}
				insertTime = time.Now()
				blockBuffer = []common.BlockInterface{}
			}
			if isNil(blk) && len(blockBuffer) == 0 {
				return
			}
		}
	}
}
