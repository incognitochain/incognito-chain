package devframework

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"time"
)

func NewAppService(fullnode string, finalizedBlock bool) *AppService {
	return &AppService{
		fullnode, finalizedBlock,
	}
}

type AppService struct {
	Fullnode       string
	FinalizedBlock bool
}

func (s *AppService) OnBeaconBlock(fromBlk uint64, f func(block types.BeaconBlock)) {
	beaconCh := make(chan types.BeaconBlock, 500)
	fullnodeRPC := RemoteRPCClient{s.Fullnode}
	go func() {
		for {
			data, err := fullnodeRPC.GetBlocksFromHeight(-1, uint64(fromBlk), 50)
			if err != nil || len(data.([]types.BeaconBlock)) == 0 {
				fmt.Println(err)
				time.Sleep(time.Minute)
				continue
			}
			for _, blk := range data.([]types.BeaconBlock) {
				beaconCh <- blk
				fromBlk = blk.GetHeight() + 1
			}
		}
	}()

	go func() {
		for {
			select {
			case blk := <-beaconCh:
				f(blk)
			}
		}
	}()

}

func (s *AppService) OnShardBlock(sid int, fromBlk uint64, f func(block types.ShardBlock)) {
	shardCh := make(chan types.ShardBlock, 500)
	fullnodeRPC := RemoteRPCClient{s.Fullnode}
	go func() {
		for {
			fmt.Println("stream sid", sid, fromBlk)
			data, err := fullnodeRPC.GetBlocksFromHeight(sid, uint64(fromBlk), 50)
			if err != nil || len(data.([]types.ShardBlock)) == 0 {
				fmt.Println(err)
				time.Sleep(time.Minute)
				continue
			}
			for _, blk := range data.([]types.ShardBlock) {
				shardCh <- blk
				fromBlk = blk.GetHeight() + 1
			}
		}
	}()

	go func() {
		for {
			select {
			case blk := <-shardCh:
				f(blk)
			}
		}
	}()
}

func (s *AppService) OnStateDBData(checkpoint string, cid int, dbType int, offset uint64, f func(data blockchain.StateDBData)) {
	fullnodeRPC := RemoteRPCClient{s.Fullnode}
	fullnodeRPC.GetStateDB(checkpoint, cid, dbType, offset, func(dataByte []byte) {
		//TODO: parse state data and then pass to f
	})
}

func (s *AppService) GetOTACoinByIndices(index uint64, shardid int, token string) string {
	fullnodeRPC := RemoteRPCClient{s.Fullnode}
	res, err := fullnodeRPC.GetOTAcoinsbyindices(index, shardid, token)
	if err != nil {
		panic(err)
	}
	if _, ok := res[index]; !ok {
		fmt.Println(res)
		panic(1)
	}
	return res[index].PublicKey
}
