package blockchain

import (
	"context"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/config"

	// "strconv"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
)

type PreFetchTx struct {
	BlockChain   *BlockChain
	BestView     *ShardBestState
	BeaconBlocks []*types.BeaconBlock
	ResponseTxs  map[common.Hash]metadata.Transaction
	CollectedTxs map[common.Hash]metadata.Transaction
	Error        string
	WgStop       *sync.WaitGroup
	Ctx          context.Context
	CancelFunc   context.CancelFunc
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
	if s.BestView == nil || s.BestView.BestBlock.Hash().String() != view.BestBlock.Hash().String() {
		s.BeaconBlocks = []*types.BeaconBlock{}
		s.CollectedTxs = make(map[common.Hash]metadata.Transaction)
		s.ResponseTxs = make(map[common.Hash]metadata.Transaction)
		s.BestView = view
		if s.CancelFunc != nil {
			s.CancelFunc()
		}
		s.Ctx, s.CancelFunc = context.WithCancel(context.Background())
	}
}

//call when start propose block
func (s *PreFetchTx) Stop() {
	if s.CancelFunc != nil {
		s.CancelFunc()
	}
	s.WgStop.Wait()
	context.WithValue(s.Ctx, "Running", false)
}

//call when next timeslot is proposer => prepare tx
func (s *PreFetchTx) Start() (context.Context, context.CancelFunc) {
	if s.Ctx.Value("Running") != nil && s.Ctx.Value("Running").(bool) {
		fmt.Println("debugprefetch: pre fetch already running")
		return s.Ctx, s.CancelFunc
	}
	s.Ctx, s.CancelFunc = context.WithCancel(context.Background())
	s.Ctx = context.WithValue(s.Ctx, "Running", true)
	s.Ctx, s.CancelFunc = context.WithDeadline(s.Ctx, time.Now().Add(time.Second*time.Duration(common.TIMESLOT)))

	defer func() {
		s.Ctx = context.WithValue(s.Ctx, "Running", false)
	}()

	blockChain := s.BestView.blockChain
	curView := s.BestView
	shardID := curView.ShardID

	totalTxsReminder := config.Param().TransactionInBlockParam.Upper
	s.Ctx = context.WithValue(s.Ctx, "maxTXs", totalTxsReminder)
	currentCtx := s.Ctx

	go func() {
		s.WgStop.Add(1)
		defer s.WgStop.Done()

		tempPrivateKey := blockChain.config.BlockGen.createTempKeyset()
		for {
			time.Sleep(time.Millisecond * 100)
			select {
			case <-currentCtx.Done():
				fmt.Println("debugprefetch: done get response from beacon block", len(s.BeaconBlocks))
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
			beaconBlocks, err := FetchBeaconBlockFromHeight(s.BestView.blockChain, beaconStartHeight, beaconProcessHeight)
			fmt.Println("debugprefetch: get beacon block", beaconStartHeight, beaconProcessHeight, len(beaconBlocks))
			if err != nil {
				fmt.Println(err)
				return
			}
			if len(beaconBlocks) > 0 {
				for _, bBlock := range beaconBlocks {
					responseTxsBeacon, errInstructions, err := blockChain.config.BlockGen.buildResponseTxsFromBeaconInstructions(curView, beaconBlocks, &tempPrivateKey, shardID)
					s.BeaconBlocks = append(s.BeaconBlocks, bBlock)
					if err != nil {
						s.Error = err.Error()
						Logger.log.Error("Error during get response tx from beacon instruction", err)
						break
					}
					if len(errInstructions) > 0 {
						s.Error = fmt.Sprintf("Error instruction: %+v", errInstructions)
						Logger.log.Error("List error instructions, which can not create tx", errInstructions)
					}
					for _, tx := range responseTxsBeacon {
						s.ResponseTxs[*tx.Hash()] = tx
					}
				}
			}
		}
	}()

	go func() {
		s.WgStop.Add(1)
		defer func() {
			fmt.Println("debugprefetch: done get tx from mempool")
			s.WgStop.Done()
		}()

		if !blockChain.config.usingNewPool {
			st := time.Now()
			//get transaction until context cancel
			s.CollectedTxs, _ = blockChain.config.BlockGen.streamPendingTransaction(
				curView.ShardID,
				currentCtx,
				curView.BeaconHeight,
				curView,
			)
			Logger.log.Infof("SHARD %v | Crawling %v txs for block %v cost %v", shardID, len(s.CollectedTxs), curView.ShardHeight+1, time.Since(st))
		}
	}()

	return s.Ctx, s.CancelFunc
}

type BlockGenerator struct {
	// blockpool   BlockPool
	txPool       TxPool
	syncker      Syncker
	chain        *BlockChain
	CQuit        chan struct{}
	CPendingTxs  <-chan metadata.Transaction
	CRemovedTxs  <-chan metadata.Transaction
	PendingTxs   map[common.Hash]metadata.Transaction
	CollectedTxs []metadata.Transaction
	mtx          sync.RWMutex
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
