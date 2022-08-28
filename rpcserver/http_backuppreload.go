package rpcserver

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"strconv"
)

func (httpServer *HttpServer) handleGetLatestBackup(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	latest := httpServer.GetBlockchain().BackupManager.GetLastestBootstrap()
	b, _ := json.Marshal(latest)
	fmt.Println("GetLastestBootstrap", b)
	return latest, nil
}

type FileObject struct {
	Name string
	Size uint64
}

func (httpServer *HttpServer) handleGetBootstrapStateDB(conn net.Conn, params interface{}) {
	paramArray, ok := params.([]interface{})
	if !ok || len(paramArray) != 4 {
		return
	}

	checkpoint, ok := paramArray[0].(string)
	chainID, ok := paramArray[1].(float64)
	dbType, ok := paramArray[2].(string)
	blkHeight, ok := paramArray[3].(float64)

	checkPointFolder := httpServer.GetBlockchain().BackupManager.GetBackupReader(checkpoint, int(chainID))
	ff_fileId := uint64(0)
	if dbType == "blockKV" {
		checkPointFolder = path.Join(checkPointFolder, "blockstorage", "blockKV")
	} else if dbType == "block" {
		checkPointFolder = path.Join(checkPointFolder, "blockstorage")
		ff_fileId = httpServer.GetBlockchain().BackupManager.GetFileID(int(chainID), uint64(blkHeight))
	}

	_, err := conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\n\r\n"))
	if err != nil {
		return
	}

	//traverse all files -> send (name,hash,body)
	files, err := ioutil.ReadDir(checkPointFolder)
	if err != nil {
		log.Fatal(err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if dbType == "block" {
			fileNameID, err := strconv.ParseInt(file.Name(), 10, 64)
			if err != nil {
				log.Println(err)
			}
			if ff_fileId > uint64(fileNameID) {
				continue
			}
		}

		fileInfo := FileObject{
			file.Name(), uint64(file.Size()),
		}
		data := new(bytes.Buffer)
		enc := gob.NewEncoder(data)
		err = enc.Encode(fileInfo)
		if err != nil {
			panic(err)
		}

		var dataLen = make([]byte, 8)
		binary.LittleEndian.PutUint64(dataLen, uint64(data.Len()))
		_, err = conn.Write(dataLen)
		_, err = conn.Write(data.Bytes())
		fd, err := os.Open(path.Join(checkPointFolder, file.Name()))
		if err != nil {
			panic(err)
		}

		n, err := io.Copy(conn, fd)


		if file.Size() != int64(n) {
			panic("not correct " + file.Name())
		}
		if err != nil {
			break
		}
	}

	return
}
