package devframework

import (
	"fmt"
	"log"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
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
				//fmt.Println(err)
				//time.Sleep(time.Minute)
				time.Sleep(time.Second * 10)
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

func (s *AppService) PreparePRVForTest(privateKey string, receivers map[string]interface{}) {
	fullnodeRPC := RemoteRPCClient{s.Fullnode}
	_, err := fullnodeRPC.PreparePRVForTest(privateKey, receivers)
	if err != nil {
		panic(err)
	}
}

func (s *AppService) ConvertTokenV1ToV2(privateKey string) {
	fullnodeRPC := RemoteRPCClient{s.Fullnode}
	if err := fullnodeRPC.CreateConvertCoinVer1ToVer2Transaction(privateKey); err != nil {
		log.Println(err)
	}
}

func (s *AppService) SubmitKey(otaPrivateKey string) {
	fullnodeRPC := RemoteRPCClient{s.Fullnode}
	if _, err := fullnodeRPC.SubmitKey(otaPrivateKey); err != nil {
		log.Println(err)
	}
}

func (s *AppService) ShardStaking(privateKey, candidatePaymentAddress, privateSeed, rewardReceiverPaymentAddress, delegate string, autoReStaking bool) {
	fullnodeRPC := RemoteRPCClient{s.Fullnode}
	bAddr, err := fullnodeRPC.GetBurningAddress(1)
	if err != nil {
		panic(err)
	}
	if resp, err := fullnodeRPC.CreateAndSendStakingTransaction(
		privateKey,
		map[string]interface{}{
			bAddr: 1750000000000,
		},
		-1,
		0,
		map[string]interface{}{
			"StakingType":                  63,
			"CandidatePaymentAddress":      candidatePaymentAddress,
			"PrivateSeed":                  privateSeed,
			"RewardReceiverPaymentAddress": rewardReceiverPaymentAddress,
			"Delegate":                     delegate,
			"AutoReStaking":                autoReStaking,
		},
	); err != nil {
		panic(err)
	} else {
		log.Println("shard stake with tx ", resp.TxID)
	}
}

func (s *AppService) BeaconStaking(privateKey, candidatePaymentAddress, privateSeed, rewardReceiverPaymentAddress, delegate string, autoReStaking bool) {
	fullnodeRPC := RemoteRPCClient{s.Fullnode}
	bAddr, err := fullnodeRPC.GetBurningAddress(1)
	if err != nil {
		panic(err)
	}
	if resp, err := fullnodeRPC.CreateAndSendStakingTransaction(
		privateKey,
		map[string]interface{}{
			bAddr: 87500000000000,
		},
		-1,
		0,
		map[string]interface{}{
			"StakingType":                  64,
			"CandidatePaymentAddress":      candidatePaymentAddress,
			"PrivateSeed":                  privateSeed,
			"RewardReceiverPaymentAddress": rewardReceiverPaymentAddress,
			"Delegate":                     delegate,
			"AutoReStaking":                autoReStaking,
		},
	); err != nil {
		panic(err)
	} else {
		log.Println("beacon stake with tx ", resp.TxID)
	}
}

func (s *AppService) ShardUnstaking(privateKey, candidatePaymentAddress, privateSeed string) {
	fullnodeRPC := RemoteRPCClient{s.Fullnode}
	bAddr, err := fullnodeRPC.GetBurningAddress(1)
	if err != nil {
		panic(err)
	}
	if _, err := fullnodeRPC.CreateAndSendUnStakingTransaction(
		privateKey,
		map[string]interface{}{
			bAddr: 0,
		},
		-1,
		0,
		map[string]interface{}{
			"StopAutoStakingType":     127,
			"CandidatePaymentAddress": candidatePaymentAddress,
			"PrivateSeed":             privateSeed,
		},
	); err != nil {
		panic(err)
	}
}

func (s *AppService) GetBeaconBestState() (jsonresult.GetBeaconBestState, error) {
	fullnodeRPC := RemoteRPCClient{s.Fullnode}
	return fullnodeRPC.GetBeaconBestState()
}

func (s *AppService) GetCommitteeState(height uint64, hash string) (*jsonresult.CommiteeState, error) {
	fullnodeRPC := RemoteRPCClient{s.Fullnode}
	return fullnodeRPC.GetCommitteeState(height, hash)
}

func (s *AppService) GetShardStakerInfo(height uint64, stakerPubkey string) (*statedb.ShardStakerInfo, error) {
	fullnodeRPC := RemoteRPCClient{s.Fullnode}
	return fullnodeRPC.GetShardStakerInfo(height, stakerPubkey)
}

func (s *AppService) GetBeaconStakerInfo(height uint64, stakerPubkey string) (*statedb.BeaconStakerInfo, error) {
	fullnodeRPC := RemoteRPCClient{s.Fullnode}
	return fullnodeRPC.GetBeaconStakerInfo(height, stakerPubkey)
}

func (s *AppService) StopAutoStaking(privateKey, candidatePaymentAddress, privateSeed string) (jsonresult.CreateTransactionResult, error) {
	fullnodeRPC := RemoteRPCClient{s.Fullnode}
	bAddr, err := fullnodeRPC.GetBurningAddress(1)
	if err != nil {
		panic(err)
	}
	return fullnodeRPC.CreateAndSendStopAutoStakingTransaction(
		privateKey,
		map[string]interface{}{
			bAddr: 100,
		},
		-1,
		0,
		map[string]interface{}{
			"StopAutoStakingType":     127,
			"CandidatePaymentAddress": candidatePaymentAddress,
			"PrivateSeed":             privateSeed,
		},
	)
}

func (s *AppService) AddStaking(privateKey, privateSeed, candidatePaymentAddress string, amount uint64) (jsonresult.CreateTransactionResult, error) {
	fullnodeRPC := RemoteRPCClient{s.Fullnode}
	bAddr, err := fullnodeRPC.GetBurningAddress(1)
	if err != nil {
		panic(err)
	}

	return fullnodeRPC.CreateAndSendAddStakingTransaction(
		privateKey,
		map[string]interface{}{
			bAddr: amount,
		},
		-1,
		0,
		map[string]interface{}{
			"AddStakingAmount":        amount,
			"PrivateSeed":             privateSeed,
			"CandidatePaymentAddress": candidatePaymentAddress,
		},
	)
}
