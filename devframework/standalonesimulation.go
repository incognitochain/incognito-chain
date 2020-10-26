package devframework

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/devframework/mock"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
	"github.com/incognitochain/incognito-chain/memcache"
	"github.com/incognitochain/incognito-chain/mempool"
	"github.com/incognitochain/incognito-chain/metadata"
	bnbrelaying "github.com/incognitochain/incognito-chain/relaying/bnb"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
	"github.com/incognitochain/incognito-chain/rpcserver"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/transaction"
)

type Config struct {
	ShardNumber   int
	RoundInterval int
}

type Hook struct {
	Create     func(chain interface{}, doCreate func(time time.Time) (blk common.BlockInterface, err error))
	Validation func(chain interface{}, block common.BlockInterface, doValidation func(blk common.BlockInterface) error)
	Insert     func(chain interface{}, block common.BlockInterface, doInsert func(blk common.BlockInterface) error)
}
type SimulationEngine struct {
	config      Config
	simName     string
	timer       *TimeEngine
	param       *blockchain.Params
	bc          *blockchain.BlockChain
	cs          *mock.Consensus
	txpool      *mempool.TxPool
	temppool    *mempool.TxPool
	btcrd       *mock.BTCRandom
	sync        *mock.Syncker
	server      *mock.Server
	cPendingTxs chan metadata.Transaction
	cRemovedTxs chan metadata.Transaction
	rpcServer   *rpcserver.RpcServer
	cQuit       chan struct{}
}

func NewStandaloneSimulation(name string, config Config) *SimulationEngine {
	os.RemoveAll(name)
	sim := &SimulationEngine{
		config:  config,
		simName: name,
		timer:   NewTimeEngine(),
	}
	sim.init()
	time.Sleep(1 * time.Second)
	return sim
}

func (sim *SimulationEngine) init() {
	simName := sim.simName
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	initLogRotator(filepath.Join(path, "sim.log"))
	dbLogger.SetLevel(common.LevelTrace)
	blockchainLogger.SetLevel(common.LevelTrace)
	bridgeLogger.SetLevel(common.LevelTrace)
	rpcLogger.SetLevel(common.LevelTrace)
	rpcServiceLogger.SetLevel(common.LevelTrace)
	rpcServiceBridgeLogger.SetLevel(common.LevelTrace)
	transactionLogger.SetLevel(common.LevelTrace)
	privacyLogger.SetLevel(common.LevelTrace)
	mempoolLogger.SetLevel(common.LevelTrace)

	//read and setup key
	blockchain.ReadKey()
	blockchain.SetupParam()

	//init blockchain
	bc := blockchain.BlockChain{}

	activeNetParams := &blockchain.ChainTest2Param
	sim.timer.init(activeNetParams.GenesisBeaconBlock.Header.Timestamp + 10)
	activeNetParams.ActiveShards = sim.config.ShardNumber

	cs := mock.Consensus{}
	txpool := mempool.TxPool{}
	temppool := mempool.TxPool{}
	btcrd := mock.BTCRandom{} // use mock for now
	sync := mock.Syncker{
		Bc:                  &bc,
		LastCrossShardState: make(map[byte]map[byte]uint64),
	}
	server := mock.Server{}
	ps := mock.Pubsub{}
	fees := make(map[byte]*mempool.FeeEstimator)
	for i := byte(0); i < byte(activeNetParams.ActiveShards); i++ {
		fees[i] = mempool.NewFeeEstimator(
			mempool.DefaultEstimateFeeMaxRollback,
			mempool.DefaultEstimateFeeMinRegisteredBlocks,
			0)
	}
	cPendingTxs := make(chan metadata.Transaction, 500)
	cRemovedTxs := make(chan metadata.Transaction, 500)
	cQuit := make(chan struct{})
	blockgen, err := blockchain.NewBlockGenerator(&txpool, &bc, &sync, cPendingTxs, cRemovedTxs)
	if err != nil {
		panic(err)
	}

	listenFunc := net.Listen
	listener, err := listenFunc("tcp", "0.0.0.0:8000")
	if err != nil {
		panic(err)
	}
	rpcConfig := rpcserver.RpcServerConfig{
		HttpListenters: []net.Listener{listener},
		RPCMaxClients:  1,
		DisableAuth:    true,
		ChainParams:    activeNetParams,
		BlockChain:     &bc,
		Blockgen:       blockgen,
		TxMemPool:      &txpool,
	}
	rpcServer := &rpcserver.RpcServer{}

	db, err := incdb.OpenMultipleDB("leveldb", filepath.Join("./"+simName, "database"))
	// Create db and use it.
	if err != nil {
		panic(err)
	}

	btcChain, err := getBTCRelayingChain(activeNetParams.BTCRelayingHeaderChainID, "btcchain", simName)
	if err != nil {
		panic(err)
	}
	bnbChainState, err := getBNBRelayingChainState(activeNetParams.BNBRelayingHeaderChainID, simName)
	if err != nil {
		panic(err)
	}

	txpool.Init(&mempool.Config{
		ConsensusEngine: &cs,
		BlockChain:      &bc,
		DataBase:        db,
		ChainParams:     activeNetParams,
		FeeEstimator:    fees,
		TxLifeTime:      100,
		MaxTx:           1000,
		// DataBaseMempool:   dbmp,
		IsLoadFromMempool: false,
		PersistMempool:    false,
		RelayShards:       nil,
		PubSubManager:     &ps,
	})
	// serverObj.blockChain.AddTxPool(serverObj.memPool)
	txpool.InitChannelMempool(cPendingTxs, cRemovedTxs)

	temppool.Init(&mempool.Config{
		BlockChain:    &bc,
		DataBase:      db,
		ChainParams:   activeNetParams,
		FeeEstimator:  fees,
		MaxTx:         1000,
		PubSubManager: &ps,
	})
	txpool.IsBlockGenStarted = true
	go temppool.Start(cQuit)
	go txpool.Start(cQuit)

	err = bc.Init(&blockchain.Config{
		BTCChain:        btcChain,
		BNBChainState:   bnbChainState,
		ChainParams:     activeNetParams,
		DataBase:        db,
		MemCache:        memcache.New(),
		BlockGen:        blockgen,
		TxPool:          &txpool,
		TempTxPool:      &temppool,
		Server:          &server,
		Syncker:         &sync,
		PubSubManager:   &ps,
		FeeEstimator:    make(map[byte]blockchain.FeeEstimator),
		RandomClient:    &btcrd,
		ConsensusEngine: &cs,
		GenesisParams:   blockchain.GenesisParam,
	})
	if err != nil {
		panic(err)
	}
	bc.InitChannelBlockchain(cRemovedTxs)

	sim.param = activeNetParams
	sim.bc = &bc
	sim.cs = &cs
	sim.txpool = &txpool
	sim.temppool = &temppool
	sim.btcrd = &btcrd
	sim.sync = &sync
	sim.server = &server
	sim.cPendingTxs = cPendingTxs
	sim.cRemovedTxs = cRemovedTxs
	sim.rpcServer = rpcServer
	sim.cQuit = cQuit

	rpcServer.Init(&rpcConfig)
	go func() {
		for {
			select {
			case <-cQuit:
				return
			case <-cRemovedTxs:
			}
		}
	}()
	go sync.UpdateConfirmCrossShard()
	go blockgen.Start(cQuit)
	go rpcServer.Start()

}

func getBTCRelayingChain(btcRelayingChainID string, btcDataFolderName string, dataFolder string) (*btcrelaying.BlockChain, error) {
	relayingChainParams := map[string]*chaincfg.Params{
		blockchain.TestnetBTCChainID:  btcrelaying.GetTestNet3Params(),
		blockchain.Testnet2BTCChainID: btcrelaying.GetTestNet3ParamsForInc2(),
		blockchain.MainnetBTCChainID:  btcrelaying.GetMainNetParams(),
	}
	relayingChainGenesisBlkHeight := map[string]int32{
		blockchain.TestnetBTCChainID:  int32(1833130),
		blockchain.Testnet2BTCChainID: int32(1833130),
		blockchain.MainnetBTCChainID:  int32(634140),
	}
	return btcrelaying.GetChainV2(
		filepath.Join("./"+dataFolder, btcDataFolderName),
		relayingChainParams[btcRelayingChainID],
		relayingChainGenesisBlkHeight[btcRelayingChainID],
	)
}

func getBNBRelayingChainState(bnbRelayingChainID string, dataFolder string) (*bnbrelaying.BNBChainState, error) {
	bnbChainState := new(bnbrelaying.BNBChainState)
	err := bnbChainState.LoadBNBChainState(
		filepath.Join("./"+dataFolder, "bnbrelayingv3"),
		bnbRelayingChainID,
	)
	if err != nil {
		log.Printf("Error getBNBRelayingChainState: %v\n", err)
		return nil, err
	}
	return bnbChainState, nil
}

func (sim *SimulationEngine) Pause() {
	fmt.Print("Simulation pause! Press Enter to continue ...")
	var input string
	fmt.Scanln(&input)
}

//Auto generate block
func (sim *SimulationEngine) AutoGenerateBlock(chainID int, numBlk int) {
	for i := 0; i < numBlk; i++ {
		sim.GenerateBlock(-1, nil, false)
		for j := 0; j < sim.config.ShardNumber; j++ {
			sim.GenerateBlock(chainID, nil, false)
		}
		sim.ForwardToFuture()
	}
}

//life cycle of a block generation process:
//PreCreate -> PreValidation -> PreInsert ->
func (sim *SimulationEngine) GenerateBlock(chainID int, h *Hook, forwardTime bool) (res map[int]interface{}) {
	res = make(map[int]interface{})
	//beacon
	chain := sim.bc
	var block common.BlockInterface = nil
	var err error

	//Create
	if h != nil && h.Create != nil {
		h.Create(chain, func(time time.Time) (blk common.BlockInterface, err error) {
			if chainID == -1 {
				block, err = chain.BeaconChain.CreateNewBlock(2, "", 1, sim.timer.Now())
				if err != nil {
					block = nil
					return nil, err
				}
				block.(mock.BlockValidation).AddValidationField("test")
				return block, nil
			} else {
				block, err = chain.ShardChain[byte(chainID)].CreateNewBlock(2, "", 1, sim.timer.Now())
				if err != nil {
					return nil, err
				}
				block.(mock.BlockValidation).AddValidationField("test")
				return block, nil
			}
		})
	} else {
		if chainID == -1 {
			block, err = chain.BeaconChain.CreateNewBlock(2, "", 1, sim.timer.Now())
			if err != nil {
				block = nil
				fmt.Println("NewBlockError", err)
			}
			block.(mock.BlockValidation).AddValidationField("test")
		} else {
			block, err = chain.ShardChain[byte(chainID)].CreateNewBlock(2, "", 1, sim.timer.Now())
			if err != nil {
				block = nil
				fmt.Println("NewBlockError", err)
			}
			block.(mock.BlockValidation).AddValidationField("test")
		}
	}

	//Validation
	if h != nil && h.Validation != nil {
		h.Validation(chain, block, func(blk common.BlockInterface) (err error) {
			if blk == nil {
				return errors.New("No block for validation")
			}
			if chainID == -1 {
				err = chain.VerifyPreSignBeaconBlock(blk.(*blockchain.BeaconBlock), true)
				if err != nil {
					return err
				}
				return nil
			} else {
				err = chain.VerifyPreSignShardBlock(block.(*blockchain.ShardBlock), byte(chainID))
				if err != nil {
					return err
				}
				return nil
			}
		})
	} else {
		if block == nil {
			fmt.Println("VerifyBlockErr no block")
		} else {
			if chainID == -1 {
				err = chain.VerifyPreSignBeaconBlock(block.(*blockchain.BeaconBlock), true)
				if err != nil {
					fmt.Println("VerifyBlockErr", err)
				}
			} else {
				err = chain.VerifyPreSignShardBlock(block.(*blockchain.ShardBlock), byte(chainID))
				if err != nil {
					fmt.Println("VerifyBlockErr", err)
				}
			}
		}

	}

	//Insert
	if h != nil && h.Insert != nil {
		h.Insert(chain, block, func(blk common.BlockInterface) (err error) {
			if blk == nil {
				return errors.New("No block for insert")
			}
			if chainID == -1 {
				err = chain.InsertBeaconBlock(blk.(*blockchain.BeaconBlock), true)
				if err != nil {
					return err
				} else {
					res[chainID] = block
				}
				return
			} else {
				err = chain.InsertShardBlock(blk.(*blockchain.ShardBlock), true)
				if err != nil {
					return err
				} else {
					res[chainID] = block
				}
				return
			}
		})
	} else {
		if block == nil {
			fmt.Println("InsertBlkErr no block")
		} else {
			if chainID == -1 {
				err = chain.InsertBeaconBlock(block.(*blockchain.BeaconBlock), true)
				if err != nil {
					fmt.Println("InsertBlkErr", err)
				} else {
					res[chainID] = block
				}
			} else {
				err = chain.InsertShardBlock(block.(*blockchain.ShardBlock), true)
				if err != nil {
					fmt.Println("InsertBlkErr", err)
				} else {
					res[chainID] = block
				}

			}
		}
	}

	if forwardTime {
		sim.ForwardToFuture()
	}

	return res
}

//number of second we want simulation to forward
//default = round interval
func (sim *SimulationEngine) ForwardToFuture() {
	sim.timer.Forward(10)
}

func (sim *SimulationEngine) GenerateTxs(createTxs []GenerateTxParam) error {
	txsInject := []string{}
	for _, createTxMeta := range createTxs {
		tx, err := sim.CreateTx(createTxMeta.SenderPrK, createTxMeta.Receivers)
		if err != nil {
			return err
		}
		txsInject = append(txsInject, tx.Base58CheckData)
	}
	err := sim.InjectTxs(txsInject)
	if err != nil {
		return err
	}
	return nil
}

func (sim *SimulationEngine) CreateTx(senderPrk string, receivers map[string]int) (*jsonresult.CreateTransactionResult, error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createtransaction",
		"params":  []interface{}{senderPrk, receivers, 1, 1},
		"id":      1,
	})
	if err != nil {
		return nil, err
	}
	body, err := sendRequest(requestBody)
	if err != nil {
		return nil, err
	}
	txResp := struct {
		Result jsonresult.CreateTransactionResult
	}{}
	err = json.Unmarshal(body, &txResp)
	if err != nil {
		return nil, err
	}
	return &txResp.Result, nil
}

func (sim *SimulationEngine) InjectTxs(txsBase58 []string) error {
	for _, txB58Check := range txsBase58 {
		rawTxBytes, _, err := base58.Base58Check{}.Decode(txB58Check)
		if err != nil {
			return err
		}
		var tx transaction.Tx
		err = json.Unmarshal(rawTxBytes, &tx)
		if err != nil {
			return err
		}
		sim.cPendingTxs <- &tx
	}
	return nil
}

func (sim *SimulationEngine) GetBalance(privateKey string) (map[string]uint64, error) {
	tokenList := make(map[string]uint64)
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getbalancebyprivatekey",
		"params":  []interface{}{privateKey},
		"id":      1,
	})
	if err != nil {
		return nil, err
	}
	body, err := sendRequest(requestBody)
	if err != nil {
		return nil, err
	}
	txResp := struct {
		Result uint64
	}{}
	err = json.Unmarshal(body, &txResp)
	if err != nil {
		return nil, err
	}
	tokenList["PRV"] = txResp.Result

	requestBody2, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getlistprivacycustomtokenbalance",
		"params":  []interface{}{privateKey},
		"id":      1,
	})
	if err != nil {
		return nil, err
	}
	body2, err := sendRequest(requestBody2)
	if err != nil {
		return nil, err
	}
	txResp2 := struct {
		Result jsonresult.ListCustomTokenBalance
	}{}
	err = json.Unmarshal(body2, &txResp2)
	if err != nil {
		return nil, err
	}
	for _, token := range txResp2.Result.ListCustomTokenBalance {
		tokenList[token.Name] = token.Amount
	}
	return tokenList, nil
}

func (sim *SimulationEngine) GetBlockchain() *blockchain.BlockChain {
	return sim.bc
}

func sendRequest(requestBody []byte) ([]byte, error) {
	resp, err := http.Post("http://0.0.0.0:8000", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
