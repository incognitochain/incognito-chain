package blockchain

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"io"
	"io/ioutil"
	"log"
	mathrand "math/rand"
	"net/http"
	"os"
	"path"
	"time"
)

type remoteRPCClient struct {
	Endpoint string
}
type ErrMsg struct {
	Code       int
	Message    string
	StackTrace string
}

func (r *remoteRPCClient) sendRequest(requestBody []byte) ([]byte, error) {
	resp, err := http.Post(r.Endpoint, "application/json", bytes.NewBuffer(requestBody))
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

func (r *remoteRPCClient) GetShardCommitteeByBeaconHash(sid int, beaconHash common.Hash) (res []incognitokey.CommitteePublicKey, err error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getshardcommitteestatebybeaconhash",
		"params":  []interface{}{sid, beaconHash},
		"id":      1,
	})
	if err != nil {
		return res, err
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, err
	}

	resp := struct {
		Result []incognitokey.CommitteePublicKey
		Error  *ErrMsg
	}{}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return res, err
	}

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	return resp.Result, nil
}

func (r *remoteRPCClient) GetLatestBackup() (res BackupProcessInfo, err error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getlatestbackup",
		"params":  []interface{}{},
		"id":      1,
	})
	if err != nil {
		return res, err
	}
	body, err := r.sendRequest(requestBody)

	if err != nil {
		return res, err
	}

	resp := struct {
		Result BackupProcessInfo
		Error  *ErrMsg
	}{}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return res, err
	}
	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	return resp.Result, nil
}

func readUint64(r io.Reader) (uint64, error) {
	b := make([]byte, 8)
	n, err := r.Read(b)
	if err != nil && err.Error() != "EOF" {
		return 0, err
	}
	if n == 0 {
		return 0, nil
	}
	var dataLen uint64

	err = binary.Read(bytes.NewBuffer(b), binary.LittleEndian, &dataLen)
	if err != nil || n == 0 {
		return 0, err
	}
	return dataLen, nil
}

func readSize(r io.Reader, size uint64) ([]byte, error) {
	dataByte := make([]byte, size)
	dataRead := uint64(0)
	for {
		n, err := r.Read(dataByte[dataRead:])
		if err != nil {
			return nil, err
		}
		dataRead += uint64(n)
		if dataRead == size {
			break
		}
	}
	return dataByte, nil

}

type FileObject struct {
	Name string
	Size uint64
}

func (r *remoteRPCClient) SyncDB(checkpoint string, cid int, dbType string, offset uint64, dir string) error {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getbootstrapstatedb",
		"params":  []interface{}{checkpoint, cid, dbType, offset},
		"id":      1,
	})
	if err != nil {
		return err
	}
	resp, err := http.Post(r.Endpoint, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		panic(err)
		return err
	}
	defer resp.Body.Close()

	for {
		dataLen, err := readUint64(resp.Body)
		if err != nil {
			panic(err)
		}
		if dataLen == 0 {
			return nil
		}
		infoByte, err := readSize(resp.Body, dataLen)
		readData := bytes.NewBuffer(infoByte)
		dec := gob.NewDecoder(readData)

		fileInfo := &FileObject{}
		err = dec.Decode(fileInfo)
		if err != nil {
			panic(err)
		}
		fileByte, err := readSize(resp.Body, fileInfo.Size)
		fd, err := os.OpenFile(path.Join(dir, fileInfo.Name), os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			panic(err)
		}
		_, err = fd.Write(fileByte)
		if err != nil {
			panic(err)
		}
		fd.Close()
	}
}

func (r *remoteRPCClient) GetStateDB(checkpoint string, cid int, dbType int, offset uint64, f func([]byte)) error {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getbootstrapstatedb",
		"params":  []interface{}{checkpoint, cid, dbType, offset},
		"id":      1,
	})
	if err != nil {
		return err
	}
	resp, err := http.Post(r.Endpoint, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		panic(err)
		return err
	}
	defer resp.Body.Close()
	for {
		b := make([]byte, 8)
		n, err := resp.Body.Read(b)
		if err != nil && err.Error() != "EOF" {
			panic(err)
			return err
		}
		if n == 0 {
			f(nil)
			return nil
		}
		var dataLen uint64

		err = binary.Read(bytes.NewBuffer(b), binary.LittleEndian, &dataLen)
		if err != nil || n == 0 {
			return err
		}
		dataByte := make([]byte, dataLen)
		dataRead := uint64(0)
		for {
			n, _ := resp.Body.Read(dataByte[dataRead:])
			dataRead += uint64(n)
			if dataRead == dataLen {
				break
			}
		}
		f(dataByte)
	}
}

func (r *remoteRPCClient) GetBlocksFromHeight(shardID int, from uint64, num int) (res interface{}, err error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getblocksfromheight",
		"params":  []interface{}{shardID, from, num},
		"id":      1,
	})
	if err != nil {
		return res, err
	}
	body, err := r.sendRequest(requestBody)

	if err != nil {
		return res, err
	}

	if shardID == -1 {
		resp := struct {
			Result []types.BeaconBlock
			Error  *ErrMsg
		}{}
		err = json.Unmarshal(body, &resp)
		if resp.Error != nil && resp.Error.StackTrace != "" {
			return res, errors.New(resp.Error.StackTrace)
		}
		if err != nil {
			return res, err
		}
		return resp.Result, nil
	} else {
		resp := struct {
			Result []types.ShardBlock
			Error  *ErrMsg
		}{}
		err = json.Unmarshal(body, &resp)
		if resp.Error != nil && resp.Error.StackTrace != "" {
			return res, errors.New(resp.Error.StackTrace)
		}
		if err != nil {
			return res, err
		}
		return resp.Result, nil
	}
}

func (s *remoteRPCClient) OnShardBlock(sid int, fromBlk uint64, toBlk uint64, f func(block types.ShardBlock)) {
	shardCh := make(chan types.ShardBlock, 500)
	go func() {
		for {
			data, err := s.GetBlocksFromHeight(sid, uint64(fromBlk), 50)
			if err != nil || len(data.([]types.ShardBlock)) == 0 {
				fmt.Println(err)
				time.Sleep(time.Minute)
				continue
			}
			for _, blk := range data.([]types.ShardBlock) {
				shardCh <- blk
				fromBlk = blk.GetHeight() + 1
				if blk.GetHeight() == toBlk {
					return
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case blk := <-shardCh:
				f(blk)
				if blk.GetHeight() == toBlk {
					return
				}
			}
		}
	}()
}

func (s *remoteRPCClient) OnBeaconBlock(fromBlk uint64, toBlock uint64, f func(block types.BeaconBlock)) {
	beaconCh := make(chan types.BeaconBlock, 500)
	go func() {
		for {
			data, err := s.GetBlocksFromHeight(-1, uint64(fromBlk), 50)
			if err != nil || len(data.([]types.BeaconBlock)) == 0 {
				fmt.Println(err)
				time.Sleep(time.Minute)
				continue
			}
			for _, blk := range data.([]types.BeaconBlock) {
				beaconCh <- blk
				fromBlk = blk.GetHeight() + 1
				if blk.GetHeight() == toBlock {
					return
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case blk := <-beaconCh:
				f(blk)
				if blk.GetHeight() == toBlock {
					return
				}
			}
		}
	}()
}

type BootstrapManager struct {
	hosts      []string
	blockchain *BlockChain
}

func NewBootstrapManager(hosts []string, bc *BlockChain) *BootstrapManager {
	return &BootstrapManager{hosts, bc}
}

func (s *BootstrapManager) BootstrapBeacon() {
	randServer := mathrand.Intn(len(s.hosts))
	host := fmt.Sprintf("http://%v", s.hosts[randServer])
	rpcClient := remoteRPCClient{host}
	latestBackup, _ := rpcClient.GetLatestBackup()

	log.Println("host", host)
	tmp, _ := json.Marshal(latestBackup)
	log.Println("latestBackup", string(tmp))

	if latestBackup.CheckpointName == "" {
		return
	}
	//only backup if back 500k block
	if latestBackup.BeaconView.BeaconHeight < s.blockchain.GetBeaconBestState().BeaconHeight+500*1000 {
		return
	}

	cfg := config.Config()

	tmpDir := path.Join(cfg.DataDir, cfg.DatabaseDir, "beacon.tmp")
	os.RemoveAll(tmpDir)
	mainDir := path.Join(cfg.DataDir, cfg.DatabaseDir, "beacon")
	os.MkdirAll(path.Join(tmpDir, "blockstorage", "blockKV"), 0666)
	bestView := latestBackup.BeaconView

	Logger.log.Info("Start bootstrap beacon from host", host)
	//Logger.log.Infof("Stream block beacon from %v", latestBackup.MinBeaconHeight)
	rpcClient.SyncDB(latestBackup.CheckpointName, -1, "state", 0, tmpDir)
	rpcClient.SyncDB(latestBackup.CheckpointName, -1, "blockKV", 0, path.Join(tmpDir, "blockstorage", "blockKV"))
	rpcClient.SyncDB(latestBackup.CheckpointName, -1, "block", 0, path.Join(tmpDir, "blockstorage"))
	Logger.log.Info("Finish sync ... post processing ...")

	err := s.blockchain.BeaconChain.BlockStorage.ChangeMainDir(tmpDir, mainDir)
	if err != nil {
		panic(err)
	}

	err = s.blockchain.ReloadDatabase(-1)
	if err != nil {
		panic(err)
	}

	allViews := []*BeaconBestState{bestView}
	b, _ := json.Marshal(allViews)
	err = rawdbv2.StoreBeaconViews(s.blockchain.GetBeaconChainDatabase(), b)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	err = s.blockchain.RestoreBeaconViews()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	Logger.log.Info("Bootstrap beacon finish!")
}

func (s *BootstrapManager) BootstrapShard(sid int) {
	randServer := mathrand.Intn(len(s.hosts))
	host := fmt.Sprintf("http://%v", s.hosts[randServer])
	rpcClient := remoteRPCClient{host}
	latestBackup, _ := rpcClient.GetLatestBackup()
	fmt.Println("latestBackup", latestBackup)
	if latestBackup.CheckpointName == "" {
		return
	}

	//only backup if back 500k block
	if latestBackup.ShardView[sid].ShardHeight < s.blockchain.GetBestStateShard(byte(sid)).ShardHeight+500*1000 {
		return
	}

	cfg := config.Config()
	tmpDir := path.Join(cfg.DataDir, cfg.DatabaseDir, fmt.Sprintf("shard%v.tmp", sid))
	os.RemoveAll(tmpDir)
	os.MkdirAll(path.Join(tmpDir, "blockstorage", "blockKV"), 0666)
	mainDir := path.Join(cfg.DataDir, cfg.DatabaseDir, fmt.Sprintf("shard%v", sid))

	//retrieve beacon block -> backup height
	bestView := latestBackup.ShardView[sid]

	Logger.log.Infof("Start bootstrap shard %v from host %v", sid, host)
	rpcClient.SyncDB(latestBackup.CheckpointName, sid, "state", 0, tmpDir)
	rpcClient.SyncDB(latestBackup.CheckpointName, sid, "blockKV", 0, path.Join(tmpDir, "blockstorage", "blockKV"))
	rpcClient.SyncDB(latestBackup.CheckpointName, sid, "block", bestView.ShardHeight-500, path.Join(tmpDir, "blockstorage"))
	Logger.log.Info("Finish sync ... post processing ...")
	s.blockchain.BeaconChain.insertLock.Lock()
	defer s.blockchain.BeaconChain.insertLock.Unlock()

	err := s.blockchain.ShardChain[sid].BlockStorage.ChangeMainDir(tmpDir, mainDir)
	if err != nil {
		panic(err)
	}

	err = s.blockchain.ReloadDatabase(sid)

	if err != nil {
		panic(err)
	}

	allViews := []*ShardBestState{}
	allViews = append(allViews, bestView)
	if err := rawdbv2.StoreShardBestState(s.blockchain.GetShardChainDatabase(byte(sid)), byte(sid), allViews); err != nil {
		panic(err)
	}

	s.blockchain.RestoreShardViews(byte(sid))
	Logger.log.Infof("Bootstrap shard %v finish!", sid)

}
