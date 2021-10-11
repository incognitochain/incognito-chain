package blsbft

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

var (
	keys = []string{
		"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",
	}
	incognitoKeys                   []incognitokey.CommitteePublicKey
	shard0Committee                 []incognitokey.CommitteePublicKey
	shard0CommitteeString           []string
	subset0Shard0Committee          []incognitokey.CommitteePublicKey
	subset0Shard0CommitteeString    []string
	subset0Shard0MiningKeyString    []string
	subset1Shard0Committee          []incognitokey.CommitteePublicKey
	subset1Shard0CommitteeString    []string
	shard0CommitteeNew              []incognitokey.CommitteePublicKey
	subset0Shard0CommitteeNew       []incognitokey.CommitteePublicKey
	subset0Shard0CommitteeStringNew []string
	subset1Shard0CommitteeNew       []incognitokey.CommitteePublicKey
	subset1Shard0CommitteeStringNew []string
	miningKey0                      signatureschemes2.MiningKey
	miningKey1                      signatureschemes2.MiningKey
	logger                          common.Logger
)

var _ = func() (_ struct{}) {
	incognitoKeys, _ = incognitokey.CommitteeBase58KeyListToStruct(keys)
	shard0Committee = incognitoKeys[:8]
	shard0CommitteeString, _ = incognitokey.CommitteeKeyListToString(shard0Committee)
	subset0Shard0Committee = append([]incognitokey.CommitteePublicKey{}, incognitoKeys[0])
	subset0Shard0Committee = append(subset0Shard0Committee, incognitoKeys[2])
	subset0Shard0Committee = append(subset0Shard0Committee, incognitoKeys[4])
	subset0Shard0Committee = append(subset0Shard0Committee, incognitoKeys[6])
	subset1Shard0Committee = append([]incognitokey.CommitteePublicKey{}, incognitoKeys[1])
	subset1Shard0Committee = append(subset1Shard0Committee, incognitoKeys[3])
	subset1Shard0Committee = append(subset1Shard0Committee, incognitoKeys[5])
	subset1Shard0Committee = append(subset1Shard0Committee, incognitoKeys[7])
	subset0Shard0CommitteeString, _ = incognitokey.CommitteeKeyListToString(subset0Shard0Committee)
	subset1Shard0CommitteeString, _ = incognitokey.CommitteeKeyListToString(subset1Shard0Committee)

	shard0CommitteeNew = incognitoKeys[8:16]
	subset0Shard0CommitteeNew = append([]incognitokey.CommitteePublicKey{}, incognitoKeys[8])
	subset0Shard0CommitteeNew = append(subset0Shard0CommitteeNew, incognitoKeys[10])
	subset0Shard0CommitteeNew = append(subset0Shard0CommitteeNew, incognitoKeys[12])
	subset0Shard0CommitteeNew = append(subset0Shard0CommitteeNew, incognitoKeys[14])
	subset1Shard0CommitteeNew = append([]incognitokey.CommitteePublicKey{}, incognitoKeys[9])
	subset1Shard0CommitteeNew = append(subset1Shard0CommitteeNew, incognitoKeys[11])
	subset1Shard0CommitteeNew = append(subset1Shard0CommitteeNew, incognitoKeys[13])
	subset1Shard0CommitteeNew = append(subset1Shard0CommitteeNew, incognitoKeys[15])
	subset0Shard0CommitteeStringNew, _ = incognitokey.CommitteeKeyListToString(subset0Shard0CommitteeNew)
	subset1Shard0CommitteeStringNew, _ = incognitokey.CommitteeKeyListToString(subset1Shard0CommitteeNew)

	miningKey0String := "{\"PriKey\":{\"bls\":\"AwQOHxkPATo6bFFBStpT4U5IgqF/hV5sj3pbTSYbwL4=\",\"dsa\":\"in6dagXYGOMH5zfhBoBdALHSZCzVZQjkmacKTafthU8=\"},\"PubKey\":{\"bls\":\"B6aWM0bu47n7J2acppv55/BvOx+96736VEpiLYia5fEh+ZTjjfxVb4ENJ8NYLi9Iqxi0em9uLQJ3IB4/41s6Ag0rK/t0tPn1uW826hPGAKxQ+vIWj0ObQs7b2l/DAeLKF7izVRn3xZ4+xJn7Pnzv9/xIWld6NgatDiCgglA3xRY=\",\"dsa\":\"A6AjeNXtbUuXgtOs9fC2PMl7v+VryOoZxV18bDaLJdtx\"}}"
	miningKey1String := "{\"PriKey\":{\"bls\":\"DLbqNY4oaqVA+8v4HjQHK2UMxLnt3C2Fm8KeV/5ivgA=\",\"dsa\":\"byqapP9S9HAsnuL5xlpby7xWcBKHTUcLSmSuOrN9htI=\"},\"PubKey\":{\"bls\":\"MAkG6Ub2NXk9tCwiZ4Zv/CyRIMSVukGndHrky6yc+csFPdNwiEdVz4Ci9BLEHShkMCy4nBQ1uWtPlZNmhdaIFBLcnQJ2p56CRUBoFrvCC1KtDTFAYr0MUpWhEZYgnCq+L/pFGXUz3znLEVTAt6MeJsVEx/rSqZS5HBGb0YshEOA=\",\"dsa\":\"A1XEKsZ/6eDoHJK/+8jErVpk1J9Kh30KTtgrp3TqLS6v\"}}"
	miningKey0 = signatureschemes2.MiningKey{}
	json.Unmarshal([]byte(miningKey0String), &miningKey0)
	miningKey1 = signatureschemes2.MiningKey{}
	json.Unmarshal([]byte(miningKey1String), &miningKey1)

	for _, v := range subset0Shard0Committee {
		proposerMiningKeyBase58 := v.GetMiningKeyBase58(common.BlsConsensus)
		subset0Shard0MiningKeyString = append(subset0Shard0MiningKeyString, proposerMiningKeyBase58)
	}
	logger = common.NewBackend(nil).Logger("test", true)

	common.TIMESLOT = 10
	MAX_FINALITY_PROOF = 64
	return
}()

//func Test_actorV2_getValidProposeBlocks(t *testing.T) {
//	common.TIMESLOT = 10
//
//	tc1MockView := &mocksView.View{}
//	tc1ProposeTime := int64(1626755704)
//	tc1BlockHeight := uint64(10)
//	tc1BlockHash := common.HashH([]byte("1"))
//	tc1MockView.On("GetHeight", tc1BlockHeight)
//	tc1Block := &mocksTypes.BlockInterface{}
//	tc1Block.On("GetProposeTime").Return(tc1ProposeTime).Times(2)
//	tc1Block.On("GetProduceTime").Return(tc1ProposeTime).Times(2)
//	tc1Block.On("GetHeight").Return(tc1BlockHeight).Times(4)
//	tc1Block.On("Hash").Return(&tc1BlockHash).Times(4)
//	tc1CurrentTimeSlot := common.CalculateTimeSlot(tc1ProposeTime)
//	tc1MockView.On("GetHeight").Return(tc1BlockHeight)
//	tc1BlockProposeInfo := &ProposeBlockInfo{
//		block:            tc1Block,
//		isVoted:          true,
//		lastValidateTime: time.Now(),
//	}
//	tc1ReceiveBlockByHash := map[string]*ProposeBlockInfo{
//		tc1BlockHash.String(): tc1BlockProposeInfo,
//	}
//
//	oldTimeList, _ := time.Parse(time.RFC822, "Wed, 25 Aug 2021 11:47:34+0000")
//
//	tc2MockView := &mocksView.View{}
//	tc2ProposeTime := int64(1626755704)
//	tc2BlockHeight := uint64(10)
//	tc2BlockHash := common.HashH([]byte("1"))
//	tc2MockView.On("GetHeight", tc2BlockHeight)
//	tc2Block := &mocksTypes.BlockInterface{}
//	tc2Block.On("GetProposeTime").Return(tc2ProposeTime).Times(2)
//	tc2Block.On("GetProduceTime").Return(tc2ProposeTime).Times(2)
//	tc2Block.On("GetHeight").Return(tc2BlockHeight).Times(4)
//	tc2Block.On("Hash").Return(&tc2BlockHash).Times(4)
//	tc2CurrentTimeSlot := common.CalculateTimeSlot(tc2ProposeTime + int64(common.TIMESLOT))
//	tc2MockView.On("GetHeight").Return(tc2BlockHeight)
//	tc2BlockProposeInfo := &ProposeBlockInfo{
//		block:            tc2Block,
//		isVoted:          true,
//		lastValidateTime: oldTimeList,
//	}
//	tc2ReceiveBlockByHash := map[string]*ProposeBlockInfo{
//		tc2BlockHash.String(): tc2BlockProposeInfo,
//	}
//
//	tc3MockView := &mocksView.View{}
//	tc3ProposeTime := int64(1626755704)
//	tc3BlockHeight := uint64(10)
//	tc3BlockHash := common.HashH([]byte("1"))
//	tc3MockView.On("GetHeight", tc3BlockHeight)
//	tc3Block := &mocksTypes.BlockInterface{}
//	tc3Block.On("GetProposeTime").Return(tc3ProposeTime).Times(2)
//	tc3Block.On("GetProduceTime").Return(tc3ProposeTime).Times(2)
//	tc3Block.On("GetHeight").Return(tc3BlockHeight).Times(4)
//	tc3Block.On("Hash").Return(&tc3BlockHash).Times(4)
//	tc3CurrentTimeSlot := common.CalculateTimeSlot(tc3ProposeTime)
//	tc3MockView.On("GetHeight").Return(tc3BlockHeight)
//	tc3BlockProposeInfo := &ProposeBlockInfo{
//		block:            tc3Block,
//		isVoted:          true,
//		lastValidateTime: oldTimeList,
//	}
//	tc3ReceiveBlockByHash := map[string]*ProposeBlockInfo{
//		tc3BlockHash.String(): tc3BlockProposeInfo,
//	}
//
//	tc4MockView := &mocksView.View{}
//	tc4ProposeTime := int64(1626755704)
//	tc4BlockHeight := uint64(10)
//	tc4BlockHash := common.HashH([]byte("1"))
//	tc4MockView.On("GetHeight", tc4BlockHeight)
//	tc4Block := &mocksTypes.BlockInterface{}
//	tc4Block.On("GetProposeTime").Return(tc4ProposeTime).Times(2)
//	tc4Block.On("GetProduceTime").Return(tc4ProposeTime + int64(common.TIMESLOT)).Times(2)
//	tc4Block.On("GetHeight").Return(tc4BlockHeight + 1).Times(4)
//	tc4Block.On("Hash").Return(&tc4BlockHash).Times(4)
//	tc4CurrentTimeSlot := common.CalculateTimeSlot(tc4ProposeTime)
//	tc4MockView.On("GetHeight").Return(tc4BlockHeight)
//	tc4BlockProposeInfo := &ProposeBlockInfo{
//		block:            tc4Block,
//		isVoted:          true,
//		lastValidateTime: oldTimeList,
//	}
//	tc4ReceiveBlockByHash := map[string]*ProposeBlockInfo{
//		tc4BlockHash.String(): tc4BlockProposeInfo,
//	}
//
//	tc5MockView := &mocksView.View{}
//	tc5ProposeTime := int64(1626755704)
//	tc5BlockHeight := uint64(10)
//	tc5BlockHash := common.HashH([]byte("1"))
//	tc5MockView.On("GetHeight", tc5BlockHeight)
//	tc5Block := &mocksTypes.BlockInterface{}
//	tc5Block.On("GetProposeTime").Return(tc5ProposeTime).Times(2)
//	tc5Block.On("GetProduceTime").Return(tc5ProposeTime).Times(2)
//	tc5Block.On("GetHeight").Return(tc5BlockHeight + 1).Times(4)
//	tc5Block.On("Hash").Return(&tc5BlockHash).Times(4)
//	tc5CurrentTimeSlot := common.CalculateTimeSlot(tc5ProposeTime)
//	tc5MockView.On("GetHeight").Return(tc5BlockHeight)
//	tc5BlockProposeInfo := &ProposeBlockInfo{
//		block:            tc5Block,
//		isVoted:          true,
//		lastValidateTime: oldTimeList,
//	}
//	tc5ReceiveBlockByHash := map[string]*ProposeBlockInfo{
//		tc5BlockHash.String(): tc5BlockProposeInfo,
//	}
//
//	tc5MockView2 := &mocksView.View{}
//	tc5MockView2.On("GetHeight").Return(tc5BlockHeight + 2)
//	tc5Chain := &mocks.Chain{}
//	tc5Chain.On("GetFinalView").Return(tc5MockView2)
//
//	tc6MockView := &mocksView.View{}
//	tc6ProposeTime := int64(1626755704)
//	tc6BlockHeight := uint64(10)
//	tc6BlockHash := common.HashH([]byte("1"))
//	tc6BlockHash2 := common.HashH([]byte("2"))
//	tc6MockView.On("GetHeight", tc6BlockHeight)
//	tc6Block := &mocksTypes.BlockInterface{}
//	tc6Block.On("GetProposeTime").Return(tc6ProposeTime).Times(2)
//	tc6Block.On("GetProduceTime").Return(tc6ProposeTime).Times(6)
//	tc6Block.On("GetHeight").Return(tc6BlockHeight + 1).Times(6)
//	tc6Block.On("Hash").Return(&tc6BlockHash).Times(4)
//
//	tc6Block2 := &mocksTypes.BlockInterface{}
//	tc6Block2.On("GetProposeTime").Return(tc6ProposeTime).Times(2)
//	tc6Block2.On("GetProduceTime").Return(tc6ProposeTime - int64(common.TIMESLOT)).Times(6)
//	tc6Block2.On("GetHeight").Return(tc6BlockHeight + 1).Times(6)
//	tc6Block2.On("Hash").Return(&tc6BlockHash2).Times(4)
//
//	tc6CurrentTimeSlot := common.CalculateTimeSlot(tc6ProposeTime)
//	tc6MockView.On("GetHeight").Return(tc6BlockHeight)
//	tc6BlockProposeInfo := &ProposeBlockInfo{
//		block:            tc6Block,
//		isVoted:          true,
//		lastValidateTime: oldTimeList,
//	}
//	tc6BlockProposeInfo2 := &ProposeBlockInfo{
//		block:            tc6Block,
//		isVoted:          true,
//		lastValidateTime: oldTimeList,
//	}
//	tc6ReceiveBlockByHash := map[string]*ProposeBlockInfo{
//		tc6BlockHash.String():  tc6BlockProposeInfo,
//		tc6BlockHash2.String(): tc6BlockProposeInfo2,
//	}
//
//	tc6MockView2 := &mocksView.View{}
//	tc6MockView2.On("GetHeight").Return(tc6BlockHeight - 2)
//	tc6Chain := &mocks.Chain{}
//	tc6Chain.On("GetFinalView").Return(tc6MockView2)
//
//	type fields struct {
//		currentTimeSlot      int64
//		chain                Chain
//		receiveBlockByHeight map[uint64][]*ProposeBlockInfo
//		receiveBlockByHash   map[string]*ProposeBlockInfo
//		voteHistory          map[uint64]types.BlockInterface
//		node                 NodeInterface
//	}
//	type args struct {
//		bestView multiview.View
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//		want   []*ProposeBlockInfo
//	}{
//		//TODO: Add test cases.
//		{
//			name: "just validate recently",
//			fields: fields{
//				currentTimeSlot:    tc1CurrentTimeSlot,
//				receiveBlockByHash: tc1ReceiveBlockByHash,
//			},
//			args: args{
//				tc1MockView,
//			},
//			want: []*ProposeBlockInfo{},
//		},
//		{
//			name: "not propose in time slot",
//			fields: fields{
//				currentTimeSlot:    tc2CurrentTimeSlot,
//				receiveBlockByHash: tc2ReceiveBlockByHash,
//			},
//			args: args{
//				tc2MockView,
//			},
//			want: []*ProposeBlockInfo{},
//		},
//		{
//			name: "not connect to best height",
//			fields: fields{
//				currentTimeSlot:    tc3CurrentTimeSlot,
//				receiveBlockByHash: tc3ReceiveBlockByHash,
//			},
//			args: args{
//				tc3MockView,
//			},
//			want: []*ProposeBlockInfo{},
//		},
//		{
//			name: "producer time < current timeslot",
//			fields: fields{
//				currentTimeSlot:    tc4CurrentTimeSlot,
//				receiveBlockByHash: tc4ReceiveBlockByHash,
//			},
//			args: args{
//				tc4MockView,
//			},
//			want: []*ProposeBlockInfo{},
//		},
//		{
//			name: "propose block info height < final view",
//			fields: fields{
//				currentTimeSlot:    tc5CurrentTimeSlot,
//				receiveBlockByHash: tc5ReceiveBlockByHash,
//				chain:              tc5Chain,
//			},
//			args: args{
//				tc5MockView,
//			},
//			want: []*ProposeBlockInfo{},
//		},
//		{
//			name: "add valid propose block info",
//			fields: fields{
//				currentTimeSlot:    tc6CurrentTimeSlot,
//				receiveBlockByHash: tc6ReceiveBlockByHash,
//				chain:              tc6Chain,
//			},
//			args: args{
//				tc6MockView,
//			},
//			want: []*ProposeBlockInfo{tc6BlockProposeInfo2, tc6BlockProposeInfo},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			a := &actorV2{
//				currentTimeSlot:    tt.fields.currentTimeSlot,
//				receiveBlockByHash: tt.fields.receiveBlockByHash,
//				chain:              tt.fields.chain,
//				voteHistory:        tt.fields.voteHistory,
//				node:               tt.fields.node,
//			}
//			if got := a.getValidProposeBlocks(tt.args.bestView); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("getValidProposeBlocks() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

//func Test_actorV2_validatePreSignBlock(t *testing.T) {
//
//	common.TIMESLOT = 10
//	tc1BlockHash := common.HashH([]byte("1"))
//	tc1PrevBlockHash := common.HashH([]byte("0"))
//	tc1ProducerString := subset0Shard0CommitteeString[0]
//	tc1CommitteeViewHash := common.HashH([]byte("view-hash-1"))
//	tc1ProposeTime := int64(1626755704)
//	tc1BlockHeight := uint64(10)
//	tc1Block := &mocksTypes.BlockInterface{}
//	tc1Block.On("CommitteeFromBlock").Return(tc1CommitteeViewHash).Times(2)
//	tc1Block.On("GetVersion").Return(types.BLOCK_PRODUCINGV3_VERSION).Times(2)
//	tc1Block.On("GetProposeTime").Return(tc1ProposeTime).Times(2)
//	tc1Block.On("GetProduceTime").Return(tc1ProposeTime).Times(2)
//	tc1Block.On("GetProducer").Return(tc1ProducerString).Times(2)
//	tc1Block.On("GetHeight").Return(tc1BlockHeight).Times(4)
//	tc1Block.On("Hash").Return(&tc1BlockHash).Times(4)
//	tc1Block.On("GetPrevHash").Return(tc1PrevBlockHash).Times(2)
//	tc1ProposeBlockInfo := &ProposeBlockInfo{
//		block:             tc1Block,
//		signingCommittees: subset0Shard0Committee,
//		committees:        shard0Committee,
//	}
//	tc1MockView := &mocksView.View{}
//
//	tc1Chain := &mocks.Chain{}
//	tc1Chain.On("GetViewByHash", tc1PrevBlockHash).Return(tc1MockView)
//	tc1Chain.On("ValidatePreSignBlock",
//		tc1ProposeBlockInfo.block,
//		tc1ProposeBlockInfo.signingCommittees,
//		tc1ProposeBlockInfo.committees).Return(nil)
//
//	tc2BlockHash := common.HashH([]byte("2"))
//	tc2PrevBlockHash := common.HashH([]byte("00"))
//	tc2ProducerString := subset0Shard0CommitteeString[0]
//	tc2CommitteeViewHash := common.HashH([]byte("view-hash-2"))
//	tc2ProposeTime := int64(1626755704)
//	tc2BlockHeight := uint64(10)
//	tc2Block := &mocksTypes.BlockInterface{}
//	tc2Block.On("CommitteeFromBlock").Return(tc2CommitteeViewHash).Times(2)
//	tc2Block.On("GetVersion").Return(types.BLOCK_PRODUCINGV3_VERSION).Times(2)
//	tc2Block.On("GetProposeTime").Return(tc2ProposeTime).Times(2)
//	tc2Block.On("GetProduceTime").Return(tc2ProposeTime).Times(2)
//	tc2Block.On("GetProducer").Return(tc2ProducerString).Times(2)
//	tc2Block.On("GetHeight").Return(tc2BlockHeight).Times(4)
//	tc2Block.On("Hash").Return(&tc2BlockHash).Times(4)
//	tc2Block.On("GetPrevHash").Return(tc2PrevBlockHash).Times(2)
//	tc2ProposeBlockInfo := &ProposeBlockInfo{
//		block:             tc2Block,
//		signingCommittees: subset0Shard0Committee,
//		committees:        shard0Committee,
//	}
//	//tc2MockView := &mocksView.View{}
//
//	tc2Chain := &mocks.Chain{}
//	tc2Chain.On("GetViewByHash", tc2PrevBlockHash).Return(nil)
//	tc2Chain.On("ValidatePreSignBlock",
//		tc2ProposeBlockInfo.block,
//		tc2ProposeBlockInfo.signingCommittees,
//		tc2ProposeBlockInfo.committees).Return(nil)
//
//	tc3BlockHash := common.HashH([]byte("3"))
//	tc3PrevBlockHash := common.HashH([]byte("000"))
//	tc3ProducerString := subset0Shard0CommitteeString[0]
//	tc3CommitteeViewHash := common.HashH([]byte("view-hash-3"))
//	tc3ProposeTime := int64(1626755704)
//	tc3BlockHeight := uint64(10)
//	tc3Block := &mocksTypes.BlockInterface{}
//	tc3Block.On("CommitteeFromBlock").Return(tc3CommitteeViewHash).Times(2)
//	tc3Block.On("GetVersion").Return(types.BLOCK_PRODUCINGV3_VERSION).Times(2)
//	tc3Block.On("GetProposeTime").Return(tc3ProposeTime).Times(2)
//	tc3Block.On("GetProduceTime").Return(tc3ProposeTime).Times(2)
//	tc3Block.On("GetProducer").Return(tc3ProducerString).Times(2)
//	tc3Block.On("GetHeight").Return(tc3BlockHeight).Times(4)
//	tc3Block.On("Hash").Return(&tc3BlockHash).Times(4)
//	tc3Block.On("GetPrevHash").Return(tc3PrevBlockHash).Times(2)
//	tc3ProposeBlockInfo := &ProposeBlockInfo{
//		block:             tc3Block,
//		signingCommittees: subset0Shard0Committee,
//		committees:        shard0Committee,
//	}
//	tc3MockView := &mocksView.View{}
//
//	tc3Chain := &mocks.Chain{}
//	tc3Chain.On("GetViewByHash", tc3PrevBlockHash).Return(tc3MockView)
//	tc3Chain.On("ValidatePreSignBlock",
//		tc3ProposeBlockInfo.block,
//		tc3ProposeBlockInfo.signingCommittees,
//		tc3ProposeBlockInfo.committees).Return(errors.New("some error"))
//
//	type fields struct {
//		chain        Chain
//		logger       common.Logger
//		blockVersion int
//	}
//	type args struct {
//		proposeBlockInfo *ProposeBlockInfo
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		{
//			name: "validate pre sign block success",
//			fields: fields{
//				chain:        tc1Chain,
//				logger:       logger,
//				blockVersion: types.BLOCK_PRODUCINGV3_VERSION,
//			},
//			args: args{
//				proposeBlockInfo: tc1ProposeBlockInfo,
//			},
//			wantErr: false,
//		},
//		{
//			name: "block propose info previous hash view not found",
//			fields: fields{
//				chain:        tc2Chain,
//				logger:       logger,
//				blockVersion: types.BLOCK_PRODUCINGV3_VERSION,
//			},
//			args: args{
//				proposeBlockInfo: tc2ProposeBlockInfo,
//			},
//			wantErr: true,
//		},
//		{
//			name: "validate pre sign block fail",
//			fields: fields{
//				chain:        tc3Chain,
//				logger:       logger,
//				blockVersion: types.BLOCK_PRODUCINGV3_VERSION,
//			},
//			args: args{
//				proposeBlockInfo: tc3ProposeBlockInfo,
//			},
//			wantErr: true,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			a := &actorV2{
//				chain:        tt.fields.chain,
//				logger:       tt.fields.logger,
//				blockVersion: tt.fields.blockVersion,
//			}
//			if err := a.validatePreSignBlock(tt.args.proposeBlockInfo); (err != nil) != tt.wantErr {
//				t.Errorf("ValidatePreSignBlock() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
