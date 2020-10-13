package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/simulation/mock"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
	"github.com/incognitochain/incognito-chain/memcache"
	"github.com/incognitochain/incognito-chain/metadata"
	bnbrelaying "github.com/incognitochain/incognito-chain/relaying/bnb"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
)

type simInstance struct {
	dbDir       string
	param       *blockchain.Params
	bc          *blockchain.BlockChain
	cs          *mock.Consensus
	txpool      *mock.TxPool
	temppool    *mock.TxPool
	btcrd       *mock.BTCRandom
	sync        *mock.Syncker
	server      *mock.Server
	cPendingTxs chan metadata.Transaction
	cRemovedTxs chan metadata.Transaction
	cQuit       chan struct{}
}

func main() {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	initLogRotator(filepath.Join(path, "log.log"))
	shardNumber := 2
	activeNetParams := &blockchain.ChainTest2Param
	cs := mock.Consensus{}
	txpool := mock.TxPool{}
	temppool := mock.TxPool{}
	btcrd := mock.BTCRandom{} // use mock for now
	sync := mock.Syncker{}
	server := mock.Server{}
	fees := make(map[byte]blockchain.FeeEstimator)
	bc := blockchain.BlockChain{}
	for i := byte(0); i < byte(shardNumber); i++ {
		fees[i] = &mock.Fee{}
	}
	cPendingTxs := make(chan metadata.Transaction, 500)
	cRemovedTxs := make(chan metadata.Transaction, 500)
	cQuit := make(chan struct{})
	blockgen, err := blockchain.NewBlockGenerator(&txpool, &bc, &sync, cPendingTxs, cRemovedTxs)
	if err != nil {
		panic(err)
	}

	db, err := incdb.OpenMultipleDB("leveldb", filepath.Join("./testdb/", "database"))
	// Create db and use it.
	if err != nil {
		panic(err)
	}

	btcChain, err := getBTCRelayingChain(
		activeNetParams.BTCRelayingHeaderChainID,
		"btcchain",
	)
	if err != nil {
		panic(err)
	}
	bnbChainState, err := getBNBRelayingChainState(activeNetParams.BNBRelayingHeaderChainID)
	if err != nil {
		panic(err)
	}

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
		FeeEstimator:    make(map[byte]blockchain.FeeEstimator),
		RandomClient:    &btcrd,
		ConsensusEngine: &cs,
		GenesisParams:   blockchain.GenesisParam,
	})
	if err != nil {
		panic(err)
	}
	bc.InitChannelBlockchain(cRemovedTxs)

	go func() {
		for {
			<-cRemovedTxs
		}
	}()
	go blockgen.Start(cQuit)
}

func getBTCRelayingChain(btcRelayingChainID string, btcDataFolderName string) (*btcrelaying.BlockChain, error) {
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
		filepath.Join("./", btcDataFolderName),
		relayingChainParams[btcRelayingChainID],
		relayingChainGenesisBlkHeight[btcRelayingChainID],
	)
}

func getBNBRelayingChainState(bnbRelayingChainID string) (*bnbrelaying.BNBChainState, error) {
	bnbChainState := new(bnbrelaying.BNBChainState)
	err := bnbChainState.LoadBNBChainState(
		filepath.Join("./testdb/", "bnbrelayingv3"),
		bnbRelayingChainID,
	)
	if err != nil {
		log.Printf("Error getBNBRelayingChainState: %v\n", err)
		return nil, err
	}
	return bnbChainState, nil
}

func newSimInstance() {

}
