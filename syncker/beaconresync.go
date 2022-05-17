package syncker

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
)

type SyncPair struct {
	from, to uint64
}

type ResyncManager struct {
	Net          Network
	Chain        *blockchain.BlockChain
	HeightFilter *AVLSync
	Data         map[uint64]types.BeaconBlock
	locker       *sync.RWMutex
	RequestSync  chan uint64
	RequestPair  chan SyncPair
	ExpHeight    chan uint64
	PreSync      chan struct {
		from uint64
		to   uint64
		blks []types.BlockInterface
	}
}

func (reSync *ResyncManager) Start() {
	for {
		if reSync.Net.IsReady() {
			break
		}
		time.Sleep(time.Second)
	}
	for {
		select {
		case h := <-reSync.RequestSync:
			Logger.Debugf("Got new request from %v, to %v", h, h+350)
			reSync.resyncPair(h, h+350)
		case h := <-reSync.RequestPair:
			reSync.resyncPair(h.from, h.to)
		case preSync := <-reSync.PreSync:
			reSync.HeightFilter.InsertPair(preSync.from, preSync.to)
			blks := map[uint64]types.BeaconBlock{}
			for _, blk := range preSync.blks {
				blks[blk.GetHeight()] = *blk.(*types.BeaconBlock)
			}
			reSync.addData(preSync.from, preSync.to, blks)
		case h := <-reSync.ExpHeight:
			Logger.Infof("Got exp height %v", h)
			if found := reSync.HeightFilter.Find(h); found != nil {
				Logger.Infof("Found node %v, start delete value from %v to %v", h, found.key, found.Value)
				reSync.deleteData(found.key, found.Value)
				reSync.HeightFilter.Remove(h)
			}
		}
	}
}

func NewReSyncManager(
	net Network,
	chain *blockchain.BlockChain,
) *ResyncManager {
	res := &ResyncManager{
		RequestSync: make(chan uint64, 10),
		RequestPair: make(chan SyncPair, 10),
		ExpHeight:   make(chan uint64, 10),
		locker:      &sync.RWMutex{},
		Data:        map[uint64]types.BeaconBlock{},
		PreSync: make(chan struct {
			from uint64
			to   uint64
			blks []types.BlockInterface
		}, 10),
	}
	cacher := cache.New(time.Second, time.Second)
	cacher.OnEvicted(func(key string, value interface{}) {
		h, err := strconv.ParseUint(key, 10, 64)
		if err != nil {
			panic(err)
			Logger.Error(err)
		}
		res.ExpHeight <- h
	})
	res.Chain = chain
	res.Net = net
	res.HeightFilter = NewAVLSync(cacher)
	return res
}

func (reSync *ResyncManager) deleteData(from, to uint64) {
	reSync.locker.Lock()
	defer reSync.locker.Unlock()
	for height := from; height <= to; height++ {
		delete(reSync.Data, height)
	}
}

func (reSync *ResyncManager) addData(from, to uint64, blks map[uint64]types.BeaconBlock) {
	reSync.locker.Lock()
	defer reSync.locker.Unlock()
	Logger.Debugf("Sync ok, got data from %v to %v", from, to)
	for h := from; h <= to; h++ {
		reSync.Data[h] = blks[h]
	}
	reSync.HeightFilter.cache.Add(strconv.FormatUint(from, 10), struct{}{}, 30*time.Second)
}

func (reSync *ResyncManager) resyncPair(from, to uint64) error {
	if !reSync.Net.IsReady() {
		time.Sleep(50 * time.Millisecond)
		go func(f, t uint64) {
			reSync.RequestPair <- SyncPair{from: f, to: t}
		}(from, to)
		return errors.Errorf("Requester is not ready")
	}
	uid := common.GenUUID()
	newPairs := reSync.HeightFilter.SearchMissingPair(from, to)
	reSync.HeightFilter.DisplayInOrder()
	for id := 0; id < len(newPairs[0]); id++ {
		from, to := newPairs[0][id], newPairs[1][id]
		for i := 0; i <= 5; i++ {
			got, blks, err := reSync.syncData(from, to, uid)
			if err != nil {
				Logger.Errorf("Sync block from %v to %v failed, got to %v, error %v uid %v ", from, to, got, err, uid)
			}
			if got == 0 {
				got = from - 1
			}
			if len(blks) > 0 {
				reSync.HeightFilter.InsertPair(from, got)
				reSync.addData(from, got, blks)
			}
			if got+1 <= to {
				from = got + 1
				time.Sleep(10 * time.Millisecond)
			} else {
				break
			}
		}
	}
	return nil
}

func (reSync *ResyncManager) checkData(from, to uint64) {
	reSync.locker.RLock()
	defer reSync.locker.RUnlock()
	f, t := uint64(0), uint64(0)
	for h := from; h <= to; h++ {
		if _, ok := reSync.Data[h]; !ok {
			if f == 0 {
				f = h
			}
		} else {
			if f != 0 {
				t = h - 1
				go func(f, t uint64) {
					reSync.RequestPair <- SyncPair{from: f, to: t}
				}(f, t)
				f = 0
			}
		}
	}
	if (f != 0) && (t == 0) {
		go func(f, t uint64) {
			reSync.RequestPair <- SyncPair{from: f, to: t}
		}(f, to)
	}
}

func (reSync *ResyncManager) getData(from, to uint64) (
	blks []types.BeaconBlock,
	err error,
) {
	reSync.locker.RLock()
	defer reSync.locker.RUnlock()
	for h := from; h <= to; h++ {
		if blk, ok := reSync.Data[h]; ok {
			blks = append(blks, blk)
		} else {
			go reSync.checkData(from, to)
			return nil, errors.Errorf("Not enough data, just have from %v to %v", from, h-1)
		}
	}
	return blks, nil
}

func (reSync *ResyncManager) syncData(from, to uint64, uid string) (
	gotTo uint64,
	blks map[uint64]types.BeaconBlock,
	err error,
) {
	timeout := 1 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	ctx = context.WithValue(ctx, common.CtxUUID, uid)
	defer cancel()
	blks = map[uint64]types.BeaconBlock{}
	ch, err := reSync.Net.RequestBeaconBlocksViaStream(ctx, "", from, to)
	if err != nil || ch == nil {
		err = errors.Errorf("Syncker: create channel failed, error %v", err)
		return 0, nil, err
	}
	msgCounter := 0
	got := uint64(0)
	for {
		select {
		case blk := <-ch:
			if !isNil(blk) {
				beaconBlk, ok := blk.(*types.BeaconBlock)
				if !ok {
					return got, blks, errors.Errorf("Received invalid block type")
				}
				blkHash, err := reSync.Chain.GetBeaconBlockHashByHeight(reSync.Chain.BeaconChain.GetFinalView(), reSync.Chain.GetBeaconBestState(), beaconBlk.Header.Height)
				if err != nil {
					return got, blks, err
				}
				if blkHash.String() != beaconBlk.Hash().String() {
					return got, blks, errors.Errorf("Got beacon block hash %v, expected %v, height %v, uuid %v", beaconBlk.Hash().String(), blkHash.String(), beaconBlk.Header.Height, uid)
				}
				blks[beaconBlk.Header.Height] = *beaconBlk
				msgCounter++
				got = beaconBlk.Header.Height
			} else {
				if msgCounter == int(to-from+1) {
					return to, blks, nil
				}
				if got != to {
					err = errors.Errorf("Can not get all BeaconBlockByHash in time, missed block %v to %v, uid %v", got+1, to, uid)
				} else {
					err = nil
				}
				return got, blks, err
			}
		case <-ctx.Done():
			if got != to {
				err = errors.Errorf("Can not get all BeaconBlockByHash in time, missed block %v to %v, uid %v", got+1, to, uid)
			} else {
				err = nil
			}
			return got, blks, err
		}
	}
}
