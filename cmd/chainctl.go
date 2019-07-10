package main

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/mempool"
	"github.com/incognitochain/incognito-chain/pubsub"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
)

//default chainDataDir is data/testnet/block
func BackupShardChain(shardID byte, chainDataDir string, outDatadir string, fileName string, testNet bool) error {
	if fileName == "" {
		fileName = "export-incognito-shard-" + strconv.Itoa(int(shardID))
	}
	if outDatadir == "" {
		outDatadir = "./"
	}
	file := filepath.Join(outDatadir,fileName)
	fileHandler, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer fileHandler.Close()
	var writer io.Writer = fileHandler
	bc, err := makeBlockChain(chainDataDir, testNet)
	if err != nil {
		return err
	}
	if err := bc.BackupShardChain(writer, shardID); err != nil {
		return err
	}
	log.Printf("Backup Shard %+v Chain, file %+v", shardID, file)
	return nil
}
func BackupBeaconChain(chainDataDir string, outDatadir string, fileName string, testNet bool) error {
	if fileName == "" {
		fileName = "export-incognito-beacon"
	}
	if outDatadir == "" {
		outDatadir = "./"
	}
	file := filepath.Join(outDatadir,fileName)
	fileHandler, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer fileHandler.Close()
	var writer io.Writer = fileHandler
	bc, err := makeBlockChain(chainDataDir, testNet)
	if err != nil {
		return err
	}
	if err := bc.BackupBeaconChain(writer); err != nil {
		return err
	}
	log.Printf("Backup Beacon Chain, file %+v", file)
	return nil
}
func makeBlockChain(databaseDir string, testNet bool) (*blockchain.BlockChain, error) {
	blockchain.Logger.Init(common.NewBackend(nil).Logger("ChainCMD", true))
	mempool.Logger.Init(common.NewBackend(nil).Logger("ChainCMD", true))
	db, err := database.Open("leveldb", filepath.Join(databaseDir))
	if err != nil {
		return nil, err
	}
	log.Printf("Open leveldb at %+v successfully", filepath.Join(databaseDir))
	bc := blockchain.NewBlockChain(&blockchain.Config{}, false)
	var bcParams *blockchain.Params
	if testNet {
		bcParams = &blockchain.ChainTestParam
	} else {
		bcParams = &blockchain.ChainMainParam
	}
	crossShardPoolMap := make(map[byte]blockchain.CrossShardPool)
	shardPoolMap := make(map[byte]blockchain.ShardPool)
	for i := 0; i< 255; i++ {
		shardID := byte(i)
		crossShardPoolMap[shardID] = mempool.GetCrossShardPool(shardID)
		shardPoolMap[shardID] = mempool.GetShardPool(shardID)
	}
	pb := pubsub.NewPubSubManager()
	txPool := &mempool.TxPool{}
	txPool.Init(&mempool.Config{
		PubSubManager: pb,
		DataBase: db,
		BlockChain: bc,
		ChainParams: bcParams,
	})
	err = bc.Init(&blockchain.Config{
		ChainParams: bcParams,
		DataBase:          db,
		BeaconPool:        mempool.GetBeaconPool(),
		ShardToBeaconPool: mempool.GetShardToBeaconPool(),
		PubSubManager:     pb,
		CrossShardPool:  crossShardPoolMap,
		ShardPool: shardPoolMap,
		TxPool: txPool,
	})
	if err != nil {
		return nil, err
	}
	return bc, nil
}
func RestoreShardChain(shardID byte, chainDataDir string, filename string, testNet bool) error {
	// Watch for Ctrl-C while the import is running.
	// If a signal is received, the import will stop at the next batch.
	interrupt := make(chan os.Signal, 1)
	stop := make(chan struct{})
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(interrupt)
	defer close(interrupt)
	go func() {
		if _, ok := <-interrupt; ok {
			log.Println("Interrupted during import, stopping at next block")
		}
		close(stop)
	}()
	checkInterrupt := func() bool {
		select {
		case <-stop:
			return true
		default:
			return false
		}
	}
	log.Println("Importing blockchain", "file", filename)
	// Open the file handle and potentially unwrap the gzip stream
	fileHanlder, err := os.Open(filename)
	if err != nil {
		return err
	}
	log.Println(fileHanlder.Name())
	defer fileHanlder.Close()
	var reader io.Reader = fileHanlder
	if err != nil {
		return err
	}
	bc, err := makeBlockChain(chainDataDir, testNet)
	if err != nil {
		return err
	}
	for {
		numberOfByteToRead := make([]byte, 8)
		_, err := reader.Read(numberOfByteToRead)
		if err == io.EOF {
			break
		} else {
			if err != nil {
				return err
			}
		}
		blockLength, err := blockchain.GetNumberOfByteToRead(numberOfByteToRead)
		if err != nil {
			return err
		}
		blockBytes := make([]byte, blockLength)
		_, err = reader.Read(blockBytes)
		if err == io.EOF {
			break
		} else {
			if err != nil {
				return err
			}
		}
		block := &blockchain.ShardBlock{}
		err = block.UnmarshalJSON(blockBytes)
		if err != nil {
			return err
		}
		log.Printf("Restore Shard %+v Block %+v \n", block.Header.ShardID, block.Header.Height)
		err = bc.InsertShardBlock(block, true)
		if bcErr, ok := err.(*blockchain.BlockChainError); ok {
			if bcErr.Code == blockchain.ErrCodeMessage[blockchain.DuplicateBlockErr].Code {
				continue
			}
		}
		if err != nil {
			return err
		}
		// check interupt whenever finish insert 1 block
		checkInterrupt()
	}
	log.Printf("Restore Shard %+v Chain Successfully", shardID)
	return nil
}
func RestoreBeaconChain(chainDataDir string, filename string, testNet bool) error {
	// Watch for Ctrl-C while the import is running.
	// If a signal is received, the import will stop at the next batch.
	interrupt := make(chan os.Signal, 1)
	stop := make(chan struct{})
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(interrupt)
	defer close(interrupt)
	go func() {
		if _, ok := <-interrupt; ok {
			log.Println("Interrupted during import, stopping at next block")
		}
		close(stop)
	}()
	checkInterrupt := func() bool {
		select {
		case <-stop:
			return true
		default:
			return false
		}
	}
	log.Println("Importing blockchain", "file", filename)
	// Open the file handle and potentially unwrap the gzip stream
	fh, err := os.Open(filename)
	if err != nil {
		return err
	}
	log.Println(fh.Name())
	defer fh.Close()
	var reader io.Reader = fh
	if err != nil {
		return err
	}
	bc, err := makeBlockChain(chainDataDir, testNet)
	if err != nil {
		return err
	}
	for {
		numberOfByteToRead := make([]byte, 8)
		_, err := reader.Read(numberOfByteToRead)
		if err == io.EOF {
			break
		} else {
			if err != nil {
				return err
			}
		}
		blockLength, err := blockchain.GetNumberOfByteToRead(numberOfByteToRead)
		if err != nil {
			return err
		}
		blockBytes := make([]byte, blockLength)
		_, err = reader.Read(blockBytes)
		if err == io.EOF {
			break
		} else {
			if err != nil {
				return err
			}
		}
		block := &blockchain.BeaconBlock{}
		err = block.UnmarshalJSON(blockBytes)
		if err != nil {
			return err
		}
		log.Printf("Restore Block %+v \n", block.Header.Height)
		if block.Header.Height == 1 {
			continue
		}
		err = bc.InsertBeaconBlock(block, true)
		if bcErr, ok := err.(*blockchain.BlockChainError); ok {
			if bcErr.Code == blockchain.ErrCodeMessage[blockchain.DuplicateBlockErr].Code {
				continue
			}
		}
		if err != nil {
			return err
		}
		// check interupt whenever finish insert 1 block
		checkInterrupt()
	}
	log.Println("Restore Beacon Chain Successfully")
	return nil
}
