package blockchain

import (
	"context"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/config"
	"log"
	"sync/atomic"

	// "strconv"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
)

type PreFetchContext struct {
	context.Context
	NumTxRemain int64
	MaxTime     time.Duration
	MaxSize     uint64
	running     bool
	cancelFunc  context.CancelFunc
}

type any = interface{}

func (s *PreFetchContext) DecreaseNumTXRemain() {
	atomic.AddInt64(&s.NumTxRemain, -1)
}

func (s *PreFetchContext) GetMaxTime() time.Duration {
	return s.MaxTime
}
func (s *PreFetchContext) GetMaxSize() uint64 {
	return s.MaxSize
}
func (s *PreFetchContext) IsRunning() bool {
	return s.running
}
func (s *PreFetchContext) GetNumTxRemain() int64 {
	return atomic.LoadInt64(&s.NumTxRemain)
}

func NewPreFetchContext() *PreFetchContext {
	return &PreFetchContext{context.Background(), 0, 0, 4096, false, nil}
}

type PreFetchTx struct {
	BlockChain   *BlockChain
	BestView     *ShardBestState
	BeaconBlocks []*types.BeaconBlock
	ResponseTxs  map[common.Hash]metadata.Transaction
	CollectedTxs map[common.Hash]metadata.Transaction
	Error        string
	WgStop       *sync.WaitGroup
	Ctx          *PreFetchContext
}

//get response tx and mempool tx
func (s *PreFetchTx) GetTxForBlockProducing() []metadata.Transaction {
	txs := []metadata.Transaction{}
	for _, tx := range s.ResponseTxs {
		txs = append(txs, tx)
	}
	for _, tx := range s.CollectedTxs {
		txs = append(txs, tx)
	}
	return txs
}

//call whenever there is new view
func (s *PreFetchTx) Reset(view *ShardBestState) {
	s.Stop()
	s.BeaconBlocks = []*types.BeaconBlock{}
	s.CollectedTxs = make(map[common.Hash]metadata.Transaction)
	s.ResponseTxs = make(map[common.Hash]metadata.Transaction)
	s.BestView = view

}

//call when start propose block
func (s *PreFetchTx) Stop() {
	if s.Ctx != nil && s.Ctx.cancelFunc != nil {
		s.Ctx.cancelFunc()
	}
	s.WgStop.Wait()
	s.Ctx = NewPreFetchContext()

}

//call when next timeslot is proposer => prepare tx
func (s *PreFetchTx) Start(curView *ShardBestState) {
	if s.BestView.BestBlock.Hash().String() != curView.BestBlock.Hash().String() {
		s.Reset(curView)
	}

	if s.Ctx.running {
		log.Println("debugprefetch: pre fetch already running")
		return
	}
	Logger.log.Info("debugprefetch: running...")
	s.Reset(curView)
	s.Ctx.running = true

	s.Ctx.Context, s.Ctx.cancelFunc = context.WithDeadline(s.Ctx.Context, time.Now().Add(time.Second*time.Duration(common.TIMESLOT)))
	currentCtx := s.Ctx

	blockChain := s.BestView.blockChain
	shardID := curView.ShardID
	bView, err := blockChain.GetBeaconViewStateDataFromBlockHash(curView.BestBeaconHash, true, false, false)
	if err != nil {
		Logger.log.Info("debugprefetch: cannot dinf beacon view", curView.BestBeaconHash.String())
		return
	}

	numTxRemain := curView.MaxTxsPerBlockRemainder
	if curView.BestBlock.GetVersion() >= types.INSTANT_FINALITY_VERSION_V2 {
		if numTxRemain > int64(config.Param().TransactionInBlockParam.Upper) {
			numTxRemain = int64(config.Param().TransactionInBlockParam.Upper)
		}
	}
	s.Ctx.NumTxRemain = numTxRemain

	go func() {
		s.WgStop.Add(1)
		defer s.WgStop.Done()

		tempPrivateKey := blockChain.config.BlockGen.createTempKeyset()
		for {
			time.Sleep(time.Millisecond * 100)
			select {
			case <-currentCtx.Context.Done():
				Logger.log.Info("debugprefetch: done get response from beacon block", len(s.BeaconBlocks), len(s.ResponseTxs))
				return
			default:
			}
			getBeaconFinalHeightForProcess := func() uint64 {
				view := blockChain.BeaconChain.GetFinalView().(*BeaconBestState)
				height := view.GetHeight()
				if height > MAX_BEACON_BLOCK+curView.BeaconHeight {
					height = curView.BeaconHeight + MAX_BEACON_BLOCK
				}
				return height
			}
			beaconProcessHeight := getBeaconFinalHeightForProcess()
			beaconStartHeight := curView.BeaconHeight + 1
			if len(s.BeaconBlocks) > 0 {
				beaconStartHeight = s.BeaconBlocks[len(s.BeaconBlocks)-1].GetHeight() + 1
			}

			if s.BestView.CommitteeStateVersion() == committeestate.STAKING_FLOW_V2 {
				if beaconProcessHeight > config.Param().ConsensusParam.StakingFlowV3Height {
					beaconProcessHeight = config.Param().ConsensusParam.StakingFlowV3Height
				}
			}

			beaconBlocks, err := FetchBeaconBlockFromHeight(s.BestView.blockChain, beaconStartHeight, beaconProcessHeight)
			//fmt.Println("debugprefetch: get beacon block", beaconStartHeight, beaconProcessHeight, len(beaconBlocks))
			if err != nil {
				fmt.Println(err)
				return
			}
			if len(beaconBlocks) > 0 {
				for _, bBlock := range beaconBlocks {
					responseTxsBeacon, errInstructions, err := blockChain.config.BlockGen.buildResponseTxsFromBeaconInstructions(curView, beaconBlocks, &tempPrivateKey, shardID)
					if err != nil {
						s.Error = err.Error()
						Logger.log.Error("Error during get response tx from beacon instruction", err)
						break
					}
					if len(errInstructions) > 0 {
						s.Error = fmt.Sprintf("Error instruction: %+v", errInstructions)
						Logger.log.Error("List error instructions, which can not create tx", errInstructions)
					}
					s.BeaconBlocks = append(s.BeaconBlocks, bBlock)
					for _, tx := range responseTxsBeacon {
						s.ResponseTxs[*tx.Hash()] = tx
					}
					s.Ctx.NumTxRemain -= int64(len(responseTxsBeacon))
					if s.Ctx.NumTxRemain < 0 {
						break
					}
				}
			}
		}
	}()

	go func() {
		s.WgStop.Add(1)
		defer func() {
			Logger.log.Info("debugprefetch: done get tx from mempool", len(s.CollectedTxs))
			s.WgStop.Done()
		}()

		if !blockChain.config.usingNewPool {
			st := time.Now()
			//get transaction until context cancel
			s.CollectedTxs, _ = blockChain.config.BlockGen.getPendingTransaction(
				curView.ShardID,
				currentCtx,
				curView.BeaconHeight,
				curView,
			)
			Logger.log.Infof("SHARD %v | Crawling %v txs for block %v cost %v", shardID, len(s.CollectedTxs), curView.ShardHeight+1, time.Since(st))
		} else {
			currentCtx.MaxTime = time.Second * time.Duration(common.TIMESLOT) * 4
			txsToAdd := blockChain.ShardChain[shardID].TxPool.GetTxsTranferForNewBlock(
				blockChain,
				curView,
				bView,
				currentCtx,
			)
			for _, tx := range txsToAdd {
				s.CollectedTxs[*tx.Hash()] = tx
			}
		}
	}()

	return
}

type BlockGenerator struct {
	// blockpool   BlockPool
	txPool      TxPool
	syncker     Syncker
	chain       *BlockChain
	CQuit       chan struct{}
	CPendingTxs <-chan metadata.Transaction
	CRemovedTxs <-chan metadata.Transaction
	PendingTxs  map[common.Hash]metadata.Transaction
	mtx         sync.RWMutex
}

func NewBlockGenerator(txPool TxPool, chain *BlockChain, syncker Syncker, cPendingTxs chan metadata.Transaction, cRemovedTxs chan metadata.Transaction) (*BlockGenerator, error) {
	return &BlockGenerator{
		txPool:      txPool,
		syncker:     syncker,
		chain:       chain,
		PendingTxs:  make(map[common.Hash]metadata.Transaction),
		CPendingTxs: cPendingTxs,
		CRemovedTxs: cRemovedTxs,
	}, nil
}

func (blockGenerator *BlockGenerator) Start(cQuit chan struct{}) {
	Logger.log.Critical("Block Gen is starting")
	for w := 0; w < WorkerNumber; w++ {
		go blockGenerator.AddTransactionV2Worker(blockGenerator.CPendingTxs)
	}
	for w := 0; w < WorkerNumber; w++ {
		go blockGenerator.RemoveTransactionV2Worker(blockGenerator.CRemovedTxs)
	}
	for {
		select {
		case <-cQuit:
			return
		}
	}
}
func (blockGenerator *BlockGenerator) AddTransactionV2(tx metadata.Transaction) {
	blockGenerator.mtx.Lock()
	defer blockGenerator.mtx.Unlock()
	blockGenerator.PendingTxs[*tx.Hash()] = tx
}
func (blockGenerator *BlockGenerator) AddTransactionV2Worker(cPendingTx <-chan metadata.Transaction) {
	for tx := range cPendingTx {
		blockGenerator.AddTransactionV2(tx)
		time.Sleep(time.Nanosecond)
	}
}
func (blockGenerator *BlockGenerator) RemoveTransactionV2(tx metadata.Transaction) {
	blockGenerator.mtx.Lock()
	defer blockGenerator.mtx.Unlock()
	delete(blockGenerator.PendingTxs, *tx.Hash())
}
func (blockGenerator *BlockGenerator) RemoveTransactionV2Worker(cRemoveTx <-chan metadata.Transaction) {
	for tx := range cRemoveTx {
		blockGenerator.RemoveTransactionV2(tx)
		time.Sleep(time.Nanosecond)
	}
}
func (blockGenerator *BlockGenerator) GetPendingTxsV2(shardID byte) []metadata.Transaction {
	blockGenerator.mtx.Lock()
	defer blockGenerator.mtx.Unlock()
	pendingTxs := []metadata.Transaction{}
	for _, tx := range blockGenerator.PendingTxs {
		txShardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
		if shardID != 255 && txShardID != shardID {
			continue
		}
		pendingTxs = append(pendingTxs, tx)
	}
	return pendingTxs
}
