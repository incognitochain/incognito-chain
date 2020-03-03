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
	Status                      string //stop, running
	IsCommittee                 bool
	FewBlockBehind              bool
	BeaconPeerStates            map[string]BeaconPeerState //sender -> state
	BeaconPeerStateCh           chan *wire.MessagePeerState
	Server                      Server
	Chain                       Chain
	ChainID                     int
	BeaconPool                  *BlkPool
	S2BSyncProcess              *S2BSyncProcess
	actionCh                    chan func()
	lastUpdateConfirmCrossShard uint64
}

func NewBeaconSyncProcess(server Server, chain BeaconChainInterface) *BeaconSyncProcess {
	s := &BeaconSyncProcess{
		Status:                      STOP_SYNC,
		Server:                      server,
		Chain:                       chain,
		BeaconPool:                  NewBlkPool("BeaconPool"),
		BeaconPeerStates:            make(map[string]BeaconPeerState),
		BeaconPeerStateCh:           make(chan *wire.MessagePeerState),
		actionCh:                    make(chan func()),
		lastUpdateConfirmCrossShard: 1,
	}
	s.S2BSyncProcess = NewS2BSyncProcess(server, s, chain)
	go s.syncBeacon()
	go s.insertBeaconBlockFromPool()
	go s.updateConfirmCrossShard()
	return s
}

func (s *BeaconSyncProcess) updateConfirmCrossShard() {
	//TODO: update lastUpdateConfirmCrossShard using DB
	fmt.Println("crossdebug lastUpdateConfirmCrossShard ", s.lastUpdateConfirmCrossShard)
	for {
		if s.lastUpdateConfirmCrossShard > s.Chain.GetBestViewHeight() { //TODO: get confirm height
			time.Sleep(time.Second * 5)
			continue
		}

		blk, err := s.Server.FetchBeaconBlock(s.lastUpdateConfirmCrossShard)
		if err != nil {
			time.Sleep(time.Second * 5)
			continue
		}
		s.lastUpdateConfirmCrossShard++
		for fromSID, shardState := range blk.Body.ShardState {
			for _, blockState := range shardState {
				for _, toSID := range blockState.CrossShard {
					fmt.Printf("crossdebug: from %d to %d with crossshard height %d confirmed by beacon hash %s\n", int(fromSID), int(toSID), blockState.Height, blk.Hash().String())
					err := s.Server.StoreBeaconHashConfirmCrossShardHeight(int(fromSID), int(toSID), blockState.Height, blk.Hash().String())
					if err != nil {
						panic(err)
					}
				}
			}
		}
	}
}

func (s *BeaconSyncProcess) Start(chainID int) {
	if s.Status == RUNNING_SYNC {
		return
	}
	s.ChainID = chainID
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
			case beaconPeerState := <-s.BeaconPeerStateCh:
				s.BeaconPeerStates[beaconPeerState.SenderID] = BeaconPeerState{
					Timestamp:      beaconPeerState.Timestamp,
					BestViewHash:   beaconPeerState.Beacon.BlockHash.String(),
					BestViewHeight: beaconPeerState.Beacon.Height,
				}
			case <-ticker.C:
				s.Chain.SetReady(s.FewBlockBehind)
			}
		}
	}()
	s.S2BSyncProcess.Start()
}

func (s *BeaconSyncProcess) GetBeaconPeerStates() map[string]BeaconPeerState {
	res := make(chan map[string]BeaconPeerState)
	s.actionCh <- func() {
		ps := make(map[string]BeaconPeerState)
		for k, v := range s.BeaconPeerStates {
			ps[k] = v
		}
		res <- ps
	}
	return <-res
}

func (s *BeaconSyncProcess) Stop() {
	s.Status = STOP_SYNC
	s.S2BSyncProcess.Stop()
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
	if s.ChainID == -1 {
		blk = s.BeaconPool.GetNextBlock(s.Chain.GetBestViewHash(), true)
	} else {
		blk = s.BeaconPool.GetNextBlock(s.Chain.GetBestViewHash(), false)
	}

	if isNil(blk) {
		return
	}
	fmt.Println("Syncker: Insert beacon from pool", blk.(common.BlockInterface).GetHeight())
	s.BeaconPool.RemoveBlock(blk.Hash().String())
	if err := s.Chain.ValidateBlockSignatures(blk.(common.BlockInterface), s.Chain.GetCommittee()); err != nil {
		return
	}

	if err := s.Chain.InsertBlk(blk.(common.BlockInterface)); err != nil {
	}
}

func (s *BeaconSyncProcess) syncBeacon() {
	for {
		requestCnt := 0
		if s.Status != RUNNING_SYNC {
			s.FewBlockBehind = false
			time.Sleep(time.Second)
			continue
		}

		for peerID, pState := range s.GetBeaconPeerStates() {
			requestCnt += s.streamFromPeer(peerID, pState)
		}

		//last check, if we still need to sync more
		if requestCnt > 0 {
			s.FewBlockBehind = false
		} else {
			if len(s.BeaconPeerStates) > 0 {
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
	if s.ChainID != -1 { //if not beacon committee, not insert the newest block (incase we need to revert beacon block)
		toHeight -= 1
	}

	if toHeight <= s.Chain.GetBestViewHeight() {
		return
	}

	//stream
	ch, err := s.Server.RequestBeaconBlocksViaStream(ctx, peerID, s.Chain.GetBestViewHeight()+1, toHeight)
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
					if successBlk, err := InsertBatchBlock(s.Chain, blockBuffer); err != nil {
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
