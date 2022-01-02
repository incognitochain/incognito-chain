package blsbft

import (
	mocksTypes "github.com/incognitochain/incognito-chain/blockchain/types/mocks"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
	"testing"
	"time"
)

func TestConsensusValidatorLemma1_FilterValidProposeBlockInfo(t *testing.T) {
	tc1ProposeTime := int64(1626755704)
	tc1BestViewHeight := uint64(9)
	tc1BestViewHash := common.HashH([]byte("2"))
	tc1FinalViewHeight := uint64(8)
	tc1BlockHeight := uint64(10)
	tc1BlockHash := common.HashH([]byte("1"))
	tc1Block := &mocksTypes.BlockInterface{}
	tc1Block.On("GetProposeTime").Return(tc1ProposeTime).Times(2)
	tc1Block.On("GetProduceTime").Return(tc1ProposeTime).Times(2)
	tc1Block.On("GetHeight").Return(tc1BlockHeight).Times(4)
	tc1Block.On("Hash").Return(&tc1BlockHash).Times(4)
	tc1CurrentTimeSlot := common.CalculateTimeSlot(tc1ProposeTime)
	tc1BlockProposeInfo := &ProposeBlockInfo{
		block:            tc1Block,
		IsVoted:          true,
		LastValidateTime: time.Now(),
	}
	tc1ReceiveBlockByHash := map[string]*ProposeBlockInfo{
		tc1BlockHash.String(): tc1BlockProposeInfo,
	}

	oldTimeList, _ := time.Parse(time.RFC822, "Wed, 25 Aug 2021 11:47:34+0000")

	tc2ProposeTime := int64(1626755704)
	tc2BlockHeight := uint64(10)
	tc2BestViewHeight := uint64(9)
	tc2BestViewHash := common.HashH([]byte("2"))
	tc2FinalViewHeight := uint64(8)
	tc2BlockHash := common.HashH([]byte("1"))
	tc2Block := &mocksTypes.BlockInterface{}
	tc2Block.On("GetProposeTime").Return(tc2ProposeTime).Times(2)
	tc2Block.On("GetProduceTime").Return(tc2ProposeTime).Times(2)
	tc2Block.On("GetHeight").Return(tc2BlockHeight).Times(4)
	tc2Block.On("Hash").Return(&tc2BlockHash).Times(4)
	tc2CurrentTimeSlot := common.CalculateTimeSlot(tc2ProposeTime + int64(common.TIMESLOT))
	tc2BlockProposeInfo := &ProposeBlockInfo{
		block:            tc2Block,
		IsVoted:          true,
		LastValidateTime: oldTimeList,
	}
	tc2ReceiveBlockByHash := map[string]*ProposeBlockInfo{
		tc2BlockHash.String(): tc2BlockProposeInfo,
	}

	tc3ProposeTime := int64(1626755704)
	tc3BlockHeight := uint64(10)
	tc3BlockHash := common.HashH([]byte("1"))
	tc3BestViewHeight := uint64(8)
	tc3BestViewHash := common.HashH([]byte("2"))
	tc3FinalViewHeight := uint64(8)
	tc3Block := &mocksTypes.BlockInterface{}
	tc3Block.On("GetProposeTime").Return(tc3ProposeTime).Times(2)
	tc3Block.On("GetProduceTime").Return(tc3ProposeTime).Times(2)
	tc3Block.On("GetHeight").Return(tc3BlockHeight).Times(4)
	tc3Block.On("Hash").Return(&tc3BlockHash).Times(4)
	tc3CurrentTimeSlot := common.CalculateTimeSlot(tc3ProposeTime)
	tc3BlockProposeInfo := &ProposeBlockInfo{
		block:            tc3Block,
		IsVoted:          true,
		LastValidateTime: oldTimeList,
	}
	tc3ReceiveBlockByHash := map[string]*ProposeBlockInfo{
		tc3BlockHash.String(): tc3BlockProposeInfo,
	}

	tc4ProposeTime := int64(1626755704)
	tc4BlockHeight := uint64(10)
	tc4BlockHash := common.HashH([]byte("1"))
	tc4BestViewHeight := uint64(10)
	tc4BestViewHash := common.HashH([]byte("2"))
	tc4FinalViewHeight := uint64(8)
	tc4Block := &mocksTypes.BlockInterface{}
	tc4Block.On("GetProposeTime").Return(tc4ProposeTime).Times(2)
	tc4Block.On("GetProduceTime").Return(tc4ProposeTime + int64(common.TIMESLOT)).Times(2)
	tc4Block.On("GetHeight").Return(tc4BlockHeight + 1).Times(4)
	tc4Block.On("Hash").Return(&tc4BlockHash).Times(4)
	tc4CurrentTimeSlot := common.CalculateTimeSlot(tc4ProposeTime)
	tc4BlockProposeInfo := &ProposeBlockInfo{
		block:            tc4Block,
		IsVoted:          true,
		LastValidateTime: oldTimeList,
	}
	tc4ReceiveBlockByHash := map[string]*ProposeBlockInfo{
		tc4BlockHash.String(): tc4BlockProposeInfo,
	}

	tc5ProposeTime := int64(1626755704)
	tc5BlockHeight := uint64(10)
	tc5BlockHash := common.HashH([]byte("1"))
	tc5BestViewHeight := uint64(9)
	tc5BestViewHash := common.HashH([]byte("2"))
	tc5FinalViewHeight := uint64(11)
	tc5Block := &mocksTypes.BlockInterface{}
	tc5Block.On("GetProposeTime").Return(tc5ProposeTime).Times(2)
	tc5Block.On("GetProduceTime").Return(tc5ProposeTime).Times(2)
	tc5Block.On("GetHeight").Return(tc5BlockHeight).Times(4)
	tc5Block.On("Hash").Return(&tc5BlockHash).Times(4)
	tc5CurrentTimeSlot := common.CalculateTimeSlot(tc5ProposeTime)
	tc5BlockProposeInfo := &ProposeBlockInfo{
		block:            tc5Block,
		IsVoted:          true,
		LastValidateTime: oldTimeList,
	}
	tc5ReceiveBlockByHash := map[string]*ProposeBlockInfo{
		tc5BlockHash.String(): tc5BlockProposeInfo,
	}

	tc6ProposeTime := int64(1626755704)
	tc6BlockHeight := uint64(10)
	tc6BestViewHeight := uint64(10)
	tc6BestViewHash := common.HashH([]byte("2"))
	tc6FinalViewHeight := uint64(8)
	tc6BlockHash := common.HashH([]byte("1"))
	tc6BlockHash2 := common.HashH([]byte("2"))
	tc6Block := &mocksTypes.BlockInterface{}
	tc6Block.On("GetProposeTime").Return(tc6ProposeTime).Times(2)
	tc6Block.On("GetProduceTime").Return(tc6ProposeTime).Times(6)
	tc6Block.On("GetHeight").Return(tc6BlockHeight + 1).Times(6)
	tc6Block.On("Hash").Return(&tc6BlockHash).Times(4)

	tc6Block2 := &mocksTypes.BlockInterface{}
	tc6Block2.On("GetProposeTime").Return(tc6ProposeTime).Times(2)
	tc6Block2.On("GetProduceTime").Return(tc6ProposeTime - int64(common.TIMESLOT)).Times(6)
	tc6Block2.On("GetHeight").Return(tc6BlockHeight + 1).Times(6)
	tc6Block2.On("Hash").Return(&tc6BlockHash2).Times(4)

	tc6CurrentTimeSlot := common.CalculateTimeSlot(tc6ProposeTime)
	tc6BlockProposeInfo := &ProposeBlockInfo{
		block:            tc6Block,
		IsVoted:          true,
		LastValidateTime: oldTimeList,
	}
	tc6BlockProposeInfo2 := &ProposeBlockInfo{
		block:            tc6Block2,
		IsVoted:          true,
		LastValidateTime: oldTimeList,
	}
	tc6ReceiveBlockByHash := map[string]*ProposeBlockInfo{
		tc6BlockHash.String():  tc6BlockProposeInfo,
		tc6BlockHash2.String(): tc6BlockProposeInfo2,
	}

	tc7ProposeTime := int64(1626755704)
	tc7BestViewHeight := uint64(10)
	tc7BestViewHash := common.HashH([]byte("1"))
	tc7FinalViewHeight := uint64(8)
	tc7BlockHeight := uint64(10)
	tc7BlockHash := common.HashH([]byte("1"))
	tc7Block := &mocksTypes.BlockInterface{}
	tc7Block.On("GetProposeTime").Return(tc7ProposeTime).Times(2)
	tc7Block.On("GetProduceTime").Return(tc7ProposeTime).Times(2)
	tc7Block.On("GetHeight").Return(tc7BlockHeight).Times(4)
	tc7Block.On("Hash").Return(&tc7BlockHash).Times(4)
	tc7CurrentTimeSlot := common.CalculateTimeSlot(tc7ProposeTime)
	tc7BlockProposeInfo := &ProposeBlockInfo{
		block:            tc7Block,
		IsVoted:          false,
		LastValidateTime: time.Now(),
	}
	tc7ReceiveBlockByHash := map[string]*ProposeBlockInfo{
		tc7BlockHash.String(): tc7BlockProposeInfo,
	}

	type fields struct {
		logger common.Logger
		chain  Chain
	}
	type args struct {
		bestViewHash      common.Hash
		bestViewHeight    uint64
		finalViewHeight   uint64
		currentTimeSlot   int64
		proposeBlockInfos map[string]*ProposeBlockInfo
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantValid   []*ProposeBlockInfo
		wantReVote  []*ProposeBlockInfo
		wantInvalid []string
	}{
		{
			name:   "tc1: just validate recently",
			fields: fields{},
			args: args{
				bestViewHash:      tc1BestViewHash,
				bestViewHeight:    tc1BestViewHeight,
				finalViewHeight:   tc1FinalViewHeight,
				currentTimeSlot:   tc1CurrentTimeSlot,
				proposeBlockInfos: tc1ReceiveBlockByHash,
			},
			wantValid:   []*ProposeBlockInfo{},
			wantReVote:  []*ProposeBlockInfo{},
			wantInvalid: []string{},
		},
		{
			name:   "tc2: not propose in time slot",
			fields: fields{},
			args: args{
				bestViewHash:      tc2BestViewHash,
				bestViewHeight:    tc2BestViewHeight,
				finalViewHeight:   tc2FinalViewHeight,
				currentTimeSlot:   tc2CurrentTimeSlot,
				proposeBlockInfos: tc2ReceiveBlockByHash,
			},
			wantValid:   []*ProposeBlockInfo{},
			wantReVote:  []*ProposeBlockInfo{},
			wantInvalid: []string{},
		},
		{
			name:   "tc3: not connect to best height",
			fields: fields{},
			args: args{
				bestViewHash:      tc3BestViewHash,
				bestViewHeight:    tc3BestViewHeight,
				finalViewHeight:   tc3FinalViewHeight,
				currentTimeSlot:   tc3CurrentTimeSlot,
				proposeBlockInfos: tc3ReceiveBlockByHash,
			},
			wantValid:   []*ProposeBlockInfo{},
			wantReVote:  []*ProposeBlockInfo{},
			wantInvalid: []string{},
		},
		{
			name:   "tc4: producer time < current timeslot",
			fields: fields{},
			args: args{
				bestViewHash:      tc4BestViewHash,
				bestViewHeight:    tc4BestViewHeight,
				finalViewHeight:   tc4FinalViewHeight,
				currentTimeSlot:   tc4CurrentTimeSlot,
				proposeBlockInfos: tc4ReceiveBlockByHash,
			},
			wantValid:   []*ProposeBlockInfo{},
			wantReVote:  []*ProposeBlockInfo{},
			wantInvalid: []string{},
		},
		{
			name:   "tc5: propose block info height < final view",
			fields: fields{},
			args: args{
				bestViewHash:      tc5BestViewHash,
				bestViewHeight:    tc5BestViewHeight,
				finalViewHeight:   tc5FinalViewHeight,
				currentTimeSlot:   tc5CurrentTimeSlot,
				proposeBlockInfos: tc5ReceiveBlockByHash,
			},
			wantValid:   []*ProposeBlockInfo{},
			wantReVote:  []*ProposeBlockInfo{},
			wantInvalid: []string{tc5BlockHash.String()},
		},
		{
			name:   "tc6: add valid propose block info",
			fields: fields{},
			args: args{
				bestViewHash:      tc6BestViewHash,
				bestViewHeight:    tc6BestViewHeight,
				finalViewHeight:   tc6FinalViewHeight,
				currentTimeSlot:   tc6CurrentTimeSlot,
				proposeBlockInfos: tc6ReceiveBlockByHash,
			},
			wantValid:   []*ProposeBlockInfo{tc6BlockProposeInfo2, tc6BlockProposeInfo},
			wantReVote:  []*ProposeBlockInfo{},
			wantInvalid: []string{},
		},
		{
			name:   "tc7: re-vote insert block",
			fields: fields{},
			args: args{
				bestViewHash:      tc7BestViewHash,
				bestViewHeight:    tc7BestViewHeight,
				finalViewHeight:   tc7FinalViewHeight,
				currentTimeSlot:   tc7CurrentTimeSlot,
				proposeBlockInfos: tc7ReceiveBlockByHash,
			},
			wantValid:   []*ProposeBlockInfo{},
			wantReVote:  []*ProposeBlockInfo{tc7BlockProposeInfo},
			wantInvalid: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := ConsensusValidatorLemma1{
				logger: tt.fields.logger,
				chain:  tt.fields.chain,
			}
			got, got1, got2 := c.FilterValidProposeBlockInfo(tt.args.bestViewHash, tt.args.bestViewHeight, tt.args.finalViewHeight, tt.args.currentTimeSlot, tt.args.proposeBlockInfos)
			if !reflect.DeepEqual(got, tt.wantValid) {
				t.Errorf("FilterValidProposeBlockInfo() got = %v, want %v", got, tt.wantValid)
			}
			if !reflect.DeepEqual(got1, tt.wantReVote) {
				t.Errorf("FilterValidProposeBlockInfo() got1 = %v, want %v", got1, tt.wantReVote)
			}
			if !reflect.DeepEqual(got2, tt.wantInvalid) {
				t.Errorf("FilterValidProposeBlockInfo() got2 = %v, want %v", got2, tt.wantInvalid)
			}
		})
	}
}

func TestConsensusValidatorLemma2_FilterValidProposeBlockInfo(t *testing.T) {
	tc1ProposeTime := int64(1626755704)
	tc1BestViewHeight := uint64(9)
	tc1BestViewHash := common.HashH([]byte("2"))
	tc1FinalViewHeight := uint64(8)
	tc1BlockHeight := uint64(10)
	tc1BlockHash := common.HashH([]byte("1"))
	tc1Block := &mocksTypes.BlockInterface{}
	tc1Block.On("GetProposeTime").Return(tc1ProposeTime).Times(2)
	tc1Block.On("GetProduceTime").Return(tc1ProposeTime).Times(2)
	tc1Block.On("GetHeight").Return(tc1BlockHeight).Times(4)
	tc1Block.On("Hash").Return(&tc1BlockHash).Times(4)
	tc1CurrentTimeSlot := common.CalculateTimeSlot(tc1ProposeTime)
	tc1BlockProposeInfo := &ProposeBlockInfo{
		block:            tc1Block,
		IsVoted:          true,
		LastValidateTime: time.Now(),
	}
	tc1ReceiveBlockByHash := map[string]*ProposeBlockInfo{
		tc1BlockHash.String(): tc1BlockProposeInfo,
	}

	oldTimeList, _ := time.Parse(time.RFC822, "Wed, 25 Aug 2021 11:47:34+0000")

	tc2ProposeTime := int64(1626755704)
	tc2BlockHeight := uint64(10)
	tc2BestViewHeight := uint64(9)
	tc2BestViewHash := common.HashH([]byte("2"))
	tc2FinalViewHeight := uint64(8)
	tc2BlockHash := common.HashH([]byte("1"))
	tc2Block := &mocksTypes.BlockInterface{}
	tc2Block.On("GetProposeTime").Return(tc2ProposeTime).Times(2)
	tc2Block.On("GetProduceTime").Return(tc2ProposeTime).Times(2)
	tc2Block.On("GetHeight").Return(tc2BlockHeight).Times(4)
	tc2Block.On("Hash").Return(&tc2BlockHash).Times(4)
	tc2CurrentTimeSlot := common.CalculateTimeSlot(tc2ProposeTime + int64(common.TIMESLOT))
	tc2BlockProposeInfo := &ProposeBlockInfo{
		block:            tc2Block,
		IsVoted:          true,
		LastValidateTime: oldTimeList,
	}
	tc2ReceiveBlockByHash := map[string]*ProposeBlockInfo{
		tc2BlockHash.String(): tc2BlockProposeInfo,
	}

	tc3ProposeTime := int64(1626755704)
	tc3BlockHeight := uint64(10)
	tc3BlockHash := common.HashH([]byte("1"))
	tc3BestViewHeight := uint64(8)
	tc3BestViewHash := common.HashH([]byte("2"))
	tc3FinalViewHeight := uint64(8)
	tc3Block := &mocksTypes.BlockInterface{}
	tc3Block.On("GetProposeTime").Return(tc3ProposeTime).Times(2)
	tc3Block.On("GetProduceTime").Return(tc3ProposeTime).Times(2)
	tc3Block.On("GetHeight").Return(tc3BlockHeight).Times(4)
	tc3Block.On("Hash").Return(&tc3BlockHash).Times(4)
	tc3CurrentTimeSlot := common.CalculateTimeSlot(tc3ProposeTime)
	tc3BlockProposeInfo := &ProposeBlockInfo{
		block:            tc3Block,
		IsVoted:          true,
		LastValidateTime: oldTimeList,
	}
	tc3ReceiveBlockByHash := map[string]*ProposeBlockInfo{
		tc3BlockHash.String(): tc3BlockProposeInfo,
	}

	tc4ProposeTime := int64(1626755704)
	tc4BlockHeight := uint64(10)
	tc4BlockHash := common.HashH([]byte("1"))
	tc4BestViewHeight := uint64(10)
	tc4BestViewHash := common.HashH([]byte("2"))
	tc4FinalViewHeight := uint64(8)
	tc4Block := &mocksTypes.BlockInterface{}
	tc4Block.On("GetProposeTime").Return(tc4ProposeTime).Times(2)
	tc4Block.On("GetProduceTime").Return(tc4ProposeTime + int64(common.TIMESLOT)).Times(2)
	tc4Block.On("GetHeight").Return(tc4BlockHeight + 1).Times(4)
	tc4Block.On("Hash").Return(&tc4BlockHash).Times(4)
	tc4CurrentTimeSlot := common.CalculateTimeSlot(tc4ProposeTime)
	tc4BlockProposeInfo := &ProposeBlockInfo{
		block:            tc4Block,
		IsVoted:          true,
		LastValidateTime: oldTimeList,
	}
	tc4ReceiveBlockByHash := map[string]*ProposeBlockInfo{
		tc4BlockHash.String(): tc4BlockProposeInfo,
	}

	tc5ProposeTime := int64(1626755704)
	tc5BlockHeight := uint64(10)
	tc5BlockHash := common.HashH([]byte("1"))
	tc5BestViewHeight := uint64(9)
	tc5BestViewHash := common.HashH([]byte("2"))
	tc5FinalViewHeight := uint64(11)
	tc5Block := &mocksTypes.BlockInterface{}
	tc5Block.On("GetProposeTime").Return(tc5ProposeTime).Times(2)
	tc5Block.On("GetProduceTime").Return(tc5ProposeTime).Times(2)
	tc5Block.On("GetHeight").Return(tc5BlockHeight).Times(4)
	tc5Block.On("Hash").Return(&tc5BlockHash).Times(4)
	tc5Block.On("GetPrevHash").Return(tc5BlockHash).Times(10)
	tc5Block.On("GetAggregateRootHash").Return(tc5BlockHash).Times(10)
	tc5Block.On("GetFinalityHeight").Return(uint64(0)).Times(10)
	tc5CurrentTimeSlot := common.CalculateTimeSlot(tc5ProposeTime)
	tc5BlockProposeInfo := &ProposeBlockInfo{
		block:              tc5Block,
		IsVoted:            true,
		LastValidateTime:   oldTimeList,
		IsValidLemma2Proof: false,
	}
	tc5ReceiveBlockByHash := map[string]*ProposeBlockInfo{
		tc5BlockHash.String(): tc5BlockProposeInfo,
	}

	tc6ProposeTime := int64(1626755704)
	tc6BlockHeight := uint64(10)
	tc6BestViewHeight := uint64(10)
	tc6BestViewHash := common.HashH([]byte("2"))
	tc6FinalViewHeight := uint64(8)
	tc6BlockHash := common.HashH([]byte("1"))
	tc6BlockHash2 := common.HashH([]byte("2"))
	tc6Block := &mocksTypes.BlockInterface{}
	tc6Block.On("GetProposeTime").Return(tc6ProposeTime).Times(10)
	tc6Block.On("GetProduceTime").Return(tc6ProposeTime).Times(10)
	tc6Block.On("GetHeight").Return(tc6BlockHeight + 1).Times(10)
	tc6Block.On("Hash").Return(&tc6BlockHash).Times(10)
	tc6Block.On("GetPrevHash").Return(tc6BlockHash).Times(10)
	tc6Block.On("GetAggregateRootHash").Return(tc6BlockHash).Times(10)
	tc6Block.On("GetFinalityHeight").Return(uint64(tc6BlockHeight)).Times(10)

	tc6Block2 := &mocksTypes.BlockInterface{}
	tc6Block2.On("GetProposeTime").Return(tc6ProposeTime).Times(10)
	tc6Block2.On("GetProduceTime").Return(tc6ProposeTime - int64(common.TIMESLOT)).Times(10)
	tc6Block2.On("GetHeight").Return(tc6BlockHeight + 1).Times(10)
	tc6Block2.On("Hash").Return(&tc6BlockHash2).Times(10)
	tc6Block2.On("GetPrevHash").Return(tc6BlockHash2).Times(10)
	tc6Block2.On("GetAggregateRootHash").Return(tc6BlockHash2).Times(10)
	tc6Block2.On("GetFinalityHeight").Return(uint64(0)).Times(10)

	tc6CurrentTimeSlot := common.CalculateTimeSlot(tc6ProposeTime)
	tc6BlockProposeInfo := &ProposeBlockInfo{
		block:              tc6Block,
		IsVoted:            true,
		LastValidateTime:   oldTimeList,
		IsValidLemma2Proof: true,
	}
	tc6BlockProposeInfo2 := &ProposeBlockInfo{
		block:              tc6Block2,
		IsVoted:            true,
		LastValidateTime:   oldTimeList,
		IsValidLemma2Proof: false,
	}
	tc6ReceiveBlockByHash := map[string]*ProposeBlockInfo{
		tc6BlockHash.String():  tc6BlockProposeInfo,
		tc6BlockHash2.String(): tc6BlockProposeInfo2,
	}

	tc7ProposeTime := int64(1626755704)
	tc7BestViewHeight := uint64(10)
	tc7BestViewHash := common.HashH([]byte("1"))
	tc7FinalViewHeight := uint64(8)
	tc7BlockHeight := uint64(10)
	tc7BlockHash := common.HashH([]byte("1"))
	tc7Block := &mocksTypes.BlockInterface{}
	tc7Block.On("GetProposeTime").Return(tc7ProposeTime).Times(2)
	tc7Block.On("GetProduceTime").Return(tc7ProposeTime).Times(2)
	tc7Block.On("GetHeight").Return(tc7BlockHeight).Times(4)
	tc7Block.On("Hash").Return(&tc7BlockHash).Times(4)
	tc7CurrentTimeSlot := common.CalculateTimeSlot(tc7ProposeTime)
	tc7BlockProposeInfo := &ProposeBlockInfo{
		block:            tc7Block,
		IsVoted:          false,
		LastValidateTime: time.Now(),
	}
	tc7ReceiveBlockByHash := map[string]*ProposeBlockInfo{
		tc7BlockHash.String(): tc7BlockProposeInfo,
	}

	tc8ProposeTime := int64(1626755704)
	tc8BlockHeight := uint64(10)
	tc8BestViewHeight := uint64(10)
	tc8BestViewHash := common.HashH([]byte("2"))
	tc8FinalViewHeight := uint64(8)
	tc8BlockHash := common.HashH([]byte("1"))
	tc8Block := &mocksTypes.BlockInterface{}
	tc8Block.On("GetProposeTime").Return(tc8ProposeTime).Times(10)
	tc8Block.On("GetProduceTime").Return(tc8ProposeTime).Times(10)
	tc8Block.On("GetHeight").Return(tc8BlockHeight + 1).Times(10)
	tc8Block.On("Hash").Return(&tc8BlockHash).Times(10)
	tc8Block.On("GetPrevHash").Return(&tc8BlockHash).Times(10)
	tc8Block.On("GetAggregateRootHash").Return(&tc8BlockHash).Times(10)
	tc8Block.On("GetFinalityHeight").Return(uint64(0)).Times(10)

	tc8CurrentTimeSlot := common.CalculateTimeSlot(tc8ProposeTime)
	tc8BlockProposeInfo := &ProposeBlockInfo{
		block:              tc8Block,
		IsVoted:            true,
		LastValidateTime:   oldTimeList,
		IsValidLemma2Proof: true,
	}
	tc8ReceiveBlockByHash := map[string]*ProposeBlockInfo{
		tc8BlockHash.String(): tc8BlockProposeInfo,
	}

	tc9ProposeTime := int64(1626755704)
	tc9BlockHeight := uint64(10)
	tc9BestViewHeight := uint64(10)
	tc9BestViewHash := common.HashH([]byte("2"))
	tc9FinalViewHeight := uint64(8)
	tc9BlockHash := common.HashH([]byte("1"))
	tc9Block := &mocksTypes.BlockInterface{}
	tc9Block.On("GetProposeTime").Return(tc9ProposeTime).Times(10)
	tc9Block.On("GetProduceTime").Return(tc9ProposeTime).Times(10)
	tc9Block.On("GetHeight").Return(tc9BlockHeight + 1).Times(10)
	tc9Block.On("Hash").Return(&tc9BlockHash).Times(10)
	tc9Block.On("GetPrevHash").Return(tc9BlockHash).Times(10)
	tc9Block.On("GetAggregateRootHash").Return(tc9BlockHash).Times(10)
	tc9Block.On("GetFinalityHeight").Return(uint64(tc9BlockHeight)).Times(10)

	tc9CurrentTimeSlot := common.CalculateTimeSlot(tc9ProposeTime)
	tc9BlockProposeInfo := &ProposeBlockInfo{
		block:              tc9Block,
		IsVoted:            true,
		LastValidateTime:   oldTimeList,
		IsValidLemma2Proof: false,
	}
	tc9ReceiveBlockByHash := map[string]*ProposeBlockInfo{
		tc9BlockHash.String(): tc9BlockProposeInfo,
	}

	type fields struct {
		logger common.Logger
		chain  Chain
	}
	type args struct {
		bestViewHash      common.Hash
		bestViewHeight    uint64
		finalViewHeight   uint64
		currentTimeSlot   int64
		proposeBlockInfos map[string]*ProposeBlockInfo
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantValid   []*ProposeBlockInfo
		wantReVote  []*ProposeBlockInfo
		wantInvalid []string
	}{
		{
			name:   "tc1: just validate recently",
			fields: fields{},
			args: args{
				bestViewHash:      tc1BestViewHash,
				bestViewHeight:    tc1BestViewHeight,
				finalViewHeight:   tc1FinalViewHeight,
				currentTimeSlot:   tc1CurrentTimeSlot,
				proposeBlockInfos: tc1ReceiveBlockByHash,
			},
			wantValid:   []*ProposeBlockInfo{},
			wantReVote:  []*ProposeBlockInfo{},
			wantInvalid: []string{},
		},
		{
			name:   "tc2: not propose in time slot",
			fields: fields{},
			args: args{
				bestViewHash:      tc2BestViewHash,
				bestViewHeight:    tc2BestViewHeight,
				finalViewHeight:   tc2FinalViewHeight,
				currentTimeSlot:   tc2CurrentTimeSlot,
				proposeBlockInfos: tc2ReceiveBlockByHash,
			},
			wantValid:   []*ProposeBlockInfo{},
			wantReVote:  []*ProposeBlockInfo{},
			wantInvalid: []string{},
		},
		{
			name:   "tc3: not connect to best height",
			fields: fields{},
			args: args{
				bestViewHash:      tc3BestViewHash,
				bestViewHeight:    tc3BestViewHeight,
				finalViewHeight:   tc3FinalViewHeight,
				currentTimeSlot:   tc3CurrentTimeSlot,
				proposeBlockInfos: tc3ReceiveBlockByHash,
			},
			wantValid:   []*ProposeBlockInfo{},
			wantReVote:  []*ProposeBlockInfo{},
			wantInvalid: []string{},
		},
		{
			name:   "tc4: producer time < current timeslot",
			fields: fields{},
			args: args{
				bestViewHash:      tc4BestViewHash,
				bestViewHeight:    tc4BestViewHeight,
				finalViewHeight:   tc4FinalViewHeight,
				currentTimeSlot:   tc4CurrentTimeSlot,
				proposeBlockInfos: tc4ReceiveBlockByHash,
			},
			wantValid:   []*ProposeBlockInfo{},
			wantReVote:  []*ProposeBlockInfo{},
			wantInvalid: []string{},
		},
		{
			name:   "tc5: propose block info height < final view",
			fields: fields{},
			args: args{
				bestViewHash:      tc5BestViewHash,
				bestViewHeight:    tc5BestViewHeight,
				finalViewHeight:   tc5FinalViewHeight,
				currentTimeSlot:   tc5CurrentTimeSlot,
				proposeBlockInfos: tc5ReceiveBlockByHash,
			},
			wantValid:   []*ProposeBlockInfo{},
			wantReVote:  []*ProposeBlockInfo{},
			wantInvalid: []string{tc5BlockHash.String()},
		},
		{
			name:   "tc6: add valid propose block info",
			fields: fields{},
			args: args{
				bestViewHash:      tc6BestViewHash,
				bestViewHeight:    tc6BestViewHeight,
				finalViewHeight:   tc6FinalViewHeight,
				currentTimeSlot:   tc6CurrentTimeSlot,
				proposeBlockInfos: tc6ReceiveBlockByHash,
			},
			wantValid:   []*ProposeBlockInfo{tc6BlockProposeInfo2, tc6BlockProposeInfo},
			wantReVote:  []*ProposeBlockInfo{},
			wantInvalid: []string{},
		},
		{
			name:   "tc7: re-vote insert block",
			fields: fields{},
			args: args{
				bestViewHash:      tc7BestViewHash,
				bestViewHeight:    tc7BestViewHeight,
				finalViewHeight:   tc7FinalViewHeight,
				currentTimeSlot:   tc7CurrentTimeSlot,
				proposeBlockInfos: tc7ReceiveBlockByHash,
			},
			wantValid:   []*ProposeBlockInfo{},
			wantReVote:  []*ProposeBlockInfo{tc7BlockProposeInfo},
			wantInvalid: []string{},
		},
		{
			name:   "tc8: lemma2 = true but finality height = 0",
			fields: fields{},
			args: args{
				bestViewHash:      tc8BestViewHash,
				bestViewHeight:    tc8BestViewHeight,
				finalViewHeight:   tc8FinalViewHeight,
				currentTimeSlot:   tc8CurrentTimeSlot,
				proposeBlockInfos: tc8ReceiveBlockByHash,
			},
			wantValid:   []*ProposeBlockInfo{},
			wantReVote:  []*ProposeBlockInfo{},
			wantInvalid: []string{},
		},
		{
			name:   "tc9: lemma2 = false but finality height != 0",
			fields: fields{},
			args: args{
				bestViewHash:      tc9BestViewHash,
				bestViewHeight:    tc9BestViewHeight,
				finalViewHeight:   tc9FinalViewHeight,
				currentTimeSlot:   tc9CurrentTimeSlot,
				proposeBlockInfos: tc9ReceiveBlockByHash,
			},
			wantValid:   []*ProposeBlockInfo{},
			wantReVote:  []*ProposeBlockInfo{},
			wantInvalid: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := ConsensusValidatorLemma2{
				logger: logger,
				chain:  tt.fields.chain,
			}
			got, got1, got2 := c.FilterValidProposeBlockInfo(tt.args.bestViewHash, tt.args.bestViewHeight, tt.args.finalViewHeight, tt.args.currentTimeSlot, tt.args.proposeBlockInfos)
			if !reflect.DeepEqual(got, tt.wantValid) {
				t.Errorf("FilterValidProposeBlockInfo() got = %v, want %v", got, tt.wantValid)
			}
			if !reflect.DeepEqual(got1, tt.wantReVote) {
				t.Errorf("FilterValidProposeBlockInfo() got1 = %v, want %v", got1, tt.wantReVote)
			}
			if !reflect.DeepEqual(got2, tt.wantInvalid) {
				t.Errorf("FilterValidProposeBlockInfo() got2 = %v, want %v", got2, tt.wantInvalid)
			}
		})
	}
}
