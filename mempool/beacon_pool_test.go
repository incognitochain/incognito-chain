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
			PrevBlockHash: beaconBlock2.Header.Hash(),
		},
	}
	beaconBlock3Forked = &blockchain.BeaconBlock{
		Header: blockchain.BeaconHeader{
			Height: 3,
			PrevBlockHash: common.HashH([]byte{0}),
		},
	}
	beaconBlock4 = &blockchain.BeaconBlock{
		Header: blockchain.BeaconHeader{
			Height: 4,
			PrevBlockHash: beaconBlock3.Header.Hash(),
		},
	}
	beaconBlock5 = &blockchain.BeaconBlock{
		Header: blockchain.BeaconHeader{
			Height: 5,
			PrevBlockHash: beaconBlock4.Header.Hash(),
		},
	}
	beaconBlock6 = &blockchain.BeaconBlock{
		Header: blockchain.BeaconHeader{
			Height: 6,
			PrevBlockHash: beaconBlock5.Header.Hash(),
		},
	}
	beaconBlock7 = &blockchain.BeaconBlock{
		Header: blockchain.BeaconHeader{
			Height: 7,
			PrevBlockHash: beaconBlock6.Header.Hash(),
		},
	}
	pendingBlocks = []*blockchain.BeaconBlock{}
	validBlocks = []*blockchain.BeaconBlock{}
	defaultLatestValidHeight = uint64(1)
	testLatestValidHeight = uint64(4)
)
var InitBeaconPoolTest = func(pubsubManager *pubsub.PubSubManager) {
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
	beaconPoolTest.PubSubManager = pubsubManager
	_, subChanRole, _ := beaconPoolTest.PubSubManager.RegisterNewSubscriber(pubsub.BeaconRoleTopic)
	beaconPoolTest.RoleInCommitteesEvent = subChanRole
}
var _ = func() (_ struct{}) {
	GetBeaconPool()
	InitBeaconPool(pb)
	InitBeaconPoolTest(pb)
	go pb.Start()
	oldBlockHash := common.Hash{}
	for i := testLatestValidHeight + 1; i < MAX_VALID_BEACON_BLK_IN_POOL + testLatestValidHeight+2; i++ {
		beaconBlock := &blockchain.BeaconBlock{
			Header: blockchain.BeaconHeader{
				Height: uint64(i),
			},
		}
		if i != 0 {
			beaconBlock.Header.PrevBlockHash = oldBlockHash
		}
		oldBlockHash = beaconBlock.Header.Hash()
		validBlocks = append(validBlocks, beaconBlock)
	}
	for i := MAX_VALID_BEACON_BLK_IN_POOL + testLatestValidHeight + 2; i < MAX_VALID_BEACON_BLK_IN_POOL + MAX_PENDING_BEACON_BLK_IN_POOL + testLatestValidHeight + 3; i++ {
		beaconBlock := &blockchain.BeaconBlock{
			Header: blockchain.BeaconHeader{
				Height: uint64(i),
			},
		}
		if i != 0 {
			beaconBlock.Header.PrevBlockHash = oldBlockHash
		}
		oldBlockHash = beaconBlock.Header.Hash()
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
	// reset beacon pool test value
	InitBeaconPoolTest(pb)
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
	// Test receive old block than latestvalidheight
	// - Test old block is less than latestvalidheight 2 value => store in conflicted block
	InitBeaconPoolTest(pb)
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
	delete(beaconPoolTest.conflictedPool, beaconBlock3.Header.Hash())
	err = beaconPoolTest.validateBeaconBlock(beaconBlock4,true)
	if err != nil {
		if err.(*BlockPoolError).Code != ErrCodeMessage[OldBlockError].Code {
			t.Fatalf("Block %+v should return error %+v but get %+v", beaconBlock4.Header.Height, ErrCodeMessage[OldBlockError].Code, err.(*BlockPoolError).Code)
		}
		if block, ok := beaconPoolTest.conflictedPool[beaconBlock4.Header.Hash()]; !ok {
			t.Fatalf("Block %+v should be in conflict pool but get %+v", beaconBlock4.Header.Height, block.Header.Height)
		}
	}
	delete(beaconPoolTest.conflictedPool, beaconBlock4.Header.Hash())
	// - Test old block discard and not store in conflicted pool
	err = beaconPoolTest.validateBeaconBlock(beaconBlock2, false)
	if err == nil {
		t.Fatalf("Block %+v should be discard with state %+v", beaconBlock2.Header.Height, beaconPoolTest.latestValidHeight)
	} else {
		if err.(*BlockPoolError).Code != ErrCodeMessage[OldBlockError].Code {
			t.Fatalf("Block %+v should be discard with state %+v, error should be %+v but get %+v", beaconBlock2.Header.Height, beaconPoolTest.latestValidHeight, ErrCodeMessage[OldBlockError].Code, err.(*BlockPoolError).Code)
		}
		if block, ok := beaconPoolTest.conflictedPool[beaconBlock2.Header.Hash()]; ok {
			t.Fatalf("Block %+v should NOT be in conflict pool but get %+v", beaconBlock2.Header.Height, block.Header.Height)
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
func TestBeaconPoolInsertNewBeaconBlockToPool(t *testing.T) {
	InitBeaconPoolTest(pb)
	// Condition 1: check height
	// Test Height is not equal to latestvalidheight + 1 (not expected block)
	isOk := beaconPoolTest.insertNewBeaconBlockToPool(beaconBlock3)
	if isOk {
		t.Fatalf("Block %+v is invalid with state %+v", beaconBlock3.Header.Height, beaconPoolTest.latestValidHeight)
	} else {
		if _, ok := beaconPoolTest.pendingPool[beaconBlock3.Header.Height]; !ok {
			t.Fatalf("Block %+v should be in pending pool", beaconBlock3.Header.Height)
		}
	}
	// Test Height equal to latestvalidheight + 1
	// Condition 2: Pool is full capacity -> push to pending pool
	for index, beaconBlock := range validBlocks {
		if index < len(validBlocks) - 1 {
			beaconPoolTest.validPool = append(beaconPoolTest.validPool, beaconBlock)
			beaconPoolTest.latestValidHeight = beaconBlock.Header.Height
		} else {
			isOk := beaconPoolTest.insertNewBeaconBlockToPool(beaconBlock)
			if isOk {
				t.Fatalf("Block %+v is valid with state %+v but pool cappacity reach max %+v", beaconBlock.Header.Height, beaconPoolTest.latestValidHeight, len(beaconPoolTest.validPool))
			} else {
				if _, ok := beaconPoolTest.pendingPool[beaconBlock.Header.Height]; !ok {
					t.Fatalf("Block %+v should be in pending pool", beaconBlock.Header.Height)
				}
			}
		}
	}
	// reset valid pool and pending pool
	InitBeaconPoolTest(pb)
	// Condition 3: check next block
	// - Next block doesn't exist
	isOk = beaconPoolTest.insertNewBeaconBlockToPool(beaconBlock2)
	if isOk {
		t.Fatalf("Block %+v is invalid because next block does not exit", beaconBlock3.Header.Height)
	} else {
		if len(beaconPoolTest.validPool) != 0 {
			t.Fatalf("No block should enter pool")
		}
		if _, ok := beaconPoolTest.pendingPool[beaconBlock2.Header.Height]; !ok {
			t.Fatalf("Block %+v should be in pending pool ", beaconBlock2.Header.Height)
		}
		if _, ok := beaconPoolTest.pendingPool[beaconBlock3.Header.Height]; ok {
			t.Fatalf("Block %+v exist in pending pool but block %+v still not in pool", beaconBlock3.Header.Height, beaconBlock2.Header.Height)
		}
	}
	delete(beaconPoolTest.pendingPool, beaconBlock2.Header.Height)
	// push next block to pending pool
	// Condition 4: next block does not point to this block
	beaconPoolTest.pendingPool[beaconBlock3Forked.Header.Height] = beaconBlock3Forked
	isOk = beaconPoolTest.insertNewBeaconBlockToPool(beaconBlock2)
	if isOk {
		t.Fatalf("Block %+v is invalid because next block exit but not ponit to this block", beaconBlock3.Header.Height)
	} else {
		//isExist := beaconPool.cache.Contains(beaconBlock2.Header.Hash())
		//if !isExist {
		//	t.Fatalf("Forked block %+v should be push into conflict pool", beaconBlock2.Header.Hash())
		//} else {
		//	err = beaconPoolTest.validateBeaconBlock(beaconBlock2, false)
		//	if err == nil {
		//		t.Fatalf("Forked block %+v should failed to validate", beaconBlock2.Header.Height)
		//	} else {}
		//	if err.(*BlockPoolError).Code != ErrCodeMessage[OldBlockError].Code {
		//		t.Fatal("Wrong Error Code")
		//	}
		//	if err.(*BlockPoolError).Err.Error() != errors.New("Receive Old Block, this block maybe insert to blockchain already or invalid because of fork: "+fmt.Sprintf("%d", beaconBlock2.Header.Height)).Error() {
		//		t.Fatal("Wrong Expected Error")
		//	}
		//}
	}
	// delete forked block out of pool and push next valid block
	delete(beaconPoolTest.pendingPool, beaconBlock3Forked.Header.Height)
	// push next valid block to pending pool => current block should get in valid pool
	beaconPoolTest.pendingPool[beaconBlock3.Header.Height] = beaconBlock3
	isOk = beaconPoolTest.insertNewBeaconBlockToPool(beaconBlock2)
	if !isOk {
		t.Fatalf("Block %+v should be able to get in valid pool", beaconBlock3.Header.Height)
	} else {
		if len(beaconPoolTest.validPool) != 1 {
			t.Fatalf("Expect length pool to be 1 but get %+v", len(beaconPoolTest.validPool))
		}
		tempBlock := beaconPoolTest.validPool[0]
		tempHash := tempBlock.Header.Hash()
		beaconBlock2Hash := beaconBlock2.Header.Hash()
		if !tempHash.IsEqual(&beaconBlock2Hash) && tempBlock.Header.Height != beaconBlock2.Header.Height {
			t.Fatalf("Block %+v with hash %+v expected but get %+v with hash %+v", beaconBlock2.Header.Height, beaconBlock2Hash, tempHash, tempBlock.Header.Height)
		}
	}
	// check lastest valid height
	if beaconPoolTest.latestValidHeight != 2 {
		t.Fatalf("Latest valid height should update to %+v but get %+v", 2, beaconPoolTest.latestValidHeight)
	}
}

func TestBeaconPoolPromotePendingPool(t *testing.T) {
	InitBeaconPoolTest(pb)
	beaconPoolTest.pendingPool[beaconBlock2.Header.Height] = beaconBlock2
	beaconPoolTest.pendingPool[beaconBlock3.Header.Height] = beaconBlock3
	beaconPoolTest.pendingPool[beaconBlock4.Header.Height] = beaconBlock4
	beaconPoolTest.pendingPool[beaconBlock5.Header.Height] = beaconBlock5
	beaconPoolTest.pendingPool[beaconBlock6.Header.Height] = beaconBlock6
	beaconPoolTest.promotePendingPool()
	if len(beaconPoolTest.validPool) != 4 {
		t.Fatalf("Shoud have 4 block in valid pool but get %+v ", len(beaconPoolTest.validPool))
	}
	for index, block := range beaconPoolTest.validPool {
		switch index {
		case 0:
			if block.Header.Height != 2 {
				t.Fatalf("Expect block 2 but get %+v ", block.Header.Height)
			}
		case 1:
			if block.Header.Height != 3 {
				t.Fatalf("Expect block 3 but get %+v ", block.Header.Height)
			}
		case 2:
			if block.Header.Height != 4 {
				t.Fatalf("Expect block 4 but get %+v ", block.Header.Height)
			}
		case 3:
			if block.Header.Height != 5 {
				t.Fatalf("Expect block 5 but get %+v ", block.Header.Height)
			}
		}
	}
	if len(beaconPoolTest.pendingPool) != 1 {
		t.Fatalf("Shoud have 1 block in valid pool but get %+v ", len(beaconPoolTest.pendingPool))
	}
	if _, ok := beaconPoolTest.pendingPool[beaconBlock6.Header.Height]; !ok {
		t.Fatalf("Expect Block %+v in pending pool", beaconBlock6.Header.Height)
	}
}

func TestBeaconPoolAddBeaconBlock(t *testing.T) {
	InitBeaconPoolTest(pb)
	beaconPoolTest.SetBeaconState(testLatestValidHeight)
	for _, block := range validBlocks {
			err := beaconPoolTest.AddBeaconBlock(block)
			if err != nil {
				t.Fatalf("Block %+v should be added into pool but get %+v", block.Header.Height, err )
			}
		}
	if len(beaconPoolTest.validPool) != MAX_VALID_BEACON_BLK_IN_POOL {
		t.Fatalf("Expected number of block %+v in valid pool but get %+v", MAX_VALID_BEACON_BLK_IN_POOL, len(beaconPoolTest.validPool))
	}
	if len(beaconPoolTest.pendingPool) != 1 {
		t.Fatalf("Expected number of block %+v in pending pool but get %+v", 1, len(beaconPoolTest.pendingPool))
	}
	if _, isOk := beaconPoolTest.pendingPool[validBlocks[len(validBlocks)-1].Header.Height]; !isOk {
		t.Fatalf("Expect block %+v to be in pending pool", validBlocks[len(validBlocks)-1].Header.Height)
	}
	delete(beaconPoolTest.pendingPool, validBlocks[len(validBlocks)-1].Header.Height)
	for index, block := range pendingBlocks {
		if index < len(pendingBlocks) - 1 {
			err := beaconPoolTest.AddBeaconBlock(block)
			if err != nil {
				t.Fatalf("Block %+v should be added into pool but get %+v", block.Header.Height, err)
			}
		} else {
			err := beaconPoolTest.AddBeaconBlock(block)
			if err == nil {
				t.Fatalf("Block %+v should NOT be added into pool but get no error", block.Header.Height)
			} else {
				if err.(*BlockPoolError).Code != ErrCodeMessage[MaxPoolSizeError].Code {
					t.Fatalf("Expect err %+v but get %+v", ErrCodeMessage[MaxPoolSizeError], err)
				}
 			}
		}
	}
	if len(beaconPoolTest.pendingPool) != MAX_PENDING_BEACON_BLK_IN_POOL {
		t.Fatalf("Expected number of block %+v in pending pool but get %+v", MAX_PENDING_BEACON_BLK_IN_POOL, len(beaconPoolTest.pendingPool))
	}
}
func TestBeaconPoolUpdateLatestBeaconState(t *testing.T) {
	InitBeaconPoolTest(pb)
	// init state of latestvalidheight
	if beaconPoolTest.latestValidHeight != 1 {
		t.Fatalf("Expect to latestvalidheight is 1 but get %+v", beaconPoolTest.latestValidHeight)
	}
	// with no block and no blockchain state => latestvalidheight is 0
	beaconPoolTest.updateLatestBeaconState()
	if beaconPoolTest.latestValidHeight != 0 {
		t.Fatalf("Expect to latestvalidheight is 0 but get %+v", beaconPoolTest.latestValidHeight)
	}
	// if valid block list is not empty then each time update latest state
	// it will set to the height of last block in valid block list
	beaconPoolTest.validPool = append(beaconPoolTest.validPool, beaconBlock2)
	beaconPoolTest.updateLatestBeaconState()
	if beaconPoolTest.latestValidHeight != 2 {
		t.Fatalf("Expect to latestvalidheight is 0 but get %+v", beaconPoolTest.latestValidHeight)
	}
	beaconPoolTest.validPool = append(beaconPoolTest.validPool, beaconBlock3)
	beaconPoolTest.updateLatestBeaconState()
	if beaconPoolTest.latestValidHeight != 3 {
		t.Fatalf("Expect to latestvalidheight is 0 but get %+v", beaconPoolTest.latestValidHeight)
	}
	beaconPoolTest.validPool = append(beaconPoolTest.validPool, beaconBlock4)
	beaconPoolTest.updateLatestBeaconState()
	if beaconPoolTest.latestValidHeight != 4 {
		t.Fatalf("Expect to latestvalidheight is 0 but get %+v", beaconPoolTest.latestValidHeight)
	}
	beaconPoolTest.validPool = append(beaconPoolTest.validPool, beaconBlock5)
	beaconPoolTest.updateLatestBeaconState()
	if beaconPoolTest.latestValidHeight != 5 {
		t.Fatalf("Expect to latestvalidheight is 0 but get %+v", beaconPoolTest.latestValidHeight)
	}
	beaconPoolTest.validPool = append(beaconPoolTest.validPool, beaconBlock6)
	beaconPoolTest.updateLatestBeaconState()
	if beaconPoolTest.latestValidHeight != 6 {
		t.Fatalf("Expect to latestvalidheight is 0 but get %+v", beaconPoolTest.latestValidHeight)
	}
	beaconPoolTest.validPool = []*blockchain.BeaconBlock{}
	beaconPoolTest.updateLatestBeaconState()
	if beaconPoolTest.latestValidHeight != 0 {
		t.Fatalf("Expect to latestvalidheight is 0 but get %+v", beaconPoolTest.latestValidHeight)
	}
}
func TestBeaconPoolRemoveBlock(t *testing.T) {
	InitBeaconPoolTest(pb)
	beaconPoolTest.validPool = append(beaconPoolTest.validPool, beaconBlock2)
	beaconPoolTest.validPool = append(beaconPoolTest.validPool, beaconBlock3)
	beaconPoolTest.validPool = append(beaconPoolTest.validPool, beaconBlock4)
	beaconPoolTest.validPool = append(beaconPoolTest.validPool, beaconBlock5)
	beaconPoolTest.validPool = append(beaconPoolTest.validPool, beaconBlock6)
	// remove VALID block according to latestblockheight input
	// also update latest valid height after call this function
	beaconPoolTest.removeBlock(4)
	if len(beaconPoolTest.validPool) != 2 {
		t.Fatalf("Expect to get only 2 block left but got %+v", len(beaconPoolTest.validPool))
	}
	if beaconPoolTest.latestValidHeight != 6 {
		t.Fatalf("Expect to latestvalidheight is 6 but get %+v", beaconPoolTest.latestValidHeight)
	}
	beaconPoolTest.removeBlock(6)
	if len(beaconPoolTest.validPool) != 0 {
		t.Fatalf("Expect to have NO block left but got %+v", len(beaconPoolTest.validPool))
	}
	// because no block left in valid pool and blockchain state is 0 so latest valid state should be 0
	if beaconPoolTest.latestValidHeight != 0 {
		t.Fatalf("Expect to latestvalidheight is 0 but get %+v", beaconPoolTest.latestValidHeight)
	}
}
func TestBeaconPoolCleanOldBlock(t *testing.T) {
	InitBeaconPoolTest(pb)
	if len(beaconPoolTest.pendingPool) != 0 {
		t.Fatalf("Expected number of block 0 in pending pool but get %+v", len(beaconPoolTest.pendingPool))
	}
	if len(beaconPoolTest.conflictedPool) != 0 {
		t.Fatalf("Expected number of block 0 in pending pool but get %+v", len(beaconPoolTest.conflictedPool))
	}
	// clean OLD block in Pending and Conflict pool
	// old block in pending pool has height < latestvalidheight
	// old block in conflicted pool has height < latestvalidheight - 2
	beaconPoolTest.pendingPool[beaconBlock2.Header.Height] = beaconBlock2
	beaconPoolTest.pendingPool[beaconBlock3.Header.Height] = beaconBlock3
	beaconPoolTest.conflictedPool[beaconBlock3Forked.Header.Hash()] = beaconBlock3Forked
	beaconPoolTest.pendingPool[beaconBlock4.Header.Height] = beaconBlock4
	beaconPoolTest.pendingPool[beaconBlock5.Header.Height] = beaconBlock5
	beaconPoolTest.pendingPool[beaconBlock6.Header.Height] = beaconBlock6
	if len(beaconPoolTest.pendingPool) != 5 {
		t.Fatalf("Expected number of block 5 in pending pool but get %+v", len(beaconPoolTest.pendingPool))
	}
	if len(beaconPoolTest.conflictedPool) != 1 {
		t.Fatalf("Expected number of block 1 in pending pool but get %+v", len(beaconPoolTest.conflictedPool))
	}
	beaconPoolTest.CleanOldBlock(2)
	if len(beaconPoolTest.pendingPool) != 4 {
		t.Fatalf("Expected number of block 4 in pending pool but get %+v", len(beaconPoolTest.pendingPool))
	}
	if len(beaconPoolTest.conflictedPool) != 1 {
		t.Fatalf("Expected number of block 1 in pending pool but get %+v", len(beaconPoolTest.conflictedPool))
	}
	beaconPoolTest.CleanOldBlock(3)
	if len(beaconPoolTest.pendingPool) != 3 {
		t.Fatalf("Expected number of block 3 in pending pool but get %+v", len(beaconPoolTest.pendingPool))
	}
	if len(beaconPoolTest.conflictedPool) != 1 {
		t.Fatalf("Expected number of block 1 in pending pool but get %+v", len(beaconPoolTest.conflictedPool))
	}
	beaconPoolTest.CleanOldBlock(5)
	if len(beaconPoolTest.conflictedPool) != 1 {
		t.Fatalf("Expected number of block 1 in pending pool but get %+v", len(beaconPoolTest.conflictedPool))
	}
	beaconPoolTest.CleanOldBlock(6)
	if len(beaconPoolTest.pendingPool) != 0 {
		t.Fatalf("Expected number of block 0 in pending pool but get %+v", len(beaconPoolTest.pendingPool))
	}
	if len(beaconPoolTest.conflictedPool) != 0 {
		t.Fatalf("Expected number of block 0 in pending pool but get %+v", len(beaconPoolTest.conflictedPool))
	}
}
func TestBeaconPoolGetValidBlock(t *testing.T) {
	InitBeaconPoolTest(pb)
	beaconPoolTest.validPool = append(beaconPoolTest.validPool, beaconBlock2)
	beaconPoolTest.validPool = append(beaconPoolTest.validPool, beaconBlock3)
	beaconPoolTest.validPool = append(beaconPoolTest.validPool, beaconBlock4)
	beaconPoolTest.validPool = append(beaconPoolTest.validPool, beaconBlock5)
	beaconPoolTest.updateLatestBeaconState()
	beaconPoolTest.pendingPool[beaconBlock6.Header.Height] = beaconBlock6
	beaconPoolTest.pendingPool[beaconBlock7.Header.Height] = beaconBlock7
	// no role in committee then return only valid pool
	beaconPoolTest.RoleInCommittees = false
	validBlocks := beaconPoolTest.GetValidBlock()
	if len(validBlocks) != 4 {
		t.Fatalf("Expect return 4 blocks but get %+v", len(validBlocks))
	}
	for _, block := range validBlocks {
		if block.Header.Height == beaconBlock6.Header.Height {
			t.Fatal("Return block height 6 should not have block in pending pool")
		}
	}
	// if has role in beacon committee then return valid pool
	// IF VALID POOL IS EMPTY ONLY return one block from pending pool if condition is match
	// - Condition: block with height = latestvalidheight + 1 (if present) in pending poool
	beaconPoolTest.RoleInCommittees = true
	// if beacon pool in committee but valid block list not empty then return NO block from pending pool
	validAnd0PendingBlocks := beaconPoolTest.GetValidBlock()
	if len(validAnd0PendingBlocks) != 4 {
		t.Fatalf("Expect return 4 blocks but get %+v", len(validAnd0PendingBlocks))
	}
	for _, block := range validAnd0PendingBlocks {
		if block.Header.Height == beaconBlock6.Header.Height {
			t.Fatal("Return block height 6 should not have block in pending pool")
		}
		if block.Header.Height == beaconBlock7.Header.Height {
			t.Fatal("Return block height 7 should not have block in pending pool")
		}
	}
	// if beacon pool in committee but valid block list IS EMPTY
	// then return ONLY 1 block from pending pool that match condition (see above)
	beaconPoolTest.validPool = []*blockchain.BeaconBlock{}
	oneValidFromPendingBlocks := beaconPoolTest.GetValidBlock()
	if len(oneValidFromPendingBlocks) != 1 {
		t.Fatalf("Expect return 1 blocks but get %+v", len(oneValidFromPendingBlocks))
	}
	if oneValidFromPendingBlocks[0].Header.Height != beaconBlock6.Header.Height {
		t.Fatalf("Expect return block height 6 but get %+v", oneValidFromPendingBlocks[0].Header.Height)
	}
	if oneValidFromPendingBlocks[0].Header.Height == beaconBlock7.Header.Height {
		t.Fatalf("DONT expect return block height 7 but get %+v", oneValidFromPendingBlocks[0].Header.Height)
	}
}
