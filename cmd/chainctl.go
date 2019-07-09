package main

import (
	"compress/gzip"
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

func BackupShardChain(shardID byte, chainDataDir string, outDatadir string) error {
	fileName := "export-incognito-shard-" + strconv.Itoa(int(shardID)) + ".gz"
	file := fileName
	fileHandler, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer fileHandler.Close()
	var writer io.Writer
	writer = gzip.NewWriter(fileHandler)
	defer writer.(*gzip.Writer).Close()
	bc, err := makeBlockChain(chainDataDir)
	if err != nil {
		return err
	}
	if err := bc.BackupShardChain(writer, shardID); err != nil {
		return err
	}
	log.Printf("Backup Shard %+v Chain, file %+v", shardID, file)
	return nil
}
func BackupBeaconChain(chainDataDir string, outDatadir string) error {
	fileName := "export-incognito-beacon"
	file := fileName
	fileHandler, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer fileHandler.Close()
	var writer io.Writer = fileHandler
	bc, err := makeBlockChain(chainDataDir)
	if err != nil {
		return err
	}
	if err := bc.BackupBeaconChain(writer); err != nil {
		return err
	}
	log.Printf("Backup Beacon Chain, file %+v", file)
	return nil
}
func makeBlockChain(databaseDir string) (*blockchain.BlockChain, error) {
	db, err := database.Open("leveldb", filepath.Join(databaseDir))
	if err != nil {
		return nil, err
	}
	log.Printf("Open leveldb at %+v successfully", filepath.Join(databaseDir))
	bc := blockchain.NewBlockChain(&blockchain.Config{
	
	}, false)
	crossShardPoolMap := make(map[byte]blockchain.CrossShardPool)
	for i := 0; i< 255; i++ {
		shardID := byte(i)
		crossShardPoolMap[shardID] = mempool.GetCrossShardPool(shardID)
	}
	bc.Init(&blockchain.Config{
		ChainParams: &blockchain.ChainTestParam,
		DataBase:          db,
		BeaconPool:        mempool.GetBeaconPool(),
		ShardToBeaconPool: mempool.GetShardToBeaconPool(),
		PubSubManager:     pubsub.NewPubSubManager(),
		CrossShardPool:  crossShardPoolMap,
	})
	blockchain.Logger.Init(common.NewBackend(nil).Logger("ChainCMD", true))
	mempool.Logger.Init(common.NewBackend(nil).Logger("ChainCMD", true))
	return bc, nil
}
func RestoreShardChain(shardID byte, chainDataDir string, filename string) error {
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
	reader, err := gzip.NewReader(fh)
	if err != nil {
		return err
	}
	bc, err := makeBlockChain(chainDataDir)
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
		log.Printf("Block length %+v", len(blockBytes))
		log.Printf("Block bytes", blockBytes)
		block := &blockchain.ShardBlock{}
		err = block.UnmarshalJSON(blockBytes)
		if err != nil {
			return err
		}
		log.Printf("Block %+v length %+v", block.Header.Height, len(blockBytes))
		log.Println(block)
		err = bc.ProcessStoreShardBlock(block)
		if err != nil {
			return err
		}
		// check interupt whenever finish insert 1 block
		checkInterrupt()
	}
	log.Printf("Restore Shard %+v Chain Successfully", shardID)
	return nil
}
func RestoreBeaconChain(chainDataDir string, filename string) error {
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
	bc, err := makeBlockChain(chainDataDir)
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
		counter, err := reader.Read(blockBytes)
		if err == io.EOF {
			break
		} else {
			if err != nil {
				return err
			}
		}
		log.Printf("numberOfByteToRead: %+v \n", numberOfByteToRead)
		log.Printf("blockLength: %+v \n", blockLength)
		log.Printf("counter of block: %+v \n", counter)
		log.Printf("Block length %+v", len(blockBytes))
		log.Printf("Block bytes", blockBytes)
		block := &blockchain.BeaconBlock{}
		err = block.UnmarshalJSON(blockBytes)
		if err != nil {
			return err
		}
		log.Printf("Block %+v length %+v", block.Header.Height, len(blockBytes))
		log.Println(block)
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
