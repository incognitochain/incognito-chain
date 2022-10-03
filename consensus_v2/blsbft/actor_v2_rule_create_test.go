package blsbft

//
//import (
//	"github.com/incognitochain/incognito-chain/blockchain/types"
//	mocksTypes "github.com/incognitochain/incognito-chain/blockchain/types/mocks"
//	"github.com/incognitochain/incognito-chain/common"
//	"github.com/incognitochain/incognito-chain/consensus_v2/blsbft/mocks"
//	"github.com/incognitochain/incognito-chain/incognitokey"
//	"reflect"
//	"testing"
//	"time"
//)
//
//func TestNormalCreateBlockRule_CreateBlock(t *testing.T) {
//	tc1Hash := common.HashH([]byte{1})
//	tc1CurrentTime := time.Now().Unix()
//	tc1CommitteeViewHash := common.HashH([]byte("view-hash-1"))
//	tc1OutputBlock := &mocksTypes.BlockInterface{}
//	tc1OutputBlock.On("GetHeight").Return(uint64(2))
//	tc1OutputBlock.On("GetHeight").Return(uint64(2))
//	tc1OutputBlock.On("Hash").Return(&tc1Hash)
//	tc1Chain := &mocks.Chain{}
//	tc1Chain.On("CreateNewBlock", types.BLOCK_PRODUCINGV3_VERSION,
//		subset0Shard0CommitteeString[0], 1, tc1CurrentTime, shard0Committee, tc1CommitteeViewHash).
//		Return(tc1OutputBlock, nil)
//
//	tc2Hash := common.HashH([]byte{2})
//	tc2CurrentTime := time.Now().Unix()
//	tc2ProposeTime := int64(1626755704)
//	tc2CommitteeViewHashOld := common.HashH([]byte("view-hash-2-old"))
//	tc2CommitteeViewHashNew := common.HashH([]byte("view-hash-2-new"))
//	tc2InputBlock := &mocksTypes.BlockInterface{}
//	tc2InputBlock.On("CommitteeFromBlock").Return(tc2CommitteeViewHashOld).Times(2)
//	tc2InputBlock.On("GetVersion").Return(types.BLOCK_PRODUCINGV3_VERSION).Times(2)
//	tc2InputBlock.On("GetProposeTime").Return(tc2ProposeTime).Times(2)
//	tc2OutputBlock := &mocksTypes.BlockInterface{}
//	tc2OutputBlock.On("GetHeight").Return(uint64(2))
//	tc2OutputBlock.On("Hash").Return(&tc2Hash)
//	tc2Chain := &mocks.Chain{}
//	tc2Chain.On("IsBeaconChain").Return(false)
//	tc2Chain.On("CreateNewBlock", types.BLOCK_PRODUCINGV3_VERSION,
//		subset1Shard0CommitteeStringNew[0], 1, tc2CurrentTime, shard0CommitteeNew, tc2CommitteeViewHashNew).
//		Return(tc2OutputBlock, nil)
//	tc2Chain.On("GetProposerByTimeSlotFromCommitteeList", common.CalculateTimeSlot(tc2ProposeTime), shard0Committee).
//		Return(shard0Committee[2], 2, nil)
//	tc2Chain.On("GetSigningCommittees", int(2), shard0Committee, types.BLOCK_PRODUCINGV3_VERSION).
//		Return(subset0Shard0Committee)
//
//	tc3Hash := common.HashH([]byte{3})
//	tc3CurrentTime := time.Now().Unix()
//	tc3ProposeTime := int64(1626755704)
//	tc3CommitteeViewHash := common.HashH([]byte("view-hash-3"))
//	tc3InputBlock := &mocksTypes.BlockInterface{}
//	tc3InputBlock.On("CommitteeFromBlock").Return(tc3CommitteeViewHash).Times(2)
//	tc3InputBlock.On("GetVersion").Return(types.BLOCK_PRODUCINGV3_VERSION).Times(2)
//	tc3InputBlock.On("GetProposeTime").Return(tc3ProposeTime).Times(2)
//	tc3InputBlock.On("GetProducer").Return(subset0Shard0CommitteeString[1]).Times(2)
//	tc3InputBlock.On("GetHeight").Return(uint64(2)).Times(2)
//	tc3InputBlock.On("Hash").Return(&tc3Hash).Times(2)
//	tc3OutputBlock := &mocksTypes.BlockInterface{}
//	tc3OutputBlock.On("GetHeight").Return(uint64(2))
//	tc3OutputBlock.On("Hash").Return(&tc3Hash)
//	tc3Chain := &mocks.Chain{}
//	tc3Chain.On("IsBeaconChain").Return(false)
//	tc3Chain.On("CreateNewBlockFromOldBlock", tc3InputBlock,
//		subset0Shard0CommitteeString[2], tc3CurrentTime, false).
//		Return(tc3OutputBlock, nil)
//	tc3Chain.On("GetProposerByTimeSlotFromCommitteeList", common.CalculateTimeSlot(tc3ProposeTime), shard0Committee).
//		Return(shard0Committee[2], 2, nil)
//	tc3Chain.On("GetSigningCommittees", int(2), shard0Committee, types.BLOCK_PRODUCINGV3_VERSION).
//		Return(subset0Shard0Committee)
//
//	tc4Hash := common.HashH([]byte{4})
//	tc4CurrentTime := time.Now().Unix()
//	tc4ProposeTime := int64(1626755704)
//	tc4CommitteeViewHash := common.HashH([]byte("view-hash-4"))
//	tc4InputBlock := &mocksTypes.BlockInterface{}
//	tc4InputBlock.On("CommitteeFromBlock").Return(tc4CommitteeViewHash).Times(2)
//	tc4InputBlock.On("GetVersion").Return(types.BLOCK_PRODUCINGV3_VERSION).Times(2)
//	tc4InputBlock.On("GetProposeTime").Return(tc4ProposeTime).Times(2)
//	tc4InputBlock.On("GetProducer").Return(subset0Shard0CommitteeString[1]).Times(2)
//	tc4InputBlock.On("GetHeight").Return(uint64(2)).Times(2)
//	tc4InputBlock.On("Hash").Return(&tc4Hash).Times(2)
//	tc4CommitteeChainHandler := &mocks.CommitteeChainHandler{}
//	tc4CommitteeChainHandler.On("CommitteesFromViewHashForShard", tc4CommitteeViewHash, byte(0)).
//		Return(shard0Committee, nil)
//	tc4OutputBlock := &mocksTypes.BlockInterface{}
//	tc4OutputBlock.On("GetHeight").Return(uint64(2))
//	tc4OutputBlock.On("Hash").Return(&tc4Hash)
//	tc4Chain := &mocks.Chain{}
//	tc4Chain.On("IsBeaconChain").Return(false)
//	tc4Chain.On("CreateNewBlockFromOldBlock", tc4InputBlock,
//		subset1Shard0CommitteeString[0], tc4CurrentTime, false).
//		Return(tc4OutputBlock, nil)
//	tc4Chain.On("GetProposerByTimeSlotFromCommitteeList", common.CalculateTimeSlot(tc4ProposeTime), shard0Committee).
//		Return(shard0Committee[2], 2, nil)
//	tc4Chain.On("GetSigningCommittees", int(2), shard0Committee, types.BLOCK_PRODUCINGV3_VERSION).
//		Return(subset0Shard0Committee)
//
//	tc5Hash := common.HashH([]byte{1})
//	tc5CurrentTime := time.Now().Unix()
//	tc5CommitteeViewHash := common.HashH([]byte("view-hash-2"))
//	tc5OutputBlock := &mocksTypes.BlockInterface{}
//	tc5OutputBlock.On("GetHeight").Return(uint64(4))
//	tc5OutputBlock.On("Hash").Return(&tc5Hash)
//	tc5OutputBlock.On("GetFinalityHeight").Return(uint64(3))
//	tc5Chain := &mocks.Chain{}
//	tc5Chain.On("CreateNewBlock", types.LEMMA2_VERSION,
//		subset0Shard0CommitteeString[0], 1, tc5CurrentTime, shard0Committee, tc5CommitteeViewHash).
//		Return(tc5OutputBlock, nil)
//
//	tc6Hash := common.HashH([]byte{6})
//	tc6CurrentTime := time.Now().Unix()
//	tc6ProposeTime := int64(1626755704)
//	tc6CommitteeViewHash := common.HashH([]byte("view-hash-6"))
//	tc6InputBlock := &mocksTypes.BlockInterface{}
//	tc6InputBlock.On("CommitteeFromBlock").Return(tc6CommitteeViewHash).Times(2)
//	tc6InputBlock.On("GetVersion").Return(types.LEMMA2_VERSION).Times(2)
//	tc6InputBlock.On("GetProposeTime").Return(tc6ProposeTime).Times(2)
//	tc6InputBlock.On("GetProducer").Return(subset0Shard0CommitteeString[1]).Times(2)
//	tc6InputBlock.On("GetHeight").Return(uint64(2)).Times(4)
//	tc6InputBlock.On("Hash").Return(&tc6Hash).Times(2)
//	tc6OutputBlock := &mocksTypes.BlockInterface{}
//	tc6OutputBlock.On("GetHeight").Return(uint64(4))
//	tc6OutputBlock.On("Hash").Return(&tc6Hash)
//	tc6OutputBlock.On("GetFinalityHeight").Return(uint64(3))
//	tc6Chain := &mocks.Chain{}
//	tc6Chain.On("IsBeaconChain").Return(false)
//	tc6Chain.On("CreateNewBlockFromOldBlock", tc6InputBlock,
//		subset0Shard0CommitteeString[0], tc6CurrentTime, true).
//		Return(tc6OutputBlock, nil)
//	tc6Chain.On("GetProposerByTimeSlotFromCommitteeList", common.CalculateTimeSlot(tc6ProposeTime), shard0Committee).
//		Return(shard0Committee[2], 2, nil)
//	tc6Chain.On("GetSigningCommittees", int(2), shard0Committee, types.LEMMA2_VERSION).
//		Return(subset0Shard0Committee)
//
//	type fields struct {
//		logger common.Logger
//		chain  Chain
//	}
//	type args struct {
//		b58Str            string
//		block             types.BlockInterface
//		committees        []incognitokey.CommitteePublicKey
//		committeeViewHash common.Hash
//		isValidRePropose  bool
//		consensusName     string
//		blockVersion      int
//		currentTime       int64
//		isRePropose       bool
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    types.BlockInterface
//		wantErr bool
//	}{
//		{
//			name: "Propose a new block when found no block to re-propose",
//			fields: fields{
//				chain:  tc1Chain,
//				logger: logger,
//			},
//			args: args{
//				b58Str:            subset0Shard0CommitteeString[0],
//				block:             nil,
//				committees:        shard0Committee,
//				committeeViewHash: tc1CommitteeViewHash,
//				isValidRePropose:  false,
//				consensusName:     common.BlsConsensus,
//				blockVersion:      types.BLOCK_PRODUCINGV3_VERSION,
//				currentTime:       tc1CurrentTime,
//				isRePropose:       false,
//			},
//			want:    tc1OutputBlock,
//			wantErr: false,
//		},
//		{
//			name: "Propose a new block when found block to re-propose but different committees view",
//			fields: fields{
//				chain:  tc2Chain,
//				logger: logger,
//			},
//			args: args{
//				b58Str:            subset1Shard0CommitteeStringNew[0],
//				block:             tc2InputBlock,
//				committees:        shard0CommitteeNew,
//				committeeViewHash: tc2CommitteeViewHashNew,
//				isValidRePropose:  false,
//				consensusName:     common.BlsConsensus,
//				blockVersion:      types.BLOCK_PRODUCINGV3_VERSION,
//				currentTime:       tc1CurrentTime,
//				isRePropose:       false,
//			},
//			want:    tc2OutputBlock,
//			wantErr: false,
//		},
//		{
//			name: "Re-propose block from the same subset, finality = 0",
//			fields: fields{
//				chain:  tc3Chain,
//				logger: logger,
//			},
//			args: args{
//				b58Str:            subset0Shard0CommitteeString[2],
//				block:             tc3InputBlock,
//				committees:        shard0Committee,
//				committeeViewHash: tc3CommitteeViewHash,
//				isValidRePropose:  false,
//				consensusName:     common.BlsConsensus,
//				blockVersion:      types.BLOCK_PRODUCINGV3_VERSION,
//				currentTime:       tc1CurrentTime,
//				isRePropose:       true,
//			},
//			want:    tc3OutputBlock,
//			wantErr: false,
//		},
//		{
//			name: "Re-propose block from the different subset, finality = 0",
//			fields: fields{
//				chain:  tc4Chain,
//				logger: logger,
//			},
//			args: args{
//				b58Str:            subset1Shard0CommitteeString[0],
//				block:             tc4InputBlock,
//				committees:        shard0Committee,
//				committeeViewHash: tc4CommitteeViewHash,
//				isValidRePropose:  false,
//				consensusName:     common.BlsConsensus,
//				blockVersion:      types.BLOCK_PRODUCINGV3_VERSION,
//				currentTime:       tc1CurrentTime,
//				isRePropose:       true,
//			},
//			want:    tc4OutputBlock,
//			wantErr: false,
//		},
//		{
//			name: "Create new lemma 2 block",
//			fields: fields{
//				chain:  tc5Chain,
//				logger: logger,
//			},
//			args: args{
//				b58Str:            subset0Shard0CommitteeString[0],
//				block:             nil,
//				committees:        shard0Committee,
//				committeeViewHash: tc5CommitteeViewHash,
//				isValidRePropose:  false,
//				consensusName:     common.BlsConsensus,
//				blockVersion:      types.LEMMA2_VERSION,
//				currentTime:       tc1CurrentTime,
//				isRePropose:       false,
//			},
//			want:    tc5OutputBlock,
//			wantErr: false,
//		},
//		{
//			name: "Repropose block with lemma 2 block, finality > 0",
//			fields: fields{
//				chain:  tc6Chain,
//				logger: logger,
//			},
//			args: args{
//				b58Str:            subset0Shard0CommitteeString[0],
//				block:             tc6InputBlock,
//				committees:        shard0Committee,
//				committeeViewHash: tc6CommitteeViewHash,
//				isValidRePropose:  true,
//				consensusName:     common.BlsConsensus,
//				blockVersion:      types.LEMMA2_VERSION,
//				currentTime:       tc1CurrentTime,
//				isRePropose:       true,
//			},
//			want:    tc6OutputBlock,
//			wantErr: false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			n := NormalCreateBlockRule{
//				logger: tt.fields.logger,
//				chain:  tt.fields.chain,
//			}
//			got, err := n.CreateBlock(tt.args.b58Str, tt.args.block, tt.args.committees, tt.args.committeeViewHash, tt.args.isValidRePropose, tt.args.consensusName, tt.args.blockVersion, tt.args.currentTime, tt.args.isRePropose)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("CreateBlock() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("CreateBlock() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
