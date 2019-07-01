package mempool

import (
	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/pubsub"
	"testing"
	"time"
)
var (
	beaconPoolTest *BeaconPool
	err error
	pb = pubsub.NewPubSubManager()
	beaconBlock2 = &blockchain.BeaconBlock{
		Header: blockchain.BeaconHeader{
			Height: 2,
		},
	}
	beaconBlock3 = &blockchain.BeaconBlock{
		Header: blockchain.BeaconHeader{
			Height: 3,
		},
	}
	beaconBlock4 = &blockchain.BeaconBlock{
		Header: blockchain.BeaconHeader{
			Height: 4,
		},
	}
	beaconBlock5 = &blockchain.BeaconBlock{
		Header: blockchain.BeaconHeader{
			Height: 5,
		},
	}
	beaconBlock6 = &blockchain.BeaconBlock{
		Header: blockchain.BeaconHeader{
			Height: 6,
		},
	}
	pendingBlocks = []*blockchain.BeaconBlock{}
	validBlocks = []*blockchain.BeaconBlock{}
	defaultLatestValidHeight = uint64(1)
	testLatestValidHeight = uint64(4)
)
var InitBeaconPoolTest = func() (_ struct{}) {
	beaconPoolTest = new(BeaconPool)
	beaconPoolTest.latestValidHeight = 1
	beaconPoolTest.validPool = []*blockchain.BeaconBlock{}
	beaconPoolTest.pendingPool = make(map[uint64]*blockchain.BeaconBlock)
	beaconPoolTest.conflictedPool = make(map[common.Hash]*blockchain.BeaconBlock)
	beaconPoolTest.config = BeaconPoolConfig{
		MaxValidBlock:   MAX_VALID_BEACON_BLK_IN_POOL,
		MaxPendingBlock: MAX_PENDING_BEACON_BLK_IN_POOL,
		CacheSize:       BEACON_CACHE_SIZE,
	}
	beaconPoolTest.cache, _ = lru.New(beaconPool.config.CacheSize)
	go pb.Start()
	for i := testLatestValidHeight + 1; i < MAX_VALID_BEACON_BLK_IN_POOL + testLatestValidHeight+2; i++ {
		beaconBlock := &blockchain.BeaconBlock{
			Header: blockchain.BeaconHeader{
				Height: uint64(i),
			},
		}
		validBlocks = append(validBlocks, beaconBlock)
	}
	for i := MAX_VALID_BEACON_BLK_IN_POOL + testLatestValidHeight + 2; i < MAX_VALID_BEACON_BLK_IN_POOL + MAX_PENDING_BEACON_BLK_IN_POOL + testLatestValidHeight+3; i++ {
		beaconBlock := &blockchain.BeaconBlock{
			Header: blockchain.BeaconHeader{
				Height: uint64(i),
			},
		}
		pendingBlocks = append(pendingBlocks, beaconBlock)
	}
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}
var _ = func() (_ struct{}) {
	GetBeaconPool()
	InitBeaconPool(pb)
	go pb.Start()
	for i := 2; i < MAX_VALID_BEACON_BLK_IN_POOL + 3; i++ {
		beaconBlock := &blockchain.BeaconBlock{
			Header: blockchain.BeaconHeader{
				Height: uint64(i),
			},
		}
		validBlocks = append(validBlocks, beaconBlock)
	}
	for i := MAX_VALID_BEACON_BLK_IN_POOL + 3; i < MAX_VALID_BEACON_BLK_IN_POOL + MAX_PENDING_BEACON_BLK_IN_POOL + 3; i++ {
		beaconBlock := &blockchain.BeaconBlock{
			Header: blockchain.BeaconHeader{
				Height: uint64(i),
			},
		}
		pendingBlocks = append(pendingBlocks, beaconBlock)
	}
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()
func ResetBeaconPool(){
	beaconPool = new(BeaconPool)
	beaconPool.latestValidHeight = 1
	beaconPool.validPool = []*blockchain.BeaconBlock{}
	beaconPool.pendingPool = make(map[uint64]*blockchain.BeaconBlock)
	beaconPool.conflictedPool = make(map[common.Hash]*blockchain.BeaconBlock)
	beaconPool.config = BeaconPoolConfig{
		MaxValidBlock:   MAX_VALID_BEACON_BLK_IN_POOL,
		MaxPendingBlock: MAX_PENDING_BEACON_BLK_IN_POOL,
		CacheSize:       BEACON_CACHE_SIZE,
	}
	beaconPool.cache, _ = lru.New(beaconPool.config.CacheSize)
	InitBeaconPool(pb)
}
func TestGetbeaconPool(t *testing.T) {
	if beaconPool.latestValidHeight != 1 {
		t.Fatal("Invalid Latest valid height")
	}
	if beaconPool.validPool == nil {
		t.Fatal("Invalid Valid Pool")
	}
	if beaconPool.pendingPool == nil {
		t.Fatal("Invalid Pending Pool")
	}
	if beaconPool.conflictedPool == nil {
		t.Fatal("Invalid Conflicted Pool")
	}
	if beaconPool.config.MaxValidBlock != MAX_VALID_BEACON_BLK_IN_POOL {
		t.Fatal("Invalid Max Valid Pool")
	}
	if beaconPool.config.MaxPendingBlock != MAX_PENDING_BEACON_BLK_IN_POOL {
		t.Fatal("Invalid Max Pending Pool")
	}
	if beaconPool.config.CacheSize != BEACON_CACHE_SIZE {
		t.Fatal("Invalid Beacon Cache Size")
	}
	if beaconPool.cache == nil {
		t.Fatal("Invalid Cache")
	}
}
func TestBeaconPoolSetBeaconState(t *testing.T) {
	beaconPool.SetBeaconState(0)
	if beaconPool.latestValidHeight != 1 {
		t.Fatal("Invalid Latest Valid Height")
	}
	latestValidHeight := beaconPool.latestValidHeight
	beaconPool.SetBeaconState(latestValidHeight+10)
	if beaconPool.latestValidHeight != latestValidHeight + 10{
		t.Fatalf("Height Should be set %+v but get %+v \n", latestValidHeight+10, beaconPool.latestValidHeight)
	}
	ResetBeaconPool()
}
func TestBeaconPoolInitBeaconPool(t *testing.T) {
	latestValidHeight := beaconPool.latestValidHeight
	//InitBeaconPool(pb)
	// because blockchain beacon beststate is nil => return latestvalidheight is 0
	if beaconPool.latestValidHeight != latestValidHeight {
		t.Fatalf("Height Should be set %+v but get %+v \n", latestValidHeight, beaconPool.latestValidHeight)
	}
	if beaconPool.PubSubManager == nil {
		t.Fatal("Pubsub manager is nil after init")
	}
	if beaconPool.RoleInCommitteesEvent == nil {
		t.Fatal("Role Event is nil after init")
	}
	ResetBeaconPool()
}
func TestBeaconPoolTestStart(t *testing.T) {
	cQuit := make(chan struct{})
	go beaconPool.Start(cQuit)
	// send event
	go pb.PublishMessage(pubsub.NewMessage(pubsub.BeaconRoleTopic, true))
	<-time.Tick(500 * time.Millisecond)
	beaconPool.mtx.RLock()
	if beaconPool.RoleInCommittees != true {
		t.Fatal("Fail to get Role In committees from event")
	}
	beaconPool.mtx.RUnlock()
	go pb.PublishMessage(pubsub.NewMessage(pubsub.BeaconRoleTopic, -1))
	<-time.Tick(500 * time.Millisecond)
	beaconPool.mtx.RLock()
	if beaconPool.RoleInCommittees != true {
		t.Fatal("Should not get wrong format information")
	}
	beaconPool.mtx.RUnlock()
	go pb.PublishMessage(pubsub.NewMessage(pubsub.BeaconRoleTopic, false))
	<-time.Tick(500 * time.Millisecond)
	beaconPool.mtx.RLock()
	if beaconPool.RoleInCommittees != false {
		t.Fatal("Fail to get Role In committees from event")
	}
	beaconPool.mtx.RUnlock()
	go pb.PublishMessage(pubsub.NewMessage(pubsub.BeaconRoleTopic, true))
	<-time.Tick(500 * time.Millisecond)
	beaconPool.mtx.RLock()
	if beaconPool.RoleInCommittees != true {
		t.Fatal("Fail to get Role In committees from event")
	}
	beaconPool.mtx.RUnlock()
	close(cQuit)
	<-time.Tick(500 * time.Millisecond)
	beaconPool.mtx.RLock()
	if beaconPool.RoleInCommittees != false {
		t.Fatal("Fail to set default Role In committees when beacon pool is stop")
	}
	beaconPool.mtx.RUnlock()
	ResetBeaconPool()
}

func TestBeaconPoolGetBeaconState(t *testing.T) {
	latestValidHeight := beaconPool.latestValidHeight
	if beaconPool.GetBeaconState() != latestValidHeight {
		t.Fatal("Fail when try to get state of beacon pool")
	}
	ResetBeaconPool()
}

func TestBeaconPoolValidateBeaconBlock(t *testing.T) {
	// skip old block
	InitBeaconPoolTest()
	//Test receive old block than latestvalidheight
	beaconPoolTest.SetBeaconState(4)
	err = beaconPoolTest.validateBeaconBlock(beaconBlock3,false)
	if err != nil {
		if err.(*BlockPoolError).Code != ErrCodeMessage[OldBlockError].Code {
			t.Fatalf("Block %+v should return error %+v but get %+v", beaconBlock3.Header.Height, ErrCodeMessage[OldBlockError].Code, err.(*BlockPoolError).Code)
		}
		if block, ok := beaconPoolTest.conflictedPool[beaconBlock3.Header.Hash()]; !ok {
			t.Fatalf("Block %+v should be in conflict pool but get %+v", beaconBlock3.Header.Height, block.Header.Height)
		}
	}
	err = beaconPoolTest.validateBeaconBlock(beaconBlock4,true)
	if err != nil {
		if err.(*BlockPoolError).Code != ErrCodeMessage[OldBlockError].Code {
			t.Fatalf("Block %+v should return error %+v but get %+v", beaconBlock3.Header.Height, ErrCodeMessage[OldBlockError].Code, err.(*BlockPoolError).Code)
		}
		if block, ok := beaconPoolTest.conflictedPool[beaconBlock3.Header.Hash()]; !ok {
			t.Fatalf("Block %+v should be in conflict pool but get %+v", beaconBlock3.Header.Height, block.Header.Height)
		}
	}
	err = beaconPoolTest.validateBeaconBlock(beaconBlock2, false)
	if err == nil {
		t.Fatalf("Block %+v should be discard with state %+v", beaconBlock2.Header.Height, beaconPoolTest.latestValidHeight)
	} else {
		if err.(*BlockPoolError).Code != ErrCodeMessage[OldBlockError].Code {
			t.Fatalf("Block %+v should be discard with state %+v, error should be %+v but get %+v", beaconBlock2.Header.Height, beaconPoolTest.latestValidHeight, ErrCodeMessage[OldBlockError].Code, err.(*BlockPoolError).Code)
		}
	}
	//test duplicate and pending
	err = beaconPoolTest.validateBeaconBlock(beaconBlock6, false)
	if err != nil {
		t.Fatalf("Block %+v should be able to get in pending pool, state %+v", beaconBlock6.Header.Height, beaconPoolTest.latestValidHeight)
	}
	beaconPoolTest.pendingPool[beaconBlock6.Header.Height] = beaconBlock6
	if _, ok := beaconPoolTest.pendingPool[beaconBlock6.Header.Height]; !ok {
		t.Fatalf("Block %+v should be in pending pool", beaconBlock6.Header.Height)
	}
	err = beaconPoolTest.validateBeaconBlock(beaconBlock6, false)
	if err == nil {
		t.Fatalf("Block %+v should be duplicate \n", beaconBlock6.Header.Height)
	} else {
		if err.(*BlockPoolError).Code != ErrCodeMessage[DuplicateBlockError].Code {
			t.Fatalf("Block %+v should return error %+v but get %+v", beaconBlock6.Header.Height, ErrCodeMessage[DuplicateBlockError].Code, err.(*BlockPoolError).Code)
		}
	}
	// ignore if block is duplicate or exceed pool size or not
	err = beaconPoolTest.validateBeaconBlock(beaconBlock6, true)
	if err != nil {
		t.Fatalf("Block %+v should not be duplicate \n", beaconBlock6.Header.Height)
	}
	delete(beaconPoolTest.pendingPool, beaconBlock6.Header.Height)
	for index, beaconBlock := range pendingBlocks {
		if index < len(pendingBlocks) - 1 {
			beaconPoolTest.pendingPool[beaconBlock.Header.Height] = beaconBlock
		}  else {
			err = beaconPoolTest.validateBeaconBlock(beaconBlock, false)
			if err == nil {
				t.Fatalf("Block %+v exceed pending pool capacity %+v \n", beaconBlock.Header.Height, len(beaconPoolTest.pendingPool))
			} else {
				if err.(*BlockPoolError).Code != ErrCodeMessage[MaxPoolSizeError].Code {
					t.Fatalf("Block %+v should return error %+v but get %+v", beaconBlock.Header.Height, ErrCodeMessage[MaxPoolSizeError].Code, err.(*BlockPoolError).Code)
				}
			}
		}
	}
	for index, beaconBlock := range validBlocks {
		if index < len(validBlocks) - 1 {
			beaconPoolTest.validPool = append(beaconPoolTest.validPool, beaconBlock)
			beaconPoolTest.latestValidHeight = beaconBlock.Header.Height
		}  else {
			err = beaconPoolTest.validateBeaconBlock(beaconBlock, false)
			if err == nil {
				t.Fatalf("Block %+v exceed valid pool capacity %+v plus pending pool capacity %+v \n", beaconBlock.Header.Height, len(beaconPool.validPool), len(beaconPool.pendingPool))
			} else {
				if err.(*BlockPoolError).Code != ErrCodeMessage[MaxPoolSizeError].Code {
					t.Fatalf("Block %+v should return error %+v but get %+v", beaconBlock.Header.Height, ErrCodeMessage[MaxPoolSizeError].Code, err.(*BlockPoolError).Code)
				}
			}
		}
	}
}
