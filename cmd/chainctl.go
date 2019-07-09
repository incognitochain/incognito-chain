package main

import (
	"compress/gzip"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/database"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
)
func BackupShardChain(shardID byte, chainDataDir string, outDatadir string) error {
	fileName := "export-incognito-shard-"+strconv.Itoa(int(shardID))+".gz"
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
func makeBlockChain (databaseDir string) (*blockchain.BlockChain, error) {
	db, err := database.Open("leveldb", filepath.Join(databaseDir))
	if err != nil {
		return nil, err
	}
	log.Printf("Open leveldb at %+v successfully", filepath.Join(databaseDir))
	bc := blockchain.NewBlockChain(&blockchain.Config{
		DataBase: db,
	}, false)
	return bc, nil
}
