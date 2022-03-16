package blsbft

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
	mocksTypes "github.com/incognitochain/incognito-chain/blockchain/types/mocks"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/consensus_v2/blsbft/mocks"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"reflect"
	"testing"
)

func TestProposeRuleLemma2_HandleBFTProposeMessage(t *testing.T) {

	config.AbortParam()
	config.Param().CommitteeSize.NumberOfFixedShardBlockValidator = 8

	tc1PreviousBlock := &mocksTypes.BlockInterface{}
	tc1PreviousBlock.On("GetProduceTime").Return(int64(1632753240))
	tc1PreviousBlock.On("GetProposeTime").Return(int64(1632753250))
	tc1Block := &mocksTypes.BlockInterface{}
	tc1Block.On("GetProduceTime").Return(int64(1632753260))
	tc1Block.On("GetProposeTime").Return(int64(1632753260))
	tc1Block.On("GetProducer").Return(shard0CommitteeString[6])
	tc1Block.On("GetProposer").Return(shard0CommitteeString[6])
	tc1Block.On("GetPrevHash").Return(common.Hash{}.NewHashFromStr2("2fc3ec987e4d0a56e344a16dbba524845f5aca6b699b0767735f0ad59926ea03")).Times(10)
	tc1Block.On("GetAggregateRootHash").Return(common.Hash{}.NewHashFromStr2("698bed377b2d7c6c280924f6af6c46c7e84acddddc99268ac76e9393b3d3b05f"))
	tc1Block.On("GetFinalityHeight").Return(uint64(971))
	tc1Block.On("GetHeight").Return(uint64(972)).Times(64)
	tc1Hash := common.Hash{}.NewHashFromStr2("1b7ad1402f18233a1e7cdf81a7e817c83cce18f210bc10a8523155e9d5f27db4")
	tc1Block.On("Hash").Return(&tc1Hash).Times(64)

	tc1FinalityProof := make(map[string]map[int64]string)
	tc1Chain := &mocks.Chain{}
	tc1Env := &ProposeMessageEnvironment{
		block:                            tc1Block,
		previousBlock:                    tc1PreviousBlock,
		committees:                       shard0Committee,
		signingCommittees:                shard0Committee,
		userKeySet:                       []signatureschemes2.MiningKey{},
		producerPublicBLSMiningKey:       "",
		NumberOfFixedShardBlockValidator: 8,
	}
	tc1ProposeMsg := &BFTPropose{
		ReProposeHashSignature: "1ec352qhKtECAtECXP5pmdAtJyfCPSjjVCnL4TXZ3DQuTkSuvmT2yNBFtc4m9phexFXZqHzaoesmmC6SHYwEVDFeysuXfRH",
		FinalityProof: FinalityProof{
			ReProposeHashSignature: []string{},
		},
	}
	tc1ProposeBlockInfo := newProposeBlockForProposeMsgLemma2(
		tc1ProposeMsg,
		tc1Env.block,
		tc1Env.committees,
		tc1Env.signingCommittees,
		tc1Env.userKeySet,
		tc1Env.producerPublicBLSMiningKey,
		true,
	)

	tc2PreviousBlock := &mocksTypes.BlockInterface{}
	tc2PreviousBlock.On("GetProduceTime").Return(int64(1632753240))
	tc2PreviousBlock.On("GetProposeTime").Return(int64(1632753250))
	tc2Block := &mocksTypes.BlockInterface{}
	tc2Block.On("GetProduceTime").Return(int64(1632753260))
	tc2Block.On("GetProposeTime").Return(int64(1632753260))
	tc2Block.On("GetProducer").Return(shard0CommitteeString[6])
	tc2Block.On("GetProposer").Return(shard0CommitteeString[6])
	tc2Block.On("GetPrevHash").Return(common.Hash{}.NewHashFromStr2("2fc3ec987e4d0a56e344a16dbba524845f5aca6b699b0767735f0ad59926ea03")).Times(10)
	tc2Block.On("GetAggregateRootHash").Return(common.Hash{}.NewHashFromStr2("698bed377b2d7c6c280924f6af6c46c7e84acddddc99268ac76e9393b3d3b05f"))
	tc2Block.On("GetFinalityHeight").Return(uint64(0))
	tc2Block.On("GetHeight").Return(uint64(972)).Times(64)
	tc2Hash := common.Hash{}.NewHashFromStr2("1b7ad1402f18233a1e7cdf81a7e817c83cce18f210bc10a8523155e9d5f27db4")
	tc2Block.On("Hash").Return(&tc2Hash).Times(64)

	tc2FinalityProof := make(map[string]map[int64]string)
	tc2Chain := &mocks.Chain{}
	tc2Env := &ProposeMessageEnvironment{
		block:                            tc2Block,
		previousBlock:                    tc2PreviousBlock,
		committees:                       shard0Committee,
		signingCommittees:                shard0Committee,
		userKeySet:                       []signatureschemes2.MiningKey{},
		producerPublicBLSMiningKey:       "",
		NumberOfFixedShardBlockValidator: 8,
	}
	tc2ProposeMsg := &BFTPropose{
		ReProposeHashSignature: "1ec352qhKtECAtECXP5pmdAtJyfCPSjjVCnL4TXZ3DQuTkSuvmT2yNBFtc4m9phexFXZqHzaoesmmC6SHYwEVDFeysuXfRH",
		FinalityProof: FinalityProof{
			ReProposeHashSignature: []string{},
		},
	}

	tc3PreviousBlock := &mocksTypes.BlockInterface{}
	tc3PreviousBlock.On("GetProduceTime").Return(int64(1632753240))
	tc3PreviousBlock.On("GetProposeTime").Return(int64(1632753250))
	tc3Block := &mocksTypes.BlockInterface{}
	tc3Block.On("GetProduceTime").Return(int64(1632753260))
	tc3Block.On("GetProposeTime").Return(int64(1632753880))
	tc3Block.On("GetProducer").Return(shard0CommitteeString[6])
	tc3Block.On("GetProposer").Return(shard0CommitteeString[4])
	tc3Block.On("GetPrevHash").Return(common.Hash{}.NewHashFromStr2("2fc3ec987e4d0a56e344a16dbba524845f5aca6b699b0767735f0ad59926ea03")).Times(10)
	tc3Block.On("GetAggregateRootHash").Return(common.Hash{}.NewHashFromStr2("698bed377b2d7c6c280924f6af6c46c7e84acddddc99268ac76e9393b3d3b05f"))
	tc3Block.On("GetFinalityHeight").Return(uint64(971))
	tc3Block.On("GetHeight").Return(uint64(972)).Times(100)
	tc3Hash := common.Hash{}.NewHashFromStr2("1b7ad1402f18233a1e7cdf81a7e817c83cce18f210bc10a8523155e9d5f27db4")
	tc3Block.On("Hash").Return(&tc3Hash).Times(64)

	tc3FinalityProof := make(map[string]map[int64]string)
	tc3Chain := &mocks.Chain{}
	tc3Env := &ProposeMessageEnvironment{
		block:                            tc3Block,
		previousBlock:                    tc3PreviousBlock,
		committees:                       shard0Committee,
		signingCommittees:                shard0Committee,
		userKeySet:                       []signatureschemes2.MiningKey{},
		producerPublicBLSMiningKey:       "",
		NumberOfFixedShardBlockValidator: 8,
	}
	tc3ProposeMsg := &BFTPropose{
		ReProposeHashSignature: "12Xzv92cdgguVCNzAFS7LANpXyXXdWAYDXFJssQgGK5236q8HPzW8RQKPmrEhVeP6A9fCvswfKihKWwKBURrzSaPZ7H5pnLm",
		FinalityProof: FinalityProof{
			ReProposeHashSignature: []string{
				"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",
			},
		},
	}
	tc3ProposeBlockInfo := newProposeBlockForProposeMsgLemma2(
		tc3ProposeMsg,
		tc3Env.block,
		tc3Env.committees,
		tc3Env.signingCommittees,
		tc3Env.userKeySet,
		tc3Env.producerPublicBLSMiningKey,
		true,
	)

	tc4PreviousBlock := &mocksTypes.BlockInterface{}
	tc4PreviousBlock.On("GetProduceTime").Return(int64(1632753240))
	tc4PreviousBlock.On("GetProposeTime").Return(int64(1632753250))
	tc4Block := &mocksTypes.BlockInterface{}
	tc4Block.On("GetProduceTime").Return(int64(1632753260))
	tc4Block.On("GetProposeTime").Return(int64(1632753880))
	tc4Block.On("GetProducer").Return(shard0CommitteeString[6])
	tc4Block.On("GetProposer").Return(shard0CommitteeString[4])
	tc4Block.On("GetPrevHash").Return(common.Hash{}.NewHashFromStr2("2fc3ec987e4d0a56e344a16dbba524845f5aca6b699b0767735f0ad59926ea03")).Times(10)
	tc4Block.On("GetAggregateRootHash").Return(common.Hash{}.NewHashFromStr2("698bed377b2d7c6c280924f6af6c46c7e84acddddc99268ac76e9393b3d3b05f"))
	tc4Block.On("GetFinalityHeight").Return(uint64(0))
	tc4Block.On("GetHeight").Return(uint64(972)).Times(100)
	tc4Hash := common.Hash{}.NewHashFromStr2("1b7ad1402f18233a1e7cdf81a7e817c83cce18f210bc10a8523155e9d5f27db4")
	tc4Block.On("Hash").Return(&tc4Hash).Times(64)

	tc4FinalityProof := make(map[string]map[int64]string)
	tc4Chain := &mocks.Chain{}
	tc4Env := &ProposeMessageEnvironment{
		block:                            tc4Block,
		previousBlock:                    tc4PreviousBlock,
		committees:                       shard0Committee,
		signingCommittees:                shard0Committee,
		userKeySet:                       []signatureschemes2.MiningKey{},
		producerPublicBLSMiningKey:       "",
		NumberOfFixedShardBlockValidator: 8,
	}
	tc4ProposeMsg := &BFTPropose{
		ReProposeHashSignature: "12Xzv92cdgguVCNzAFS7LANpXyXXdWAYDXFJssQgGK5236q8HPzW8RQKPmrEhVeP6A9fCvswfKihKWwKBURrzSaPZ7H5pnLm",
		FinalityProof: FinalityProof{
			ReProposeHashSignature: []string{
				"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",
			},
		},
	}

	type fields struct {
		logger                 common.Logger
		nextBlockFinalityProof map[string]map[int64]string
		chain                  Chain
	}
	type args struct {
		env        *ProposeMessageEnvironment
		proposeMsg *BFTPropose
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *ProposeBlockInfo
		wantErr bool
	}{
		{
			name: "tc1: handle valid first create block",
			fields: fields{
				logger:                 logger,
				chain:                  tc1Chain,
				nextBlockFinalityProof: tc1FinalityProof,
			},
			args: args{
				env:        tc1Env,
				proposeMsg: tc1ProposeMsg,
			},
			want:    tc1ProposeBlockInfo,
			wantErr: false,
		},
		{
			name: "tc2: handle invalid first create block",
			fields: fields{
				logger:                 logger,
				chain:                  tc2Chain,
				nextBlockFinalityProof: tc2FinalityProof,
			},
			args: args{
				env:        tc2Env,
				proposeMsg: tc2ProposeMsg,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "tc3: handle valid re-propose block",
			fields: fields{
				logger:                 logger,
				chain:                  tc3Chain,
				nextBlockFinalityProof: tc3FinalityProof,
			},
			args: args{
				env:        tc3Env,
				proposeMsg: tc3ProposeMsg,
			},
			want:    tc3ProposeBlockInfo,
			wantErr: false,
		},
		{
			name: "tc4: handle invalid re-propose block",
			fields: fields{
				logger:                 logger,
				chain:                  tc4Chain,
				nextBlockFinalityProof: tc4FinalityProof,
			},
			args: args{
				env:        tc4Env,
				proposeMsg: tc4ProposeMsg,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ProposeRuleLemma2{
				logger:                 tt.fields.logger,
				nextBlockFinalityProof: tt.fields.nextBlockFinalityProof,
				chain:                  tt.fields.chain,
			}
			got, err := p.HandleBFTProposeMessage(tt.args.env, tt.args.proposeMsg)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleBFTProposeMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HandleBFTProposeMessage() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProposeRuleLemma2_isFirstBlockNextHeight(t *testing.T) {

	tc1PreviousBlock := &mocksTypes.BlockInterface{}
	tc1Block := &mocksTypes.BlockInterface{}
	tc1BlockProposeTime := int64(16267557040)
	tc1BlockProduceTime := int64(16267557040) + 10
	tc1Block.On("GetProposeTime").Return(tc1BlockProposeTime)
	tc1Block.On("GetProduceTime").Return(tc1BlockProduceTime)
	tc1FinalityProof := make(map[string]map[int64]string)
	tc1Chain := &mocks.Chain{}

	tc2PreviousBlock := &mocksTypes.BlockInterface{}
	tc2Block := &mocksTypes.BlockInterface{}
	tc2BlockProposeTime := int64(16267557040)
	tc2BlockProduceTime := int64(16267557040)
	tc2Block.On("GetProposeTime").Return(tc2BlockProposeTime)
	tc2Block.On("GetProduceTime").Return(tc2BlockProduceTime)
	tc2Block.On("GetProposer").Return(shard00)
	tc2Block.On("GetProducer").Return(shard01)
	tc2FinalityProof := make(map[string]map[int64]string)
	tc2Chain := &mocks.Chain{}

	tc3PreviousBlock := &mocksTypes.BlockInterface{}
	tc3Block := &mocksTypes.BlockInterface{}
	tc3BlockProposeTime := int64(16267557040)
	tc3BlockProduceTime := int64(16267557040)
	tc3Block.On("GetProposeTime").Return(tc3BlockProposeTime)
	tc3Block.On("GetProduceTime").Return(tc3BlockProduceTime)
	tc3Block.On("GetProposer").Return(shard00)
	tc3Block.On("GetProducer").Return(shard00)
	tc3PreviousBlock.On("GetProposeTime").Return(tc3BlockProduceTime - 20)
	tc3FinalityProof := make(map[string]map[int64]string)
	tc3Chain := &mocks.Chain{}

	tc4PreviousBlock := &mocksTypes.BlockInterface{}
	tc4Block := &mocksTypes.BlockInterface{}
	tc4BlockProposeTime := int64(16267557040)
	tc4BlockProduceTime := int64(16267557040)
	tc4Block.On("GetProposeTime").Return(tc4BlockProposeTime)
	tc4Block.On("GetProduceTime").Return(tc4BlockProduceTime).Times(2)
	tc4Block.On("GetProposer").Return(shard00)
	tc4Block.On("GetProducer").Return(shard00)
	tc4PreviousBlock.On("GetProposeTime").Return(tc4BlockProduceTime - 10)
	tc4FinalityProof := make(map[string]map[int64]string)
	tc4Chain := &mocks.Chain{}

	type fields struct {
		logger                 common.Logger
		nextBlockFinalityProof map[string]map[int64]string
		chain                  Chain
	}
	type args struct {
		previousBlock types.BlockInterface
		block         types.BlockInterface
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "tc1: block is re-propose",
			fields: fields{
				logger:                 logger,
				chain:                  tc1Chain,
				nextBlockFinalityProof: tc1FinalityProof,
			},
			args: args{
				previousBlock: tc1PreviousBlock,
				block:         tc1Block,
			},
			want: false,
		},
		{
			name: "tc2: block is first propose but proposer is not produce",
			fields: fields{
				logger:                 logger,
				chain:                  tc2Chain,
				nextBlockFinalityProof: tc2FinalityProof,
			},
			args: args{
				previousBlock: tc2PreviousBlock,
				block:         tc2Block,
			},
			want: false,
		},
		{
			name: "tc3: block is first propose but not next timeslot",
			fields: fields{
				logger:                 logger,
				chain:                  tc3Chain,
				nextBlockFinalityProof: tc3FinalityProof,
			},
			args: args{
				previousBlock: tc3PreviousBlock,
				block:         tc3Block,
			},
			want: false,
		},
		{
			name: "tc4: block is valid first propose",
			fields: fields{
				logger:                 logger,
				chain:                  tc4Chain,
				nextBlockFinalityProof: tc4FinalityProof,
			},
			args: args{
				previousBlock: tc4PreviousBlock,
				block:         tc4Block,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProposeRuleLemma2{
				logger:                 tt.fields.logger,
				nextBlockFinalityProof: tt.fields.nextBlockFinalityProof,
				chain:                  tt.fields.chain,
			}
			if got := p.isFirstBlockNextHeight(tt.args.previousBlock, tt.args.block); got != tt.want {
				t.Errorf("isFirstBlockNextHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProposeRuleLemma2_isReProposeFromFirstBlockNextHeight(t *testing.T) {

	config.AbortParam()
	config.Param().CommitteeSize.NumberOfFixedShardBlockValidator = 8

	tc1PreviousBlock := &mocksTypes.BlockInterface{}
	tc1PreviousBlockProposeTime := int64(16267557029)
	tc1PreviousBlock.On("GetProposeTime").Return(tc1PreviousBlockProposeTime)
	tc1Block := &mocksTypes.BlockInterface{}
	tc1BlockProposeTime := int64(16267557040)
	tc1BlockProduceTime := int64(16267557040)
	tc1Block.On("GetProposeTime").Return(tc1BlockProposeTime)
	tc1Block.On("GetProduceTime").Return(tc1BlockProduceTime)
	tc1Block.On("GetProposer").Return(shard0CommitteeString[0])
	tc1Block.On("GetProducer").Return(shard0CommitteeString[0])
	tc1FinalityProof := make(map[string]map[int64]string)
	tc1Chain := &mocks.Chain{}

	tc2PreviousBlock := &mocksTypes.BlockInterface{}
	tc2PreviousBlockProposeTime := int64(16267557031)
	tc2PreviousBlock.On("GetProposeTime").Return(tc2PreviousBlockProposeTime)
	tc2Block := &mocksTypes.BlockInterface{}
	tc2BlockProposeTime := int64(16267557040)
	tc2BlockProduceTime := int64(16267557040)
	tc2Block.On("GetProposeTime").Return(tc2BlockProposeTime)
	tc2Block.On("GetProduceTime").Return(tc2BlockProduceTime)
	tc2Block.On("GetProposer").Return(shard0CommitteeString[0])
	tc2Block.On("GetProducer").Return(shard0CommitteeString[0])
	tc2FinalityProof := make(map[string]map[int64]string)
	tc2Chain := &mocks.Chain{}

	tc3PreviousBlock := &mocksTypes.BlockInterface{}
	tc3PreviousBlockProposeTime := int64(16267557031)
	tc3PreviousBlock.On("GetProposeTime").Return(tc3PreviousBlockProposeTime)
	tc3Block := &mocksTypes.BlockInterface{}
	tc3BlockProposeTime := int64(16267557040)
	tc3BlockProduceTime := int64(16267557040)
	tc3Block.On("GetProposeTime").Return(tc3BlockProposeTime)
	tc3Block.On("GetProduceTime").Return(tc3BlockProduceTime)
	tc3Block.On("GetProposer").Return(shard0CommitteeString[0])
	tc3Block.On("GetProducer").Return(shard0CommitteeString[0])
	tc3FinalityProof := make(map[string]map[int64]string)
	tc3Chain := &mocks.Chain{}

	tc4PreviousBlock := &mocksTypes.BlockInterface{}
	tc4PreviousBlockProposeTime := int64(16267557031)
	tc4PreviousBlock.On("GetProposeTime").Return(tc4PreviousBlockProposeTime)
	tc4Block := &mocksTypes.BlockInterface{}
	tc4BlockProposeTime := int64(16267557050)
	tc4BlockProduceTime := int64(16267557040)
	tc4Block.On("GetProposeTime").Return(tc4BlockProposeTime)
	tc4Block.On("GetProduceTime").Return(tc4BlockProduceTime)
	tc4Block.On("GetProposer").Return(shard0CommitteeString[0])
	tc4Block.On("GetProducer").Return(shard0CommitteeString[0])
	tc4FinalityProof := make(map[string]map[int64]string)
	tc4Chain := &mocks.Chain{}
	tc4Committee := shard0Committee

	tc5PreviousBlock := &mocksTypes.BlockInterface{}
	tc5PreviousBlockProposeTime := int64(16267557031)
	tc5PreviousBlock.On("GetProposeTime").Return(tc5PreviousBlockProposeTime)
	tc5Block := &mocksTypes.BlockInterface{}
	tc5BlockProposeTime := int64(16267557050)
	tc5BlockProduceTime := int64(16267557040)
	tc5Block.On("GetProposeTime").Return(tc5BlockProposeTime)
	tc5Block.On("GetProduceTime").Return(tc5BlockProduceTime)
	tc5Block.On("GetProposer").Return(shard0CommitteeString[2])
	tc5Block.On("GetProducer").Return(shard0CommitteeString[0])
	tc5FinalityProof := make(map[string]map[int64]string)
	tc5Chain := &mocks.Chain{}
	tc5Committee := shard0Committee

	tc6PreviousBlock := &mocksTypes.BlockInterface{}
	tc6PreviousBlockProposeTime := int64(16267557031)
	tc6PreviousBlock.On("GetProposeTime").Return(tc6PreviousBlockProposeTime)
	tc6Block := &mocksTypes.BlockInterface{}
	tc6BlockProposeTime := int64(16267557063)
	tc6BlockProduceTime := int64(16267557040)
	tc6Block.On("GetProposeTime").Return(tc6BlockProposeTime)
	tc6Block.On("GetProduceTime").Return(tc6BlockProduceTime)
	tc6Block.On("GetProposer").Return(shard0CommitteeString[2])
	tc6Block.On("GetProducer").Return(shard0CommitteeString[0])
	tc6FinalityProof := make(map[string]map[int64]string)
	tc6Chain := &mocks.Chain{}
	tc6Committee := shard0Committee

	type fields struct {
		logger                 common.Logger
		nextBlockFinalityProof map[string]map[int64]string
		chain                  Chain
	}
	type args struct {
		chainID       int
		previousBlock types.BlockInterface
		block         types.BlockInterface
		committees    []incognitokey.CommitteePublicKey
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "tc1: block re propose is not next timeslot",
			fields: fields{
				logger:                 logger,
				chain:                  tc1Chain,
				nextBlockFinalityProof: tc1FinalityProof,
			},
			args: args{
				previousBlock: tc1PreviousBlock,
				block:         tc1Block,
				committees:    []incognitokey.CommitteePublicKey{},
			},
			want: false,
		},
		{
			name: "tc2: block re-propose but propose time <= produce time",
			fields: fields{
				logger:                 logger,
				chain:                  tc2Chain,
				nextBlockFinalityProof: tc2FinalityProof,
			},
			args: args{
				previousBlock: tc2PreviousBlock,
				block:         tc2Block,
				committees:    []incognitokey.CommitteePublicKey{},
			},
			want: false,
		},
		{
			name: "tc3: block re-propose but propose time <= produce time",
			fields: fields{
				logger:                 logger,
				chain:                  tc3Chain,
				nextBlockFinalityProof: tc3FinalityProof,
			},
			args: args{
				previousBlock: tc3PreviousBlock,
				block:         tc3Block,
				committees:    []incognitokey.CommitteePublicKey{},
			},
			want: false,
		},
		{
			name: "tc4: block re-propose but producer is the same as proposer",
			fields: fields{
				logger:                 logger,
				chain:                  tc4Chain,
				nextBlockFinalityProof: tc4FinalityProof,
			},
			args: args{
				previousBlock: tc4PreviousBlock,
				block:         tc4Block,
				committees:    tc4Committee,
			},
			want: false,
		},
		{
			name: "tc5: block re-propose but proposer is not compatible with re-propose timeslot",
			fields: fields{
				logger:                 logger,
				chain:                  tc5Chain,
				nextBlockFinalityProof: tc5FinalityProof,
			},
			args: args{
				previousBlock: tc5PreviousBlock,
				block:         tc5Block,
				committees:    tc5Committee,
			},
			want: false,
		},
		{
			name: "tc6: valid block re-propose",
			fields: fields{
				logger:                 logger,
				chain:                  tc6Chain,
				nextBlockFinalityProof: tc6FinalityProof,
			},
			args: args{
				previousBlock: tc6PreviousBlock,
				block:         tc6Block,
				committees:    tc6Committee,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProposeRuleLemma2{
				logger:                 tt.fields.logger,
				nextBlockFinalityProof: tt.fields.nextBlockFinalityProof,
				chain:                  tt.fields.chain,
			}
			if got := p.isReProposeFromFirstBlockNextHeight(tt.args.previousBlock, tt.args.block, tt.args.committees, config.Param().CommitteeSize.NumberOfFixedShardBlockValidator); got != tt.want {
				t.Errorf("isReProposeFromFirstBlockNextHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProposeRuleLemma2_verifyLemma2FirstBlockNextHeight(t *testing.T) {

	config.AbortParam()
	config.Param().CommitteeSize.NumberOfFixedShardBlockValidator = 8

	tc1ProposeMsg := &BFTPropose{
		ReProposeHashSignature: "1sfuYPVAjFPuG7uHReskCSqN8eBvN33usJ1RZCMfbLGD6rYh4WnfsdULBMP4Wp2kw3Hw5nFSEYBY8cQjoA2vJxX5PJjWbuF",
	}
	tc1Block := &mocksTypes.BlockInterface{}
	tc1BlockProposeTime := int64(1632144400)
	tc1BlockProduceTime := int64(1632144400)
	tc1Block.On("GetProposeTime").Return(tc1BlockProposeTime)
	tc1Block.On("GetProduceTime").Return(tc1BlockProduceTime)
	tc1Block.On("GetProposer").Return(shard0CommitteeString[0])
	tc1Block.On("GetProducer").Return(shard0CommitteeString[0])
	tc1Block.On("GetPrevHash").Return(common.Hash{}.NewHashFromStr2("aeeab909b457aefd1065fcde84daf150d88d135baf99fd054beaac7dce7df776"))
	tc1Block.On("GetAggregateRootHash").Return(common.Hash{}.NewHashFromStr2("474c87f4382c026f89ef9f74ddc3b0893d11a38f4783b36ee1a140ee3417fa09"))
	tc1FinalityProof := make(map[string]map[int64]string)
	tc1Chain := &mocks.Chain{}

	tc2ProposeMsg := &BFTPropose{
		ReProposeHashSignature: "12EGPejMDU3wwnAo2drdnWw14TmnFtozW1LBDtMA4oZR4FVfhxwRqMqkXoWMQymEy2tvMbcgkeVF9woiZV8dCHk1Fi8R8Wat",
	}
	tc2Block := &mocksTypes.BlockInterface{}
	tc2BlockProposeTime := int64(1632144300)
	tc2BlockProduceTime := int64(1632144300)
	tc2Block.On("GetFinalityHeight").Return(uint64(5))
	tc2Block.On("GetHeight").Return(uint64(6))
	tc2Block.On("GetProposeTime").Return(tc2BlockProposeTime)
	tc2Block.On("GetProduceTime").Return(tc2BlockProduceTime)
	tc2Block.On("GetProposer").Return(shard0CommitteeString[0])
	tc2Block.On("GetProducer").Return(shard0CommitteeString[0])
	tc2Block.On("GetPrevHash").Return(common.Hash{}.NewHashFromStr2("aeeab909b457aefd1065fcde84daf150d88d135baf99fd054beaac7dce7df776"))
	tc2Block.On("GetAggregateRootHash").Return(common.Hash{}.NewHashFromStr2("474c87f4382c026f89ef9f74ddc3b0893d11a38f4783b36ee1a140ee3417fa09"))
	tc2FinalityProof := make(map[string]map[int64]string)
	tc2Chain := &mocks.Chain{}

	tc3ProposeMsg := &BFTPropose{
		ReProposeHashSignature: "12EGPejMDU3wwnAo2drdnWw14TmnFtozW1LBDtMA4oZR4FVfhxwRqMqkXoWMQymEy2tvMbcgkeVF9woiZV8dCHk1Fi8R8Wat",
	}
	tc3Block := &mocksTypes.BlockInterface{}
	tc3BlockProposeTime := int64(1632144400)
	tc3BlockProduceTime := int64(1632144400)
	tc3Block.On("GetProposeTime").Return(tc3BlockProposeTime)
	tc3Block.On("GetProduceTime").Return(tc3BlockProduceTime)
	tc3Block.On("GetProposer").Return(shard0CommitteeString[1])
	tc3Block.On("GetProducer").Return(shard0CommitteeString[0])
	tc3Block.On("GetPrevHash").Return(common.Hash{}.NewHashFromStr2("aeeab909b457aefd1065fcde84daf150d88d135baf99fd054beaac7dce7df776"))
	tc3Block.On("GetAggregateRootHash").Return(common.Hash{}.NewHashFromStr2("474c87f4382c026f89ef9f74ddc3b0893d11a38f4783b36ee1a140ee3417fa09"))
	tc3FinalityProof := make(map[string]map[int64]string)
	tc3Chain := &mocks.Chain{}

	tc4ProposeMsg := &BFTPropose{
		ReProposeHashSignature: "12EGPejMDU3wwnAo2drdnWw14TmnFtozW1LBDtMA4oZR4FVfhxwRqMqkXoWMQymEy2tvMbcgkeVF9woiZV8dCHk1Fi8R8Wat",
	}
	tc4Block := &mocksTypes.BlockInterface{}
	tc4BlockProposeTime := int64(1632144400)
	tc4BlockProduceTime := int64(1632144400)
	tc4Block.On("GetFinalityHeight").Return(uint64(0))
	tc4Block.On("GetHeight").Return(uint64(6))
	tc4Block.On("GetProposeTime").Return(tc4BlockProposeTime)
	tc4Block.On("GetProduceTime").Return(tc4BlockProduceTime)
	tc4Block.On("GetProposer").Return(shard0CommitteeString[0])
	tc4Block.On("GetProducer").Return(shard0CommitteeString[0])
	tc4Block.On("GetPrevHash").Return(common.Hash{}.NewHashFromStr2("aeeab909b457aefd1065fcde84daf150d88d135baf99fd054beaac7dce7df776"))
	tc4Block.On("GetAggregateRootHash").Return(common.Hash{}.NewHashFromStr2("474c87f4382c026f89ef9f74ddc3b0893d11a38f4783b36ee1a140ee3417fa09"))
	tc4FinalityProof := make(map[string]map[int64]string)
	tc4Chain := &mocks.Chain{}

	tc5ProposeMsg := &BFTPropose{
		ReProposeHashSignature: "12EGPejMDU3wwnAo2drdnWw14TmnFtozW1LBDtMA4oZR4FVfhxwRqMqkXoWMQymEy2tvMbcgkeVF9woiZV8dCHk1Fi8R8Wat",
	}
	tc5Block := &mocksTypes.BlockInterface{}
	tc5BlockProposeTime := int64(1632144400)
	tc5BlockProduceTime := int64(1632144400)
	tc5Block.On("GetFinalityHeight").Return(uint64(5))
	tc5Block.On("GetHeight").Return(uint64(6))
	tc5Block.On("GetProposeTime").Return(tc5BlockProposeTime)
	tc5Block.On("GetProduceTime").Return(tc5BlockProduceTime)
	tc5Block.On("GetProposer").Return(shard0CommitteeString[0])
	tc5Block.On("GetProducer").Return(shard0CommitteeString[0])
	tc5Block.On("GetPrevHash").Return(common.Hash{}.NewHashFromStr2("aeeab909b457aefd1065fcde84daf150d88d135baf99fd054beaac7dce7df776"))
	tc5Block.On("GetAggregateRootHash").Return(common.Hash{}.NewHashFromStr2("474c87f4382c026f89ef9f74ddc3b0893d11a38f4783b36ee1a140ee3417fa09"))
	tc5FinalityProof := make(map[string]map[int64]string)
	tc5Chain := &mocks.Chain{}

	type fields struct {
		logger                 common.Logger
		nextBlockFinalityProof map[string]map[int64]string
		chain                  Chain
	}
	type args struct {
		proposeMsg *BFTPropose
		block      types.BlockInterface
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "tc1: invalid signature",
			fields: fields{
				logger:                 logger,
				chain:                  tc1Chain,
				nextBlockFinalityProof: tc1FinalityProof,
			},
			args: args{
				proposeMsg: tc1ProposeMsg,
				block:      tc1Block,
			},
			wantErr: true,
		},
		{
			name: "tc2: invalid signature",
			fields: fields{
				logger:                 logger,
				chain:                  tc2Chain,
				nextBlockFinalityProof: tc2FinalityProof,
			},
			args: args{
				proposeMsg: tc2ProposeMsg,
				block:      tc2Block,
			},
			wantErr: true,
		},
		{
			name: "tc3: invalid signature",
			fields: fields{
				logger:                 logger,
				chain:                  tc3Chain,
				nextBlockFinalityProof: tc3FinalityProof,
			},
			args: args{
				proposeMsg: tc3ProposeMsg,
				block:      tc3Block,
			},
			wantErr: true,
		},
		{
			name: "tc4: invalid finality height",
			fields: fields{
				logger:                 logger,
				chain:                  tc4Chain,
				nextBlockFinalityProof: tc4FinalityProof,
			},
			args: args{
				proposeMsg: tc4ProposeMsg,
				block:      tc4Block,
			},
			wantErr: true,
		},
		{
			name: "tc4: finality height",
			fields: fields{
				logger:                 logger,
				chain:                  tc5Chain,
				nextBlockFinalityProof: tc5FinalityProof,
			},
			args: args{
				proposeMsg: tc5ProposeMsg,
				block:      tc5Block,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProposeRuleLemma2{
				logger:                 tt.fields.logger,
				nextBlockFinalityProof: tt.fields.nextBlockFinalityProof,
				chain:                  tt.fields.chain,
			}
			if err := p.verifyLemma2FirstBlockNextHeight(tt.args.proposeMsg, tt.args.block); (err != nil) != tt.wantErr {
				t.Errorf("verifyLemma2FirstBlockNextHeight() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProposeRuleLemma2_verifyFinalityProof(t *testing.T) {

	config.AbortParam()
	config.Param().CommitteeSize.NumberOfFixedShardBlockValidator = 8

	tc1ProposeMsg := &BFTPropose{
		ReProposeHashSignature: "1sfuYPVAjFPuG7uHReskCSqN8eBvN33usJ1RZCMfbLGD6rYh4WnfsdULBMP4Wp2kw3Hw5nFSEYBY8cQjoA2vJxX5PJjWbuF",
		FinalityProof: FinalityProof{
			ReProposeHashSignature: []string{},
		},
	}
	tc1FinalityProof := make(map[string]map[int64]string)
	tc1Chain := &mocks.Chain{}
	tc1Block := &mocksTypes.BlockInterface{}
	tc1BlockProposeTime := int64(1632710360)
	tc1BlockProduceTime := int64(1632710080)
	tc1Block.On("GetProposeTime").Return(tc1BlockProposeTime)
	tc1Block.On("GetProduceTime").Return(tc1BlockProduceTime)
	tc1Block.On("GetProducer").Return(shard0CommitteeString[0])
	tc1Block.On("GetPrevHash").Return(common.Hash{}.NewHashFromStr2("1dc43704ebd3ffd6d35be478f86740fc0cbf5b2f825e66fe3c83f31e3e4d18d8"))
	tc1Block.On("GetAggregateRootHash").Return(common.Hash{}.NewHashFromStr2("874615a2fcacb79434fb851a137da03055eddd9bd086d7b60e6f74a565c887b9"))
	tc1Committee := shard0Committee

	tc2ProposeMsg := &BFTPropose{
		ReProposeHashSignature: "1sfuYPVAjFPuG7uHReskCSqN8eBvN33usJ1RZCMfbLGD6rYh4WnfsdULBMP4Wp2kw3Hw5nFSEYBY8cQjoA2vJxX5PJjWbuF",
		FinalityProof: FinalityProof{
			ReProposeHashSignature: []string{
				"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",
			},
		},
	}
	tc2FinalityProof := make(map[string]map[int64]string)
	tc2Chain := &mocks.Chain{}
	tc2Block := &mocksTypes.BlockInterface{}
	tc2BlockProposeTime := int64(1632710350)
	tc2BlockProduceTime := int64(1632710080)
	tc2Block.On("GetProposeTime").Return(tc2BlockProposeTime)
	tc2Block.On("GetProduceTime").Return(tc2BlockProduceTime)
	tc2Block.On("GetProducer").Return(shard0CommitteeString[0])
	tc2Block.On("GetPrevHash").Return(common.Hash{}.NewHashFromStr2("1dc43704ebd3ffd6d35be478f86740fc0cbf5b2f825e66fe3c83f31e3e4d18d8"))
	tc2Block.On("GetAggregateRootHash").Return(common.Hash{}.NewHashFromStr2("874615a2fcacb79434fb851a137da03055eddd9bd086d7b60e6f74a565c887b9"))
	tc2Committee := shard0Committee

	tc3ProposeMsg := &BFTPropose{
		ReProposeHashSignature: "1sfuYPVAjFPuG7uHReskCSqN8eBvN33usJ1RZCMfbLGD6rYh4WnfsdULBMP4Wp2kw3Hw5nFSEYBY8cQjoA2vJxX5PJjWbuF",
		FinalityProof: FinalityProof{
			ReProposeHashSignature: []string{
				"1rtHzbBpvTpeEWcGAqz5pTxh3x3tH1RsBPVbFXxWHXabeApbno2We1VwEHZoLJPoYuS5HQboTfQQoRY76VyttqosGRGAheL",
				"12YUeCqoWMgMhMewWRFLwExzCANqcJLkHA6yV3rLC3NoN3RqbcfMX9Mu69Qh6zHQtFWTGNdQrxmQt7BMfZXbapKwVQfyr15Q",
				"13RUN6EEaJErNn2VbL9awy23dnLj3TJ5A3zVeFeZQf3Tr2nvvQdp5fkBm2UntiLTBuiS3LdQpTMaPYcjJVzgcAb88Ey3Jbzc",
				"12AC2zst9JdVnjzcABEuWYPogBirvhEZ8N2yva4VuBhQjMopeQ4CbXn7YhGpyRSzmAhMQ2bxPz966PVDroMxhsrSBoAksGVb",
				"12bMtLH8ieVSyHq9JpcCVj3YLygyFfHG2oCvAXgkFTsFrFAvzU7829obY4GvuTn6FE5A4UmDv7guSZ9W8TgJchKJ3YsDDVYX",
				"17JfSHDcfshBLD31yaVJP6yA1pdiT9zAEFbnc5UVUJrZur9SeK9NsYvxQMR256VyZ1se2UYfkbREdqt4jx5dvyTwvQaeD72",
				"13SQ1JhH814JQpBqt7ADLSc1W2gqrzsYvAYijCP26Gj2cSk9KS9djQfqVKJK4w7j3Fs3QRrN698U5TcDRRu1rbm9mVF5FYoM",
				"1vwtMpKiJEe9ifQAQvU3fEUFksh8ZAE8hCurxiCnJzDJPZf44u3GQbAnax7VPoSTJuMZbbfz2XnS4yq5HxxxNA6jeJqiUA1",
				"1kY9Kk6bRDxKxeofyHKPg2mqTdsTSzVD8aME3v9duZ6ak8GS2f6YrivfpYKUwGh6BqCXqr1vae5pwuX3ruUQfZa3kETYE8R",
				"1GxKCwi2vpbBoo4J1fmkDQ9ckVPJhhibGawHhrg5q7FAHLDksBHNgaxSx3GVUTp9svhkhu5FxZyNRHnJbtqb94bgSGaP3ik",
				"13JkUXhVLw793siSZjmf5A5MxS7y18jndMQzVexeGae1cp49L5Z1Edfss3A5K4Ch4yHUJBcBgWLcJLswkKGVNpwjocVgHTdu",
				"13XoDJC4FnEmZANyzkCsH4iJc4XtMWe83ddmrzZj5Jy6dJZyF4yjsjoC2R2KWipk3UTuAiwt3MqHiyTjMruDrhnyxu4RxmvY",
				"1BQDpLnnuaoGNxtrALkcYDzjawAzZCL7273CYxNkMpYGZbrVwQgrttuqyotKXSVHPXXG69BbohQ1eo9oghBwZzL833q5d4D",
				"12VY9xEet96chLnA447dHatCGGxEnYxVamXHTX7pDGUtiGThheDwwEU131GgPmFyQwJrcSEAJXAWUpE1c7XVzJ4zJD5dMhB3",
				"12Wq21wvZwRVj1vxJyuwHzxjyc8ysuNhFUhSSRFHJF2zzY2GUyVwwwwmmudoank44fnLBLBWJYnq9SvsimW3wymBjdgiETwY",
				"1XoLmw8cnNboATkgomVn2sw9rxsSygsr4MhYJmuyvx8giR2tJGjfkigjPrmdtgiVoMCKjUt77jWBAfMSuHcQYSuezhnfvDM",
				"12mC9J82yvXoKzijsBRsSpoAHU75z4ubgkqhXU2EcRfnWFox81GRpVeneVCBARiEf34ZxMz2A8X1jMsjFKF5j53B8PHqPiHA",
				"12y8m9Wjv4FS7VGnHRHBkwmPXkjpU8JEaFnHL8H2DEdZafsUdRRwBU6pvKwwo8kJNRNT54h2iLTPvwpJb89V1T5M5Ey4PadE",
				"1mPHFkZZpHdKhDNYQFxAL52hcFi9Lc6NsxL5Ry69jyW4zV7LSPmVPuW72KjfY8UbEVEFw8LwhTKuoSFnmt5miZu7CJKvxWa",
				"13U1f34ZHKknuQECUsac2oJZDyYAGsRUxkykn69tpwk7c38GixnmvKzweRsrtrpcxm2Khyo7gSXvVZ3yKMJpKDPCNpQCyxaz",
				"12ttXxmGyuNBbvyq6rC2rd138ouqsz5Lm7pYoE9GbRq2QBSSWM9smMmUEZN8auvkzrjc5k6djXhiCnYNdC8TXbEeLGsaybWH",
				"13JmZvq9vRBKbPw1qCUQcTscmUQbmSaR4C6j7p5Wo7cmkZVE9SUq9fmR9iseZLYLRu7b8h5qfUfmJ7k3GcvoLStdgveLQbcX",
				"131BYuArgebxNyb969qkQbW7nuvcsQsscC27mjVQ28b3E2W2ubcxWnV3YNVaQHfqMnybKNWwarGmTG9wV23pitAZRDa2Jpmb",
				"1q4iDQs9Awy1EM1tZUHykNzo65sWjozz6gPgQMH77QZHsMUpTUy33PyEbnGdASHkKYiQ2MjtpjWxhhVdydW35fvJbwatpxd",
				"13YPuZReCuX5UfGLKr5owYbQ61PXyRBxfuhNN38rxNztbHSSXajUNe483R6YgKdSntYcV8uCkpmezfgssFeRLTC2r6dNGiKe",
				"1ko7uCH2C5YvqzMCVeV5kr5JnFvADoy2ikSwtV1v78KYNKssUDBjgZMHkuJuShHid5dz7Svw1xcQxZXS1vkSJeyziRk2yHv",
				"1QBEjzWtF1bsCCeqBh4Jn7PYz42fRtCjv5qgBR1d23DWh6TE82WGBUjyQ1vBR9yq3n6aPmi2LjC2UYJ65EofCKUwGu9a5jA",
			},
		},
	}
	tc3FinalityProof := make(map[string]map[int64]string)
	tc3Chain := &mocks.Chain{}
	tc3Block := &mocksTypes.BlockInterface{}
	tc3BlockProposeTime := int64(1632710350)
	tc3BlockProduceTime := int64(1632710080)
	tc3Block.On("GetProposeTime").Return(tc3BlockProposeTime)
	tc3Block.On("GetProduceTime").Return(tc3BlockProduceTime)
	tc3Block.On("GetProducer").Return(shard0CommitteeString[0])
	tc3Block.On("GetPrevHash").Return(common.Hash{}.NewHashFromStr2("1dc43704ebd3ffd6d35be478f86740fc0cbf5b2f825e66fe3c83f31e3e4d18d8"))
	tc3Block.On("GetAggregateRootHash").Return(common.Hash{}.NewHashFromStr2("874615a2fcacb79434fb851a137da03055eddd9bd086d7b60e6f74a565c887b9"))
	tc3Committee := shard0Committee

	type fields struct {
		logger                 common.Logger
		nextBlockFinalityProof map[string]map[int64]string
		chain                  Chain
	}
	type args struct {
		chainID    int
		proposeMsg *BFTPropose
		block      types.BlockInterface
		committees []incognitokey.CommitteePublicKey
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "tc1: not enough proof",
			fields: fields{
				logger:                 logger,
				chain:                  tc1Chain,
				nextBlockFinalityProof: tc1FinalityProof,
			},
			args: args{
				proposeMsg: tc1ProposeMsg,
				block:      tc1Block,
				committees: tc1Committee,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "tc2: invalid proof",
			fields: fields{
				logger:                 logger,
				chain:                  tc2Chain,
				nextBlockFinalityProof: tc2FinalityProof,
			},
			args: args{
				proposeMsg: tc2ProposeMsg,
				block:      tc2Block,
				committees: tc2Committee,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "tc3: valid proof",
			fields: fields{
				logger:                 logger,
				chain:                  tc3Chain,
				nextBlockFinalityProof: tc3FinalityProof,
			},
			args: args{
				proposeMsg: tc3ProposeMsg,
				block:      tc3Block,
				committees: tc3Committee,
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProposeRuleLemma2{
				logger:                 tt.fields.logger,
				nextBlockFinalityProof: tt.fields.nextBlockFinalityProof,
				chain:                  tt.fields.chain,
			}
			got, err := p.verifyFinalityProof(tt.args.proposeMsg, tt.args.block, tt.args.committees, config.Param().CommitteeSize.NumberOfFixedShardBlockValidator)
			if (err != nil) != tt.wantErr {
				t.Errorf("verifyFinalityProof() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("verifyFinalityProof() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProposeRuleLemma2_verifyLemma2ReProposeBlockNextHeight(t *testing.T) {

	config.AbortParam()
	config.Param().CommitteeSize.NumberOfFixedShardBlockValidator = 8

	tc1ProposeMsg := &BFTPropose{
		ReProposeHashSignature: "12Xzv92cdgguVCNzAFS7LANpXyXXdWAYDXFJssQgGK5236q8HPzW8RQKPmrEhVeP6A9fCvswfKihKWwKBURrzSaPZ7H5pnLm",
		FinalityProof: FinalityProof{
			ReProposeHashSignature: []string{},
		},
	}
	tc1FinalityProof := make(map[string]map[int64]string)
	tc1Chain := &mocks.Chain{}
	tc1Block := &mocksTypes.BlockInterface{}
	tc1BlockProposeTime := int64(1632753880)
	tc1BlockProduceTime := int64(1632753260)
	tc1Block.On("GetProposeTime").Return(tc1BlockProposeTime)
	tc1Block.On("GetProduceTime").Return(tc1BlockProduceTime)
	tc1Block.On("GetProducer").Return(shard0CommitteeString[6])
	tc1Block.On("GetProposer").Return(shard0CommitteeString[5])
	tc1Block.On("GetPrevHash").Return(common.Hash{}.NewHashFromStr2("2fc3ec987e4d0a56e344a16dbba524845f5aca6b699b0767735f0ad59926ea03"))
	tc1Block.On("GetAggregateRootHash").Return(common.Hash{}.NewHashFromStr2("698bed377b2d7c6c280924f6af6c46c7e84acddddc99268ac76e9393b3d3b05f"))
	tc1Committee := shard0Committee

	tc2ProposeMsg := &BFTPropose{
		ReProposeHashSignature: "12Xzv92cdgguVCNzAFS7LANpXyXXdWAYDXFJssQgGK5236q8HPzW8RQKPmrEhVeP6A9fCvswfKihKWwKBURrzSaPZ7H5pnLm",
		FinalityProof: FinalityProof{
			ReProposeHashSignature: []string{
				"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",
			},
		},
	}
	tc2FinalityProof := make(map[string]map[int64]string)
	tc2Chain := &mocks.Chain{}
	tc2Block := &mocksTypes.BlockInterface{}
	tc2BlockProposeTime := int64(1632753880)
	tc2BlockProduceTime := int64(1632753260)
	tc2Block.On("GetProposeTime").Return(tc2BlockProposeTime)
	tc2Block.On("GetProduceTime").Return(tc2BlockProduceTime)
	tc2Block.On("GetProducer").Return(shard0CommitteeString[6])
	tc2Block.On("GetProposer").Return(shard0CommitteeString[4])
	tc2Block.On("GetPrevHash").Return(common.Hash{}.NewHashFromStr2("2fc3ec987e4d0a56e344a16dbba524845f5aca6b699b0767735f0ad59926ea03"))
	tc2Block.On("GetAggregateRootHash").Return(common.Hash{}.NewHashFromStr2("698bed377b2d7c6c280924f6af6c46c7e84acddddc99268ac76e9393b3d3b05f"))
	tc2Committee := shard0Committee

	tc3ProposeMsg := &BFTPropose{
		ReProposeHashSignature: "12Xzv92cdgguVCNzAFS7LANpXyXXdWAYDXFJssQgGK5236q8HPzW8RQKPmrEhVeP6A9fCvswfKihKWwKBURrzSaPZ7H5pnLm",
		FinalityProof: FinalityProof{
			ReProposeHashSignature: []string{},
		},
	}
	tc3FinalityProof := make(map[string]map[int64]string)
	tc3Chain := &mocks.Chain{}
	tc3Block := &mocksTypes.BlockInterface{}
	tc3BlockProposeTime := int64(1632753880)
	tc3BlockProduceTime := int64(1632753260)
	tc3Block.On("GetProposeTime").Return(tc3BlockProposeTime)
	tc3Block.On("GetProduceTime").Return(tc3BlockProduceTime)
	tc3Block.On("GetProducer").Return(shard0CommitteeString[6])
	tc3Block.On("GetProposer").Return(shard0CommitteeString[4])
	tc3Block.On("GetPrevHash").Return(common.Hash{}.NewHashFromStr2("2fc3ec987e4d0a56e344a16dbba524845f5aca6b699b0767735f0ad59926ea03"))
	tc3Block.On("GetAggregateRootHash").Return(common.Hash{}.NewHashFromStr2("698bed377b2d7c6c280924f6af6c46c7e84acddddc99268ac76e9393b3d3b05f"))
	tc3Block.On("GetFinalityHeight").Return(uint64(4))
	tc3Block.On("GetHeight").Return(uint64(5))
	tc3Committee := shard0Committee

	tc4ProposeMsg := &BFTPropose{
		ReProposeHashSignature: "12Xzv92cdgguVCNzAFS7LANpXyXXdWAYDXFJssQgGK5236q8HPzW8RQKPmrEhVeP6A9fCvswfKihKWwKBURrzSaPZ7H5pnLm",
		FinalityProof: FinalityProof{
			ReProposeHashSignature: []string{
				"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",
			},
		},
	}
	tc4FinalityProof := make(map[string]map[int64]string)
	tc4Chain := &mocks.Chain{}
	tc4Block := &mocksTypes.BlockInterface{}
	tc4BlockProposeTime := int64(1632753880)
	tc4BlockProduceTime := int64(1632753260)
	tc4Block.On("GetProposeTime").Return(tc4BlockProposeTime)
	tc4Block.On("GetProduceTime").Return(tc4BlockProduceTime)
	tc4Block.On("GetProducer").Return(shard0CommitteeString[6])
	tc4Block.On("GetProposer").Return(shard0CommitteeString[4])
	tc4Block.On("GetPrevHash").Return(common.Hash{}.NewHashFromStr2("2fc3ec987e4d0a56e344a16dbba524845f5aca6b699b0767735f0ad59926ea03"))
	tc4Block.On("GetAggregateRootHash").Return(common.Hash{}.NewHashFromStr2("698bed377b2d7c6c280924f6af6c46c7e84acddddc99268ac76e9393b3d3b05f"))
	tc4Block.On("GetFinalityHeight").Return(uint64(0))
	tc4Block.On("GetHeight").Return(uint64(5))
	tc4Committee := shard0Committee

	tc5ProposeMsg := &BFTPropose{
		ReProposeHashSignature: "12Xzv92cdgguVCNzAFS7LANpXyXXdWAYDXFJssQgGK5236q8HPzW8RQKPmrEhVeP6A9fCvswfKihKWwKBURrzSaPZ7H5pnLm",
		FinalityProof: FinalityProof{
			ReProposeHashSignature: []string{
				"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",
			},
		},
	}
	tc5FinalityProof := make(map[string]map[int64]string)
	tc5Chain := &mocks.Chain{}
	tc5Block := &mocksTypes.BlockInterface{}
	tc5BlockProposeTime := int64(1632753880)
	tc5BlockProduceTime := int64(1632753260)
	tc5Block.On("GetProposeTime").Return(tc5BlockProposeTime)
	tc5Block.On("GetProduceTime").Return(tc5BlockProduceTime)
	tc5Block.On("GetProducer").Return(shard0CommitteeString[6])
	tc5Block.On("GetProposer").Return(shard0CommitteeString[4])
	tc5Block.On("GetPrevHash").Return(common.Hash{}.NewHashFromStr2("2fc3ec987e4d0a56e344a16dbba524845f5aca6b699b0767735f0ad59926ea03"))
	tc5Block.On("GetAggregateRootHash").Return(common.Hash{}.NewHashFromStr2("698bed377b2d7c6c280924f6af6c46c7e84acddddc99268ac76e9393b3d3b05f"))
	tc5Block.On("GetFinalityHeight").Return(uint64(4))
	tc5Block.On("GetHeight").Return(uint64(5))
	tc5Committee := shard0Committee

	type fields struct {
		logger                 common.Logger
		nextBlockFinalityProof map[string]map[int64]string
		chain                  Chain
	}
	type args struct {
		proposeMsg *BFTPropose
		block      types.BlockInterface
		committees []incognitokey.CommitteePublicKey
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "tc1: invalid repropose signature",
			fields: fields{
				logger:                 logger,
				chain:                  tc1Chain,
				nextBlockFinalityProof: tc1FinalityProof,
			},
			args: args{
				proposeMsg: tc1ProposeMsg,
				block:      tc1Block,
				committees: tc1Committee,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "tc2: invalid finality proof",
			fields: fields{
				logger:                 logger,
				chain:                  tc2Chain,
				nextBlockFinalityProof: tc2FinalityProof,
			},
			args: args{
				proposeMsg: tc2ProposeMsg,
				block:      tc2Block,
				committees: tc2Committee,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "tc3: invalid finality height, expect 0",
			fields: fields{
				logger:                 logger,
				chain:                  tc3Chain,
				nextBlockFinalityProof: tc3FinalityProof,
			},
			args: args{
				proposeMsg: tc3ProposeMsg,
				block:      tc3Block,
				committees: tc3Committee,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "tc4: invalid finality height, expect height - 1",
			fields: fields{
				logger:                 logger,
				chain:                  tc4Chain,
				nextBlockFinalityProof: tc4FinalityProof,
			},
			args: args{
				proposeMsg: tc4ProposeMsg,
				block:      tc4Block,
				committees: tc4Committee,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "tc5: valid re propose block",
			fields: fields{
				logger:                 logger,
				chain:                  tc5Chain,
				nextBlockFinalityProof: tc5FinalityProof,
			},
			args: args{
				proposeMsg: tc5ProposeMsg,
				block:      tc5Block,
				committees: tc5Committee,
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProposeRuleLemma2{
				logger:                 tt.fields.logger,
				nextBlockFinalityProof: tt.fields.nextBlockFinalityProof,
				chain:                  tt.fields.chain,
			}
			got, err := p.verifyLemma2ReProposeBlockNextHeight(tt.args.proposeMsg, tt.args.block, tt.args.committees, config.Param().CommitteeSize.NumberOfFixedShardBlockValidator)
			if (err != nil) != tt.wantErr {
				t.Errorf("verifyLemma2ReProposeBlockNextHeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("verifyLemma2ReProposeBlockNextHeight() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProposeRuleLemma2_addFinalityProof(t *testing.T) {

	config.AbortParam()
	config.Param().CommitteeSize.NumberOfFixedShardBlockValidator = 8

	tc1FinalityProof := make(map[string]map[int64]string)
	tc1Chain := &mocks.Chain{}
	tc1Block := &mocksTypes.BlockInterface{}
	tc1InputProof := FinalityProof{
		ReProposeHashSignature: []string{},
	}
	tc1ReProposeHashSignature := ""
	tc1BlockProposeTime := int64(1632753910)
	tc1BlockProduceTime := int64(1632753260)
	tc1Block.On("GetProposeTime").Return(tc1BlockProposeTime)
	tc1Block.On("GetProduceTime").Return(tc1BlockProduceTime)
	tc1Block.On("GetProducer").Return(shard0CommitteeString[6])
	tc1Block.On("GetProposer").Return(shard0CommitteeString[5])
	tc1Block.On("GetPrevHash").Return(common.Hash{}.NewHashFromStr2("2fc3ec987e4d0a56e344a16dbba524845f5aca6b699b0767735f0ad59926ea03"))
	tc1Block.On("GetAggregateRootHash").Return(common.Hash{}.NewHashFromStr2("698bed377b2d7c6c280924f6af6c46c7e84acddddc99268ac76e9393b3d3b05f"))

	tc2FinalityProof := make(map[string]map[int64]string)
	tc2Chain := &mocks.Chain{}
	tc2Block := &mocksTypes.BlockInterface{}
	tc2InputProof := FinalityProof{
		ReProposeHashSignature: []string{"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",
		},
	}
	tc2ReProposeHashSignature := "12Xzv92cdgguVCNzAFS7LANpXyXXdWAYDXFJssQgGK5236q8HPzW8RQKPmrEhVeP6A9fCvswfKihKWwKBURrzSaPZ7H5pnLm"
	tc2BlockProposeTime := int64(1632753880)
	tc2BlockProduceTime := int64(1632753260)
	tc2Block.On("GetProposeTime").Return(tc2BlockProposeTime)
	tc2Block.On("GetProduceTime").Return(tc2BlockProduceTime)
	tc2Block.On("GetProducer").Return(shard0CommitteeString[6])
	tc2Block.On("GetProposer").Return(shard0CommitteeString[5])
	tc2Block.On("GetHeight").Return(uint64(972)).Times(64)
	tc2Hash := common.Hash{}.NewHashFromStr2("1b7ad1402f18233a1e7cdf81a7e817c83cce18f210bc10a8523155e9d5f27db4")
	tc2Block.On("Hash").Return(&tc2Hash).Times(64)
	tc2Block.On("GetPrevHash").Return(common.Hash{}.NewHashFromStr2("2fc3ec987e4d0a56e344a16dbba524845f5aca6b699b0767735f0ad59926ea03"))
	tc2Block.On("GetAggregateRootHash").Return(common.Hash{}.NewHashFromStr2("698bed377b2d7c6c280924f6af6c46c7e84acddddc99268ac76e9393b3d3b05f"))

	tc3FinalityProof := make(map[string]map[int64]string)
	tc3Chain := &mocks.Chain{}
	tc3Block := &mocksTypes.BlockInterface{}
	tc3InputProof := FinalityProof{
		ReProposeHashSignature: []string{
			"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",
		},
	}
	tc3ReProposeHashSignature := "12Xzv92cdgguVCNzAFS7LANpXyXXdWAYDXFJssQgGK5236q8HPzW8RQKPmrEhVeP6A9fCvswfKihKWwKBURrzSaPZ7H5pnLm"
	tc3BlockProposeTime := int64(1632753880)
	tc3BlockProduceTime := int64(1632753260)
	tc3Block.On("GetProposeTime").Return(tc3BlockProposeTime)
	tc3Block.On("GetProduceTime").Return(tc3BlockProduceTime)
	tc3Block.On("GetProducer").Return(shard0CommitteeString[6])
	tc3Block.On("GetProposer").Return(shard0CommitteeString[4])
	tc3Block.On("GetHeight").Return(uint64(972)).Times(64)
	tc3Hash := common.Hash{}.NewHashFromStr2("1b7ad1402f18233a1e7cdf81a7e817c83cce18f210bc10a8523155e9d5f27db4")
	tc3Block.On("Hash").Return(&tc3Hash).Times(64)
	tc3Block.On("GetPrevHash").Return(common.Hash{}.NewHashFromStr2("2fc3ec987e4d0a56e344a16dbba524845f5aca6b699b0767735f0ad59926ea03")).Times(10)
	tc3Block.On("GetAggregateRootHash").Return(common.Hash{}.NewHashFromStr2("698bed377b2d7c6c280924f6af6c46c7e84acddddc99268ac76e9393b3d3b05f"))

	type fields struct {
		logger                 common.Logger
		nextBlockFinalityProof map[string]map[int64]string
		chain                  Chain
	}
	type args struct {
		block                  types.BlockInterface
		reProposeHashSignature string
		proof                  FinalityProof
	}
	tests := []struct {
		name                       string
		fields                     fields
		args                       args
		wantErr                    bool
		wantNextBlockFinalityProof map[int64]string
	}{
		{
			name: "tc1: more than 64 finality proof required",
			fields: fields{
				logger:                 logger,
				nextBlockFinalityProof: tc1FinalityProof,
				chain:                  tc1Chain,
			},
			args: args{
				block:                  tc1Block,
				reProposeHashSignature: tc1ReProposeHashSignature,
				proof:                  tc1InputProof,
			},
			wantErr:                    false,
			wantNextBlockFinalityProof: nil,
		},
		{
			name: "tc2: not enough proof after validation",
			fields: fields{
				logger:                 logger,
				nextBlockFinalityProof: tc2FinalityProof,
				chain:                  tc2Chain,
			},
			args: args{
				block:                  tc2Block,
				reProposeHashSignature: tc2ReProposeHashSignature,
				proof:                  tc2InputProof,
			},
			wantErr:                    true,
			wantNextBlockFinalityProof: nil,
		},
		{
			name: "tc3: add proof success",
			fields: fields{
				logger:                 logger,
				nextBlockFinalityProof: tc3FinalityProof,
				chain:                  tc3Chain,
			},
			args: args{
				block:                  tc3Block,
				reProposeHashSignature: tc3ReProposeHashSignature,
				proof:                  tc3InputProof,
			},
			wantErr: false,
			wantNextBlockFinalityProof: map[int64]string{
				163275326: "1ec352qhKtECAtECXP5pmdAtJyfCPSjjVCnL4TXZ3DQuTkSuvmT2yNBFtc4m9phexFXZqHzaoesmmC6SHYwEVDFeysuXfRH",
				163275327: "1htdgUEgxdxTYSXNUP5pPEipknKwhiD8wwngDGXZE7qRWy43VhjvXzDRHNuyn1L4SMvxwpBMvqu1UfCkQkWV3ZpdUQAQYoX",
				163275328: "1E2x3nf8K3YLdoPDQeWh8mtiS5zhErfTYvBiniidDZXYBVfET8zWkswMepQ4hd39Q7zTLZ4UTbmx9oP4QsfBdDTLVKcgQrt",
				163275329: "12F11pox3YdzBUG3816MVeFEAxmJwUBKnfBbG1NqwAwv8Kyh4ncw9Cs5Wuazc6ufVNyjec25UWuzQzs96KKE6ZnifCyUmJPD",
				163275330: "13BEZpSwNeTfCv5WYixgfB4FxWzdub9cmREauKPmNTwRbpd5TbSt7J7na9Pbr5Rw9Ekyx4gBxtSCLintpjv89P1XfRCkDvPm",
				163275331: "12LU9ucBBdYUcATfWG4LTrcb7TBhLfCWmeNd6zKmzQGZv7r5UuwiuWHf5hyZBXjBhvy6zXQGwdKnabH3Bouky3hstGbc97Wd",
				163275332: "13WQzqFyLBCHirozUBsrhuxYaTz52H9QJktit7XxZujG5Nts7TV73vLN99ornnWzB25T6x7fDUUgpRzYokJfG43cENYsMf2U",
				163275333: "1i1vHYuXCAFeeHwFjpRVaZ3hJoosibNJmWbnJGnWjTPgwJJX9S13t9fnTz5dunTEuDfBA8hEhXQE2PKpmaJmZmDP7wpW6F1",
				163275334: "12facYZQNjAeWhN38Z1on9t9SH99ieTTt4j2keKcUGjoWfLCfUbBxooZMgMNjJnvoE4QQtyKhzgu8c2j79swyVYf5esHPHCS",
				163275335: "12CS3AnRoTZaMeM9FP2sFhRhmxLwo5mBGkzrZtM47b5ujVDMtWi1ciea9ggSAc7b2j6XJZhxXuxa8zPJsqgJe8sExJwTgaXJ",
				163275336: "1wiQyj744VuCCEob86P5bNwNngEmM5EvCrNqyxLb4mzgi6Dpjiqbuu7Xm7XABxT5FuqME5fJP3BWEdwerj3HnEhDcWDu2Hc",
				163275337: "16q2LnGWbpR2HqN9BAewiqtPd44XMkCj4bRu4j16tWb9L2hch5gYHznFfQstVv6Rj9PPi42mAGVZQZR6LDGfVSyZv6jPN7Q",
				163275338: "1JM87NAVP9cSatv5sxviJv4HZmrb2HkyDg3MvijrwHA8FdrQ8NAWLnVhVDZVpa5xwssZBWbQSrnUFeEyRJEacnyn7T4oyTg",
				163275339: "13R1QBhFyP2FVy4wnYbmfdAWNS8abJUf4tyFWL2sRNAc4QKKindDdvmCChdSQKYV3WMbZcoZfaPv3X278ZYUvh7gpJj4BWy6",
				163275340: "124h5SDb54eexCyr1Tqvi7peaJ9K4V8stfBdXr5LimSHsvyQoFveMoSB8zo527ipAWDqYjFB6cqEemBNT1unwjSZ5GcxaLPi",
				163275341: "12mVkWiqbDWBnh4mAeEon1Y4aN7XT98LKLKy96SNKUwQByroN69NCsBYbXDX2uVe75n6pooP4NuXSEtYFXBqGvWiaHko1npe",
				163275342: "1D7SRv14gf2xKmGBdCUPYjfdFRtHQ3TWymVixdUu5yvWoUP77gd6yekW4SJsKmDoD8gbm7ji6UtKj7gYayM5sUnGvJdjpXT",
				163275343: "13KJDV4BrJCitoqRj5KcxxEUHH3MXu7Mv4mQoHKyS7xhEWXX25nBo5SvFeMM27qXyShJJLuGE3hQLDd8bqBEsfJushTbgPXB",
				163275344: "1oNmKErZ9c7vgVXWvbLznKiFo7cBToLLU6vZ6RgJtEuPUpBkmDuShW8y3JNJtiEX48SjENw1wbsUEp3L2hEZEaD4YUXU95v",
				163275345: "12wDwm5pC6BHYHzLvuYGsskAhkYa3LvwiRPxg5ALnxCefod6e3XxvC8WEWMeSay71N4cyos7xNZGQ5KpWLTAtJ1wK4KdwZik",
				163275346: "12jz7iiYKVjYTxVduJeyi31jMPR6eqRmSC9G7SAYVYYnJoU5tSybKbwx7acJhMzFVafESm6P8esMNFTBmEEDgUjwMQvR14Gr",
				163275347: "1XxiMocds1N9qfq2xpfNaH9U7Sikqq5k12XLrHMWqpZp6q3bJCdt5kVH9MK93rKM6epQAUxuJHxYBm2B2PyP4HFXfCVZH2D",
				163275348: "12e5z6n492TSugBUd1Uy4ed6sdfAU7v8LGBJvEQiyCNJKWWvZWYG8b1Cd5Gs9VdudASs5fb9Zwm7rn5wLYr43A9VEYCjk7Au",
				163275349: "12xo97b2sSCyn8CPq3Ep1G17rz8GDpR427DxGkuBNkFGakjJwB2Dm8QrUPzLTM36sYzD58UNThQw9cZbuxnq7JkuqmKggi89",
				163275350: "164JGkqKjvNhmBFWvNCAVhz67yzakw4sRC3m7k544m3DLKP5F9eS9WVSyDxtmbrF83xiGCxLCP6vWih5vuuxKznxyPYeabG",
				163275351: "1vcE2eFWbgr6hijPoxcZwLwmtZaZiNyZkoE7jBZbtNf3wu7FUK4hXyfgPz3aQYvQdcDVF4QLAUcjaAisjUrYdPac83ReYeX",
				163275352: "13Uhc4nuu76t312ksD2dz7YNeZzp5HijK8QX8SzAPeUajjA1Ftvthz3ugQStCWDcy96PFA4iCWqouzpsNEK8py3S1cLcrbfC",
				163275353: "1E6oEPBormzpuAxXLo9sQhcVGgdpMEhkRKWcHViPj7NPRrJ4shND3XrbifJJRTZgiofnQF8fEq7dLR3LGfoHqqGBDkH39F6",
				163275354: "12iiJyC5GwNHoQcXazrGK32gWsVkjEzucpWdBuKMgUKFD3RFoQNYm8taE8MfSvmSQbVaNe2DuQcDKDAvaA67bdNUKe7pJ2bF",
				163275355: "12DEYEKQFZqAP3RbuCfFgSE2Yh58hyBdp47FwJ7LErkAK34MozUQzR8D41CfEryiVei1Gvf8bqAZcBasFLzJQR1K1j4wszZP",
				163275356: "1vwqK1hoWom2GyZv9QsasiFxwPiwqTHQdZ5SefJJBGMZYifr5thwaQtsC8DzgnvDxAdcFWxgZoJYDXh9omivqTWF7YYknmK",
				163275357: "1C9FwwxaXzpXLHoJcxeyWtMoGfGJQJ5hKQxLgZPNLhRNd7LPXnhuEuY7qxEQN89iAMuYVXUPf9REb1b3LcnTuZRgEpNwiZD",
				163275358: "13Y24AgSgBYi9sjtnNtDooLFi2kurhX6punvEKj75qRrtMDNrqxSMTMtDev69AHJ5sc9Ybb9Q5VK1gfpoishAVsUhfiJ6AmP",
				163275359: "1MkaKXhUq97yV73V7of1yoarVVAgfhwH1yTuojTwGwQTQQd2k6RxjQC6LDyhQvnTMHe3zEpJEoDFHvyticU7Tf5m4ayTU2N",
				163275360: "1rHFjq1UfG1eYmxGnv1cQ7ipChqX4BcmskXyPmCaGMfs5p6ncYRMKLcrpg1RV8ZjSGbffy8T63QRoBKdDHWjDVzrnd7yL6w",
				163275361: "16TSMUSBGi1mvYouvALMVUr6t94XtyBnjbPP5Xkkro15NDqYu2U6jHhWGyBNnpNQbebgr2FSRXr11Eq9KSD935aiSDQbeqZ",
				163275362: "12rFzDv9Z7vgsiReJqMgT9x9NTx6jd93RkGRcRymuD6otTHEFSB6RqaqBtHFJizY4tfULcrb1ZGoK3Sj9nmsKEvMc7QG8dHL",
				163275363: "12YswDF5qYuULX6BTiHM4gPdG2MiWCx1TzQp7h2oBpBw2RvCKPeuAt3PxXwfitdiUqEzvvAGeZYE3cMD3Uk9ZMAhxKoSugj7",
				163275364: "12QbaDwdDP7Y1PeP5nYJDjSY18x5WSpPjw4b9YRchAkJsjzbUfs2iCstLbZLH5vWGvvT9vCgQoWdMGLGpmYVQmso4HteniPa",
				163275365: "12fV3fabkvmEo49hXoygkxhteV46s1cjEpUuoMr6fwJbPJH3nsm4ofNBHoZj8zw7FAdy27HjTDywMFcV5yaSsMzC3zpQ9yoZ",
				163275366: "12iqbefziuhAzwsyRf7HiVLBzqgbcW5EVoqWQyj6iqKy5BVabsaP2aZn1rshq3vCZ4HRbyzpfxvauSzbZ1gmwKzBGobQLCG",
				163275367: "12fqNBMxMrTvTXLT5ZvG6PaUFbRFXxs82ES4GdNkCvntbK4dthcNXD1NL7WKAtgsoaWPm8azZgCN3XHkGQ1TCCGnT5QpAT44",
				163275368: "1hS8QZGbroa3X4AFc6rpLs5VPDktoqCLRrXv9Fqb1PFckfASYJ6eBSWE3iumP5pic52EJyjudBMpdkXnvtsFCnnCQ961KeL",
				163275369: "1346L8oxZGiSYAoFi3mm8GmQUcPhFsDPBDXrhDfteJssPwnkB81hoEXKwwMuPFA9FQbyMMMYF9zfhHxgP9aLqH3LvE1q2KRG",
				163275370: "12CyN7feFy9mnoC9QHHwFUgZEgAmghQG5nVNJzvWqP8qGrMYjFgP7hfzypUPt5PzxJs5Aqukobhq6GPXxoV17bUtbzMjGbPo",
				163275371: "12v6KwqVj4j3CFqqux1pk8Dse8BoroCEonbCYw75hmjkt5NppLpDDUKEmPHzvzmC3sFL4oPiEMF7N3z14W9BUaBaxqT54ej5",
				163275372: "1XPwruDEA64Fv9tMgvhhev7G9GmLw8mFjQDSdZwffKBRLen3GXBWqMmSRKNh4gTuLaZKRq4DQoVpBUiJgSP1v6a4bJv8Dg7",
				163275373: "12ucN4kCayCUuCtD5AGjRudxz37dGEt3ENWQXc5RZyerpZ2Hz3ZrqzC3YXUzCBMTBQLvYhWS5GkN6LUBPWUPr5EkxJVhy9kX",
				163275374: "12XoH5ApVf38mkbQRwgXW3VenfW4KXawir2mjjbt7KUdDiUCckKMKRj2wRARb1xPRAXUJrvmVWG8RPY2WwMNVbS9B5fHHGry",
				163275375: "12wNykBY8Evx1TypQbve6AZtaFk8EHuJACQM65vjfJAVEaQcL68nBWeRnow7MyCrfJ3Kx2GjYEvLKTq8hCbVX5RyZcovQRrd",
				163275376: "13E4xaSjPyjNPNUoqUKKdobdFN41UPRXmokVe2ahJq7T9KzPws7Q1au9LbJzX2Z1fwgUqQyuyk5oZfaiy4MkgMNh9hLdkjdQ",
				163275377: "1rhsuhhwc1Z8Tq9ZsPsV3X1SqyouDAhxJ2Mk1RYArab1mvLuXjqzcBGLMGkjkKRxPbehf9tynRx2BMPRhNfB5CHn4ZNfya8",
				163275378: "12pR2MmoBES1Np7eo25SRGQNMNFRbYbsFPAkRvyg6STQPXUcBDwyUdkGYEaV3ZDBJstaTwNABubwiuu7JpwCAvow9QnPUk2U",
				163275379: "1TEVmRy1NziqGLVpZ4f8vSJpVfkb8GFfNBEUz7weoYqaXYzWiUEPV8sRJgjRdpUgzBLC247NZ1N1KufbTyG4DBPQSNLrUbT",
				163275380: "14wuaDiVXEHbJT4z37TQhqHMCqMBnJCGRqYapx3AiJfVEXydCDdzet1Yi98NewrH6tAw5zx4VS75uWYQrk7JjbAYGMr6tL6",
				163275381: "12PcAy9oLEGzpvcDXVdTGXgdgTt95KuAbkopTZywT6zm8P74wdapKiYHWLn2KRNkhcMqLrAWFVsBLBFMS5dir2yZYpvKFeMv",
				163275382: "12mmRLASDda18D6RfGWRNNZHUMo9RhDeCvsioiwNSaQLG5qpTgULRSuMHmYPBpWkMERzPBFvi8qQWLzTBVBvAVLJKrbAMFha",
				163275383: "1j6NK4NkAGSk7ybENSrNdUX6pLe6MBv28tN4STuFhQqbLEzLFFE6NqpvscRjBhSxkeJnzGzTS3hUkRBqhmGGCguLJR4KxDY",
				163275384: "1d1DXDRvWof7V45P83k4hjCDk3VcmZeSeP2m2MzPjRrCNBGRkzoGeR6aBdT96i4ffrWx5CZXG2kf8bcGY5yx6P8Q8nbsDFD",
				163275385: "128HMLqu6MEvXmYujtBoC959dtLoTKmRrHhmYD5UqAcEu1PGoHX3dSh6f3GDf7rNuxYjSbW5vj44fbspCtyHU8ZfqoAa3xjd",
				163275386: "13K3J9z7eEjs2QfbuBgbfQgu1ejrajVg4bjJsgkUFyWNua5XqWW2otxYNE9duXBe4KJdmPv5KXnBBRDRM4CSh7bYYamCir6E",
				163275387: "13Gek6VGmapmFcLHDo8nMM49gC62sgNSY2PfzqhhUynmA29XGbeGVrZgaYURQiGskdUaJje71TvR1AGyGAzHENY2HAj5Ftey",
				163275388: "12Xzv92cdgguVCNzAFS7LANpXyXXdWAYDXFJssQgGK5236q8HPzW8RQKPmrEhVeP6A9fCvswfKihKWwKBURrzSaPZ7H5pnLm",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProposeRuleLemma2{
				logger:                 tt.fields.logger,
				nextBlockFinalityProof: tt.fields.nextBlockFinalityProof,
				chain:                  tt.fields.chain,
			}
			if err := p.addFinalityProof(tt.args.block, tt.args.reProposeHashSignature, tt.args.proof); (err != nil) != tt.wantErr {
				t.Errorf("addFinalityProof() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				if !reflect.DeepEqual(p.nextBlockFinalityProof[tt.args.block.GetPrevHash().String()], tt.wantNextBlockFinalityProof) {
					t.Errorf("addFinalityProof() nextBlockFinalityProof want %+v, got %+v", tt.wantNextBlockFinalityProof, p.nextBlockFinalityProof[tt.args.block.GetPrevHash().String()])
				}
			}
		})
	}
}

func TestProposeRuleLemma2_GetValidFinalityProof(t *testing.T) {

	config.AbortParam()
	config.Param().CommitteeSize.NumberOfFixedShardBlockValidator = 8

	tc1FinalityProof := make(map[string]map[int64]string)
	tc1Chain := &mocks.Chain{}
	tc1Block := &mocksTypes.BlockInterface{}

	tc1Block.On("GetProposeTime").Return(int64(1632753910))
	tc1Block.On("GetProduceTime").Return(int64(1632753260))
	tc1Block.On("GetProducer").Return(shard0CommitteeString[6])
	tc1Block.On("GetProposer").Return(shard0CommitteeString[5])
	tc1Block.On("GetPrevHash").Return(common.Hash{}.NewHashFromStr2("2fc3ec987e4d0a56e344a16dbba524845f5aca6b699b0767735f0ad59926ea03"))
	tc1Block.On("GetAggregateRootHash").Return(common.Hash{}.NewHashFromStr2("698bed377b2d7c6c280924f6af6c46c7e84acddddc99268ac76e9393b3d3b05f"))

	tc2FinalityProof := map[string]map[int64]string{
		"2fc3ec987e4d0a56e344a16dbba524845f5aca6b699b0767735f0ad59926ea03": {
			163275326: "1ec352qhKtECAtECXP5pmdAtJyfCPSjjVCnL4TXZ3DQuTkSuvmT2yNBFtc4m9phexFXZqHzaoesmmC6SHYwEVDFeysuXfRH",
			163275327: "1htdgUEgxdxTYSXNUP5pPEipknKwhiD8wwngDGXZE7qRWy43VhjvXzDRHNuyn1L4SMvxwpBMvqu1UfCkQkWV3ZpdUQAQYoX",
			163275328: "1E2x3nf8K3YLdoPDQeWh8mtiS5zhErfTYvBiniidDZXYBVfET8zWkswMepQ4hd39Q7zTLZ4UTbmx9oP4QsfBdDTLVKcgQrt",
			163275329: "12F11pox3YdzBUG3816MVeFEAxmJwUBKnfBbG1NqwAwv8Kyh4ncw9Cs5Wuazc6ufVNyjec25UWuzQzs96KKE6ZnifCyUmJPD",
			163275330: "13BEZpSwNeTfCv5WYixgfB4FxWzdub9cmREauKPmNTwRbpd5TbSt7J7na9Pbr5Rw9Ekyx4gBxtSCLintpjv89P1XfRCkDvPm",
			//163275331: "12LU9ucBBdYUcATfWG4LTrcb7TBhLfCWmeNd6zKmzQGZv7r5UuwiuWHf5hyZBXjBhvy6zXQGwdKnabH3Bouky3hstGbc97Wd",
			163275332: "13WQzqFyLBCHirozUBsrhuxYaTz52H9QJktit7XxZujG5Nts7TV73vLN99ornnWzB25T6x7fDUUgpRzYokJfG43cENYsMf2U",
			163275333: "1i1vHYuXCAFeeHwFjpRVaZ3hJoosibNJmWbnJGnWjTPgwJJX9S13t9fnTz5dunTEuDfBA8hEhXQE2PKpmaJmZmDP7wpW6F1",
			163275334: "12facYZQNjAeWhN38Z1on9t9SH99ieTTt4j2keKcUGjoWfLCfUbBxooZMgMNjJnvoE4QQtyKhzgu8c2j79swyVYf5esHPHCS",
			//163275335: "12CS3AnRoTZaMeM9FP2sFhRhmxLwo5mBGkzrZtM47b5ujVDMtWi1ciea9ggSAc7b2j6XJZhxXuxa8zPJsqgJe8sExJwTgaXJ",
			163275336: "1wiQyj744VuCCEob86P5bNwNngEmM5EvCrNqyxLb4mzgi6Dpjiqbuu7Xm7XABxT5FuqME5fJP3BWEdwerj3HnEhDcWDu2Hc",
			163275337: "16q2LnGWbpR2HqN9BAewiqtPd44XMkCj4bRu4j16tWb9L2hch5gYHznFfQstVv6Rj9PPi42mAGVZQZR6LDGfVSyZv6jPN7Q",
			163275338: "1JM87NAVP9cSatv5sxviJv4HZmrb2HkyDg3MvijrwHA8FdrQ8NAWLnVhVDZVpa5xwssZBWbQSrnUFeEyRJEacnyn7T4oyTg",
			//163275339: "13R1QBhFyP2FVy4wnYbmfdAWNS8abJUf4tyFWL2sRNAc4QKKindDdvmCChdSQKYV3WMbZcoZfaPv3X278ZYUvh7gpJj4BWy6",
			163275340: "124h5SDb54eexCyr1Tqvi7peaJ9K4V8stfBdXr5LimSHsvyQoFveMoSB8zo527ipAWDqYjFB6cqEemBNT1unwjSZ5GcxaLPi",
			163275341: "12mVkWiqbDWBnh4mAeEon1Y4aN7XT98LKLKy96SNKUwQByroN69NCsBYbXDX2uVe75n6pooP4NuXSEtYFXBqGvWiaHko1npe",
			163275342: "1D7SRv14gf2xKmGBdCUPYjfdFRtHQ3TWymVixdUu5yvWoUP77gd6yekW4SJsKmDoD8gbm7ji6UtKj7gYayM5sUnGvJdjpXT",
			163275343: "13KJDV4BrJCitoqRj5KcxxEUHH3MXu7Mv4mQoHKyS7xhEWXX25nBo5SvFeMM27qXyShJJLuGE3hQLDd8bqBEsfJushTbgPXB",
			163275344: "1oNmKErZ9c7vgVXWvbLznKiFo7cBToLLU6vZ6RgJtEuPUpBkmDuShW8y3JNJtiEX48SjENw1wbsUEp3L2hEZEaD4YUXU95v",
			163275345: "12wDwm5pC6BHYHzLvuYGsskAhkYa3LvwiRPxg5ALnxCefod6e3XxvC8WEWMeSay71N4cyos7xNZGQ5KpWLTAtJ1wK4KdwZik",
			163275346: "12jz7iiYKVjYTxVduJeyi31jMPR6eqRmSC9G7SAYVYYnJoU5tSybKbwx7acJhMzFVafESm6P8esMNFTBmEEDgUjwMQvR14Gr",
			163275347: "1XxiMocds1N9qfq2xpfNaH9U7Sikqq5k12XLrHMWqpZp6q3bJCdt5kVH9MK93rKM6epQAUxuJHxYBm2B2PyP4HFXfCVZH2D",
			163275348: "12e5z6n492TSugBUd1Uy4ed6sdfAU7v8LGBJvEQiyCNJKWWvZWYG8b1Cd5Gs9VdudASs5fb9Zwm7rn5wLYr43A9VEYCjk7Au",
			163275349: "12xo97b2sSCyn8CPq3Ep1G17rz8GDpR427DxGkuBNkFGakjJwB2Dm8QrUPzLTM36sYzD58UNThQw9cZbuxnq7JkuqmKggi89",
			163275350: "164JGkqKjvNhmBFWvNCAVhz67yzakw4sRC3m7k544m3DLKP5F9eS9WVSyDxtmbrF83xiGCxLCP6vWih5vuuxKznxyPYeabG",
			163275351: "1vcE2eFWbgr6hijPoxcZwLwmtZaZiNyZkoE7jBZbtNf3wu7FUK4hXyfgPz3aQYvQdcDVF4QLAUcjaAisjUrYdPac83ReYeX",
			163275352: "13Uhc4nuu76t312ksD2dz7YNeZzp5HijK8QX8SzAPeUajjA1Ftvthz3ugQStCWDcy96PFA4iCWqouzpsNEK8py3S1cLcrbfC",
			163275353: "1E6oEPBormzpuAxXLo9sQhcVGgdpMEhkRKWcHViPj7NPRrJ4shND3XrbifJJRTZgiofnQF8fEq7dLR3LGfoHqqGBDkH39F6",
			163275354: "12iiJyC5GwNHoQcXazrGK32gWsVkjEzucpWdBuKMgUKFD3RFoQNYm8taE8MfSvmSQbVaNe2DuQcDKDAvaA67bdNUKe7pJ2bF",
			163275355: "12DEYEKQFZqAP3RbuCfFgSE2Yh58hyBdp47FwJ7LErkAK34MozUQzR8D41CfEryiVei1Gvf8bqAZcBasFLzJQR1K1j4wszZP",
			163275356: "1vwqK1hoWom2GyZv9QsasiFxwPiwqTHQdZ5SefJJBGMZYifr5thwaQtsC8DzgnvDxAdcFWxgZoJYDXh9omivqTWF7YYknmK",
			163275357: "1C9FwwxaXzpXLHoJcxeyWtMoGfGJQJ5hKQxLgZPNLhRNd7LPXnhuEuY7qxEQN89iAMuYVXUPf9REb1b3LcnTuZRgEpNwiZD",
			163275358: "13Y24AgSgBYi9sjtnNtDooLFi2kurhX6punvEKj75qRrtMDNrqxSMTMtDev69AHJ5sc9Ybb9Q5VK1gfpoishAVsUhfiJ6AmP",
			163275359: "1MkaKXhUq97yV73V7of1yoarVVAgfhwH1yTuojTwGwQTQQd2k6RxjQC6LDyhQvnTMHe3zEpJEoDFHvyticU7Tf5m4ayTU2N",
			163275360: "1rHFjq1UfG1eYmxGnv1cQ7ipChqX4BcmskXyPmCaGMfs5p6ncYRMKLcrpg1RV8ZjSGbffy8T63QRoBKdDHWjDVzrnd7yL6w",
			163275361: "16TSMUSBGi1mvYouvALMVUr6t94XtyBnjbPP5Xkkro15NDqYu2U6jHhWGyBNnpNQbebgr2FSRXr11Eq9KSD935aiSDQbeqZ",
			163275362: "12rFzDv9Z7vgsiReJqMgT9x9NTx6jd93RkGRcRymuD6otTHEFSB6RqaqBtHFJizY4tfULcrb1ZGoK3Sj9nmsKEvMc7QG8dHL",
			163275363: "12YswDF5qYuULX6BTiHM4gPdG2MiWCx1TzQp7h2oBpBw2RvCKPeuAt3PxXwfitdiUqEzvvAGeZYE3cMD3Uk9ZMAhxKoSugj7",
			163275364: "12QbaDwdDP7Y1PeP5nYJDjSY18x5WSpPjw4b9YRchAkJsjzbUfs2iCstLbZLH5vWGvvT9vCgQoWdMGLGpmYVQmso4HteniPa",
			163275365: "12fV3fabkvmEo49hXoygkxhteV46s1cjEpUuoMr6fwJbPJH3nsm4ofNBHoZj8zw7FAdy27HjTDywMFcV5yaSsMzC3zpQ9yoZ",
			163275366: "12iqbefziuhAzwsyRf7HiVLBzqgbcW5EVoqWQyj6iqKy5BVabsaP2aZn1rshq3vCZ4HRbyzpfxvauSzbZ1gmwKzBGobQLCG",
			163275367: "12fqNBMxMrTvTXLT5ZvG6PaUFbRFXxs82ES4GdNkCvntbK4dthcNXD1NL7WKAtgsoaWPm8azZgCN3XHkGQ1TCCGnT5QpAT44",
			163275368: "1hS8QZGbroa3X4AFc6rpLs5VPDktoqCLRrXv9Fqb1PFckfASYJ6eBSWE3iumP5pic52EJyjudBMpdkXnvtsFCnnCQ961KeL",
			163275369: "1346L8oxZGiSYAoFi3mm8GmQUcPhFsDPBDXrhDfteJssPwnkB81hoEXKwwMuPFA9FQbyMMMYF9zfhHxgP9aLqH3LvE1q2KRG",
			163275370: "12CyN7feFy9mnoC9QHHwFUgZEgAmghQG5nVNJzvWqP8qGrMYjFgP7hfzypUPt5PzxJs5Aqukobhq6GPXxoV17bUtbzMjGbPo",
			163275371: "12v6KwqVj4j3CFqqux1pk8Dse8BoroCEonbCYw75hmjkt5NppLpDDUKEmPHzvzmC3sFL4oPiEMF7N3z14W9BUaBaxqT54ej5",
			163275372: "1XPwruDEA64Fv9tMgvhhev7G9GmLw8mFjQDSdZwffKBRLen3GXBWqMmSRKNh4gTuLaZKRq4DQoVpBUiJgSP1v6a4bJv8Dg7",
			163275373: "12ucN4kCayCUuCtD5AGjRudxz37dGEt3ENWQXc5RZyerpZ2Hz3ZrqzC3YXUzCBMTBQLvYhWS5GkN6LUBPWUPr5EkxJVhy9kX",
			163275374: "12XoH5ApVf38mkbQRwgXW3VenfW4KXawir2mjjbt7KUdDiUCckKMKRj2wRARb1xPRAXUJrvmVWG8RPY2WwMNVbS9B5fHHGry",
			163275375: "12wNykBY8Evx1TypQbve6AZtaFk8EHuJACQM65vjfJAVEaQcL68nBWeRnow7MyCrfJ3Kx2GjYEvLKTq8hCbVX5RyZcovQRrd",
			163275376: "13E4xaSjPyjNPNUoqUKKdobdFN41UPRXmokVe2ahJq7T9KzPws7Q1au9LbJzX2Z1fwgUqQyuyk5oZfaiy4MkgMNh9hLdkjdQ",
			163275377: "1rhsuhhwc1Z8Tq9ZsPsV3X1SqyouDAhxJ2Mk1RYArab1mvLuXjqzcBGLMGkjkKRxPbehf9tynRx2BMPRhNfB5CHn4ZNfya8",
			163275378: "12pR2MmoBES1Np7eo25SRGQNMNFRbYbsFPAkRvyg6STQPXUcBDwyUdkGYEaV3ZDBJstaTwNABubwiuu7JpwCAvow9QnPUk2U",
			163275379: "1TEVmRy1NziqGLVpZ4f8vSJpVfkb8GFfNBEUz7weoYqaXYzWiUEPV8sRJgjRdpUgzBLC247NZ1N1KufbTyG4DBPQSNLrUbT",
			163275380: "14wuaDiVXEHbJT4z37TQhqHMCqMBnJCGRqYapx3AiJfVEXydCDdzet1Yi98NewrH6tAw5zx4VS75uWYQrk7JjbAYGMr6tL6",
			163275381: "12PcAy9oLEGzpvcDXVdTGXgdgTt95KuAbkopTZywT6zm8P74wdapKiYHWLn2KRNkhcMqLrAWFVsBLBFMS5dir2yZYpvKFeMv",
			163275382: "12mmRLASDda18D6RfGWRNNZHUMo9RhDeCvsioiwNSaQLG5qpTgULRSuMHmYPBpWkMERzPBFvi8qQWLzTBVBvAVLJKrbAMFha",
			163275383: "1j6NK4NkAGSk7ybENSrNdUX6pLe6MBv28tN4STuFhQqbLEzLFFE6NqpvscRjBhSxkeJnzGzTS3hUkRBqhmGGCguLJR4KxDY",
			163275384: "1d1DXDRvWof7V45P83k4hjCDk3VcmZeSeP2m2MzPjRrCNBGRkzoGeR6aBdT96i4ffrWx5CZXG2kf8bcGY5yx6P8Q8nbsDFD",
			163275385: "128HMLqu6MEvXmYujtBoC959dtLoTKmRrHhmYD5UqAcEu1PGoHX3dSh6f3GDf7rNuxYjSbW5vj44fbspCtyHU8ZfqoAa3xjd",
			163275386: "13K3J9z7eEjs2QfbuBgbfQgu1ejrajVg4bjJsgkUFyWNua5XqWW2otxYNE9duXBe4KJdmPv5KXnBBRDRM4CSh7bYYamCir6E",
			163275387: "13Gek6VGmapmFcLHDo8nMM49gC62sgNSY2PfzqhhUynmA29XGbeGVrZgaYURQiGskdUaJje71TvR1AGyGAzHENY2HAj5Ftey",
			163275388: "12Xzv92cdgguVCNzAFS7LANpXyXXdWAYDXFJssQgGK5236q8HPzW8RQKPmrEhVeP6A9fCvswfKihKWwKBURrzSaPZ7H5pnLm",
		},
	}
	tc2Chain := &mocks.Chain{}
	tc2Block := &mocksTypes.BlockInterface{}

	tc2Block.On("GetProposeTime").Return(int64(1632753910))
	tc2Block.On("GetProduceTime").Return(int64(1632753260))
	tc2Block.On("GetProducer").Return(shard0CommitteeString[6])
	tc2Block.On("GetProposer").Return(shard0CommitteeString[5])
	tc2Block.On("GetPrevHash").Return(common.Hash{}.NewHashFromStr2("2fc3ec987e4d0a56e344a16dbba524845f5aca6b699b0767735f0ad59926ea03"))
	tc2Block.On("GetAggregateRootHash").Return(common.Hash{}.NewHashFromStr2("698bed377b2d7c6c280924f6af6c46c7e84acddddc99268ac76e9393b3d3b05f"))

	tc3FinalityProof := map[string]map[int64]string{
		"2fc3ec987e4d0a56e344a16dbba524845f5aca6b699b0767735f0ad59926ea03": {
			163275326: "1ec352qhKtECAtECXP5pmdAtJyfCPSjjVCnL4TXZ3DQuTkSuvmT2yNBFtc4m9phexFXZqHzaoesmmC6SHYwEVDFeysuXfRH",
			163275327: "1htdgUEgxdxTYSXNUP5pPEipknKwhiD8wwngDGXZE7qRWy43VhjvXzDRHNuyn1L4SMvxwpBMvqu1UfCkQkWV3ZpdUQAQYoX",
			163275328: "1E2x3nf8K3YLdoPDQeWh8mtiS5zhErfTYvBiniidDZXYBVfET8zWkswMepQ4hd39Q7zTLZ4UTbmx9oP4QsfBdDTLVKcgQrt",
			163275329: "12F11pox3YdzBUG3816MVeFEAxmJwUBKnfBbG1NqwAwv8Kyh4ncw9Cs5Wuazc6ufVNyjec25UWuzQzs96KKE6ZnifCyUmJPD",
			163275330: "13BEZpSwNeTfCv5WYixgfB4FxWzdub9cmREauKPmNTwRbpd5TbSt7J7na9Pbr5Rw9Ekyx4gBxtSCLintpjv89P1XfRCkDvPm",
			//163275331: "12LU9ucBBdYUcATfWG4LTrcb7TBhLfCWmeNd6zKmzQGZv7r5UuwiuWHf5hyZBXjBhvy6zXQGwdKnabH3Bouky3hstGbc97Wd",
			163275332: "13WQzqFyLBCHirozUBsrhuxYaTz52H9QJktit7XxZujG5Nts7TV73vLN99ornnWzB25T6x7fDUUgpRzYokJfG43cENYsMf2U",
			163275333: "1i1vHYuXCAFeeHwFjpRVaZ3hJoosibNJmWbnJGnWjTPgwJJX9S13t9fnTz5dunTEuDfBA8hEhXQE2PKpmaJmZmDP7wpW6F1",
			163275334: "12facYZQNjAeWhN38Z1on9t9SH99ieTTt4j2keKcUGjoWfLCfUbBxooZMgMNjJnvoE4QQtyKhzgu8c2j79swyVYf5esHPHCS",
			//163275335: "12CS3AnRoTZaMeM9FP2sFhRhmxLwo5mBGkzrZtM47b5ujVDMtWi1ciea9ggSAc7b2j6XJZhxXuxa8zPJsqgJe8sExJwTgaXJ",
			163275336: "1wiQyj744VuCCEob86P5bNwNngEmM5EvCrNqyxLb4mzgi6Dpjiqbuu7Xm7XABxT5FuqME5fJP3BWEdwerj3HnEhDcWDu2Hc",
			163275337: "16q2LnGWbpR2HqN9BAewiqtPd44XMkCj4bRu4j16tWb9L2hch5gYHznFfQstVv6Rj9PPi42mAGVZQZR6LDGfVSyZv6jPN7Q",
			163275338: "1JM87NAVP9cSatv5sxviJv4HZmrb2HkyDg3MvijrwHA8FdrQ8NAWLnVhVDZVpa5xwssZBWbQSrnUFeEyRJEacnyn7T4oyTg",
			//163275339: "13R1QBhFyP2FVy4wnYbmfdAWNS8abJUf4tyFWL2sRNAc4QKKindDdvmCChdSQKYV3WMbZcoZfaPv3X278ZYUvh7gpJj4BWy6",
			163275340: "124h5SDb54eexCyr1Tqvi7peaJ9K4V8stfBdXr5LimSHsvyQoFveMoSB8zo527ipAWDqYjFB6cqEemBNT1unwjSZ5GcxaLPi",
			163275341: "12mVkWiqbDWBnh4mAeEon1Y4aN7XT98LKLKy96SNKUwQByroN69NCsBYbXDX2uVe75n6pooP4NuXSEtYFXBqGvWiaHko1npe",
			163275342: "1D7SRv14gf2xKmGBdCUPYjfdFRtHQ3TWymVixdUu5yvWoUP77gd6yekW4SJsKmDoD8gbm7ji6UtKj7gYayM5sUnGvJdjpXT",
			163275343: "13KJDV4BrJCitoqRj5KcxxEUHH3MXu7Mv4mQoHKyS7xhEWXX25nBo5SvFeMM27qXyShJJLuGE3hQLDd8bqBEsfJushTbgPXB",
			163275344: "1oNmKErZ9c7vgVXWvbLznKiFo7cBToLLU6vZ6RgJtEuPUpBkmDuShW8y3JNJtiEX48SjENw1wbsUEp3L2hEZEaD4YUXU95v",
			163275345: "12wDwm5pC6BHYHzLvuYGsskAhkYa3LvwiRPxg5ALnxCefod6e3XxvC8WEWMeSay71N4cyos7xNZGQ5KpWLTAtJ1wK4KdwZik",
			163275346: "12jz7iiYKVjYTxVduJeyi31jMPR6eqRmSC9G7SAYVYYnJoU5tSybKbwx7acJhMzFVafESm6P8esMNFTBmEEDgUjwMQvR14Gr",
			163275347: "1XxiMocds1N9qfq2xpfNaH9U7Sikqq5k12XLrHMWqpZp6q3bJCdt5kVH9MK93rKM6epQAUxuJHxYBm2B2PyP4HFXfCVZH2D",
			163275348: "12e5z6n492TSugBUd1Uy4ed6sdfAU7v8LGBJvEQiyCNJKWWvZWYG8b1Cd5Gs9VdudASs5fb9Zwm7rn5wLYr43A9VEYCjk7Au",
			163275349: "12xo97b2sSCyn8CPq3Ep1G17rz8GDpR427DxGkuBNkFGakjJwB2Dm8QrUPzLTM36sYzD58UNThQw9cZbuxnq7JkuqmKggi89",
			163275350: "164JGkqKjvNhmBFWvNCAVhz67yzakw4sRC3m7k544m3DLKP5F9eS9WVSyDxtmbrF83xiGCxLCP6vWih5vuuxKznxyPYeabG",
			163275351: "1vcE2eFWbgr6hijPoxcZwLwmtZaZiNyZkoE7jBZbtNf3wu7FUK4hXyfgPz3aQYvQdcDVF4QLAUcjaAisjUrYdPac83ReYeX",
			163275352: "13Uhc4nuu76t312ksD2dz7YNeZzp5HijK8QX8SzAPeUajjA1Ftvthz3ugQStCWDcy96PFA4iCWqouzpsNEK8py3S1cLcrbfC",
			163275353: "1E6oEPBormzpuAxXLo9sQhcVGgdpMEhkRKWcHViPj7NPRrJ4shND3XrbifJJRTZgiofnQF8fEq7dLR3LGfoHqqGBDkH39F6",
			163275354: "12iiJyC5GwNHoQcXazrGK32gWsVkjEzucpWdBuKMgUKFD3RFoQNYm8taE8MfSvmSQbVaNe2DuQcDKDAvaA67bdNUKe7pJ2bF",
			163275355: "12DEYEKQFZqAP3RbuCfFgSE2Yh58hyBdp47FwJ7LErkAK34MozUQzR8D41CfEryiVei1Gvf8bqAZcBasFLzJQR1K1j4wszZP",
			163275356: "1vwqK1hoWom2GyZv9QsasiFxwPiwqTHQdZ5SefJJBGMZYifr5thwaQtsC8DzgnvDxAdcFWxgZoJYDXh9omivqTWF7YYknmK",
			163275357: "1C9FwwxaXzpXLHoJcxeyWtMoGfGJQJ5hKQxLgZPNLhRNd7LPXnhuEuY7qxEQN89iAMuYVXUPf9REb1b3LcnTuZRgEpNwiZD",
			163275358: "13Y24AgSgBYi9sjtnNtDooLFi2kurhX6punvEKj75qRrtMDNrqxSMTMtDev69AHJ5sc9Ybb9Q5VK1gfpoishAVsUhfiJ6AmP",
			163275359: "1MkaKXhUq97yV73V7of1yoarVVAgfhwH1yTuojTwGwQTQQd2k6RxjQC6LDyhQvnTMHe3zEpJEoDFHvyticU7Tf5m4ayTU2N",
			163275360: "1rHFjq1UfG1eYmxGnv1cQ7ipChqX4BcmskXyPmCaGMfs5p6ncYRMKLcrpg1RV8ZjSGbffy8T63QRoBKdDHWjDVzrnd7yL6w",
			163275361: "16TSMUSBGi1mvYouvALMVUr6t94XtyBnjbPP5Xkkro15NDqYu2U6jHhWGyBNnpNQbebgr2FSRXr11Eq9KSD935aiSDQbeqZ",
			163275362: "12rFzDv9Z7vgsiReJqMgT9x9NTx6jd93RkGRcRymuD6otTHEFSB6RqaqBtHFJizY4tfULcrb1ZGoK3Sj9nmsKEvMc7QG8dHL",
			163275363: "12YswDF5qYuULX6BTiHM4gPdG2MiWCx1TzQp7h2oBpBw2RvCKPeuAt3PxXwfitdiUqEzvvAGeZYE3cMD3Uk9ZMAhxKoSugj7",
			163275364: "12QbaDwdDP7Y1PeP5nYJDjSY18x5WSpPjw4b9YRchAkJsjzbUfs2iCstLbZLH5vWGvvT9vCgQoWdMGLGpmYVQmso4HteniPa",
			163275365: "12fV3fabkvmEo49hXoygkxhteV46s1cjEpUuoMr6fwJbPJH3nsm4ofNBHoZj8zw7FAdy27HjTDywMFcV5yaSsMzC3zpQ9yoZ",
			163275366: "12iqbefziuhAzwsyRf7HiVLBzqgbcW5EVoqWQyj6iqKy5BVabsaP2aZn1rshq3vCZ4HRbyzpfxvauSzbZ1gmwKzBGobQLCG",
			163275367: "12fqNBMxMrTvTXLT5ZvG6PaUFbRFXxs82ES4GdNkCvntbK4dthcNXD1NL7WKAtgsoaWPm8azZgCN3XHkGQ1TCCGnT5QpAT44",
			163275368: "1hS8QZGbroa3X4AFc6rpLs5VPDktoqCLRrXv9Fqb1PFckfASYJ6eBSWE3iumP5pic52EJyjudBMpdkXnvtsFCnnCQ961KeL",
			163275369: "1346L8oxZGiSYAoFi3mm8GmQUcPhFsDPBDXrhDfteJssPwnkB81hoEXKwwMuPFA9FQbyMMMYF9zfhHxgP9aLqH3LvE1q2KRG",
			163275370: "12CyN7feFy9mnoC9QHHwFUgZEgAmghQG5nVNJzvWqP8qGrMYjFgP7hfzypUPt5PzxJs5Aqukobhq6GPXxoV17bUtbzMjGbPo",
			163275371: "12v6KwqVj4j3CFqqux1pk8Dse8BoroCEonbCYw75hmjkt5NppLpDDUKEmPHzvzmC3sFL4oPiEMF7N3z14W9BUaBaxqT54ej5",
			163275372: "1XPwruDEA64Fv9tMgvhhev7G9GmLw8mFjQDSdZwffKBRLen3GXBWqMmSRKNh4gTuLaZKRq4DQoVpBUiJgSP1v6a4bJv8Dg7",
			163275373: "12ucN4kCayCUuCtD5AGjRudxz37dGEt3ENWQXc5RZyerpZ2Hz3ZrqzC3YXUzCBMTBQLvYhWS5GkN6LUBPWUPr5EkxJVhy9kX",
			163275374: "12XoH5ApVf38mkbQRwgXW3VenfW4KXawir2mjjbt7KUdDiUCckKMKRj2wRARb1xPRAXUJrvmVWG8RPY2WwMNVbS9B5fHHGry",
			163275375: "12wNykBY8Evx1TypQbve6AZtaFk8EHuJACQM65vjfJAVEaQcL68nBWeRnow7MyCrfJ3Kx2GjYEvLKTq8hCbVX5RyZcovQRrd",
			163275376: "13E4xaSjPyjNPNUoqUKKdobdFN41UPRXmokVe2ahJq7T9KzPws7Q1au9LbJzX2Z1fwgUqQyuyk5oZfaiy4MkgMNh9hLdkjdQ",
			163275377: "1rhsuhhwc1Z8Tq9ZsPsV3X1SqyouDAhxJ2Mk1RYArab1mvLuXjqzcBGLMGkjkKRxPbehf9tynRx2BMPRhNfB5CHn4ZNfya8",
			163275378: "12pR2MmoBES1Np7eo25SRGQNMNFRbYbsFPAkRvyg6STQPXUcBDwyUdkGYEaV3ZDBJstaTwNABubwiuu7JpwCAvow9QnPUk2U",
			163275379: "1TEVmRy1NziqGLVpZ4f8vSJpVfkb8GFfNBEUz7weoYqaXYzWiUEPV8sRJgjRdpUgzBLC247NZ1N1KufbTyG4DBPQSNLrUbT",
			163275380: "14wuaDiVXEHbJT4z37TQhqHMCqMBnJCGRqYapx3AiJfVEXydCDdzet1Yi98NewrH6tAw5zx4VS75uWYQrk7JjbAYGMr6tL6",
			163275381: "12PcAy9oLEGzpvcDXVdTGXgdgTt95KuAbkopTZywT6zm8P74wdapKiYHWLn2KRNkhcMqLrAWFVsBLBFMS5dir2yZYpvKFeMv",
			163275382: "12mmRLASDda18D6RfGWRNNZHUMo9RhDeCvsioiwNSaQLG5qpTgULRSuMHmYPBpWkMERzPBFvi8qQWLzTBVBvAVLJKrbAMFha",
			163275383: "1j6NK4NkAGSk7ybENSrNdUX6pLe6MBv28tN4STuFhQqbLEzLFFE6NqpvscRjBhSxkeJnzGzTS3hUkRBqhmGGCguLJR4KxDY",
			163275384: "1d1DXDRvWof7V45P83k4hjCDk3VcmZeSeP2m2MzPjRrCNBGRkzoGeR6aBdT96i4ffrWx5CZXG2kf8bcGY5yx6P8Q8nbsDFD",
			163275385: "128HMLqu6MEvXmYujtBoC959dtLoTKmRrHhmYD5UqAcEu1PGoHX3dSh6f3GDf7rNuxYjSbW5vj44fbspCtyHU8ZfqoAa3xjd",
			163275386: "13K3J9z7eEjs2QfbuBgbfQgu1ejrajVg4bjJsgkUFyWNua5XqWW2otxYNE9duXBe4KJdmPv5KXnBBRDRM4CSh7bYYamCir6E",
			163275387: "13Gek6VGmapmFcLHDo8nMM49gC62sgNSY2PfzqhhUynmA29XGbeGVrZgaYURQiGskdUaJje71TvR1AGyGAzHENY2HAj5Ftey",
			163275388: "12Xzv92cdgguVCNzAFS7LANpXyXXdWAYDXFJssQgGK5236q8HPzW8RQKPmrEhVeP6A9fCvswfKihKWwKBURrzSaPZ7H5pnLm",
		},
	}
	tc3Chain := &mocks.Chain{}
	tc3Block := &mocksTypes.BlockInterface{}

	tc3Block.On("GetProposeTime").Return(int64(1632753910))
	tc3Block.On("GetProduceTime").Return(int64(1632753260))
	tc3Block.On("GetProducer").Return(shard0CommitteeString[6])
	tc3Block.On("GetProposer").Return(shard0CommitteeString[5])
	tc3Block.On("GetPrevHash").Return(common.Hash{}.NewHashFromStr2("2fc3ec987e4d0a56e344a16dbba524845f5aca6b699b0767735f0ad59926ea03"))
	tc3Block.On("GetAggregateRootHash").Return(common.Hash{}.NewHashFromStr2("698bed377b2d7c6c280924f6af6c46c7e84acddddc99268ac76e9393b3d3b05f"))

	tc4FinalityProof := map[string]map[int64]string{
		"2fc3ec987e4d0a56e344a16dbba524845f5aca6b699b0767735f0ad59926ea03": {
			163275326: "1ec352qhKtECAtECXP5pmdAtJyfCPSjjVCnL4TXZ3DQuTkSuvmT2yNBFtc4m9phexFXZqHzaoesmmC6SHYwEVDFeysuXfRH",
			163275327: "1htdgUEgxdxTYSXNUP5pPEipknKwhiD8wwngDGXZE7qRWy43VhjvXzDRHNuyn1L4SMvxwpBMvqu1UfCkQkWV3ZpdUQAQYoX",
			163275328: "1E2x3nf8K3YLdoPDQeWh8mtiS5zhErfTYvBiniidDZXYBVfET8zWkswMepQ4hd39Q7zTLZ4UTbmx9oP4QsfBdDTLVKcgQrt",
			163275329: "12F11pox3YdzBUG3816MVeFEAxmJwUBKnfBbG1NqwAwv8Kyh4ncw9Cs5Wuazc6ufVNyjec25UWuzQzs96KKE6ZnifCyUmJPD",
			163275330: "13BEZpSwNeTfCv5WYixgfB4FxWzdub9cmREauKPmNTwRbpd5TbSt7J7na9Pbr5Rw9Ekyx4gBxtSCLintpjv89P1XfRCkDvPm",
			163275331: "12LU9ucBBdYUcATfWG4LTrcb7TBhLfCWmeNd6zKmzQGZv7r5UuwiuWHf5hyZBXjBhvy6zXQGwdKnabH3Bouky3hstGbc97Wd",
			163275332: "13WQzqFyLBCHirozUBsrhuxYaTz52H9QJktit7XxZujG5Nts7TV73vLN99ornnWzB25T6x7fDUUgpRzYokJfG43cENYsMf2U",
			163275333: "1i1vHYuXCAFeeHwFjpRVaZ3hJoosibNJmWbnJGnWjTPgwJJX9S13t9fnTz5dunTEuDfBA8hEhXQE2PKpmaJmZmDP7wpW6F1",
			163275334: "12facYZQNjAeWhN38Z1on9t9SH99ieTTt4j2keKcUGjoWfLCfUbBxooZMgMNjJnvoE4QQtyKhzgu8c2j79swyVYf5esHPHCS",
			163275335: "12CS3AnRoTZaMeM9FP2sFhRhmxLwo5mBGkzrZtM47b5ujVDMtWi1ciea9ggSAc7b2j6XJZhxXuxa8zPJsqgJe8sExJwTgaXJ",
			163275336: "1wiQyj744VuCCEob86P5bNwNngEmM5EvCrNqyxLb4mzgi6Dpjiqbuu7Xm7XABxT5FuqME5fJP3BWEdwerj3HnEhDcWDu2Hc",
			163275337: "16q2LnGWbpR2HqN9BAewiqtPd44XMkCj4bRu4j16tWb9L2hch5gYHznFfQstVv6Rj9PPi42mAGVZQZR6LDGfVSyZv6jPN7Q",
			163275338: "1JM87NAVP9cSatv5sxviJv4HZmrb2HkyDg3MvijrwHA8FdrQ8NAWLnVhVDZVpa5xwssZBWbQSrnUFeEyRJEacnyn7T4oyTg",
			163275339: "13R1QBhFyP2FVy4wnYbmfdAWNS8abJUf4tyFWL2sRNAc4QKKindDdvmCChdSQKYV3WMbZcoZfaPv3X278ZYUvh7gpJj4BWy6",
			163275340: "124h5SDb54eexCyr1Tqvi7peaJ9K4V8stfBdXr5LimSHsvyQoFveMoSB8zo527ipAWDqYjFB6cqEemBNT1unwjSZ5GcxaLPi",
			163275341: "12mVkWiqbDWBnh4mAeEon1Y4aN7XT98LKLKy96SNKUwQByroN69NCsBYbXDX2uVe75n6pooP4NuXSEtYFXBqGvWiaHko1npe",
			163275342: "1D7SRv14gf2xKmGBdCUPYjfdFRtHQ3TWymVixdUu5yvWoUP77gd6yekW4SJsKmDoD8gbm7ji6UtKj7gYayM5sUnGvJdjpXT",
			163275343: "13KJDV4BrJCitoqRj5KcxxEUHH3MXu7Mv4mQoHKyS7xhEWXX25nBo5SvFeMM27qXyShJJLuGE3hQLDd8bqBEsfJushTbgPXB",
			163275344: "1oNmKErZ9c7vgVXWvbLznKiFo7cBToLLU6vZ6RgJtEuPUpBkmDuShW8y3JNJtiEX48SjENw1wbsUEp3L2hEZEaD4YUXU95v",
			163275345: "12wDwm5pC6BHYHzLvuYGsskAhkYa3LvwiRPxg5ALnxCefod6e3XxvC8WEWMeSay71N4cyos7xNZGQ5KpWLTAtJ1wK4KdwZik",
			163275346: "12jz7iiYKVjYTxVduJeyi31jMPR6eqRmSC9G7SAYVYYnJoU5tSybKbwx7acJhMzFVafESm6P8esMNFTBmEEDgUjwMQvR14Gr",
			163275347: "1XxiMocds1N9qfq2xpfNaH9U7Sikqq5k12XLrHMWqpZp6q3bJCdt5kVH9MK93rKM6epQAUxuJHxYBm2B2PyP4HFXfCVZH2D",
			163275348: "12e5z6n492TSugBUd1Uy4ed6sdfAU7v8LGBJvEQiyCNJKWWvZWYG8b1Cd5Gs9VdudASs5fb9Zwm7rn5wLYr43A9VEYCjk7Au",
			163275349: "12xo97b2sSCyn8CPq3Ep1G17rz8GDpR427DxGkuBNkFGakjJwB2Dm8QrUPzLTM36sYzD58UNThQw9cZbuxnq7JkuqmKggi89",
			163275350: "164JGkqKjvNhmBFWvNCAVhz67yzakw4sRC3m7k544m3DLKP5F9eS9WVSyDxtmbrF83xiGCxLCP6vWih5vuuxKznxyPYeabG",
			163275351: "1vcE2eFWbgr6hijPoxcZwLwmtZaZiNyZkoE7jBZbtNf3wu7FUK4hXyfgPz3aQYvQdcDVF4QLAUcjaAisjUrYdPac83ReYeX",
			163275352: "13Uhc4nuu76t312ksD2dz7YNeZzp5HijK8QX8SzAPeUajjA1Ftvthz3ugQStCWDcy96PFA4iCWqouzpsNEK8py3S1cLcrbfC",
			163275353: "1E6oEPBormzpuAxXLo9sQhcVGgdpMEhkRKWcHViPj7NPRrJ4shND3XrbifJJRTZgiofnQF8fEq7dLR3LGfoHqqGBDkH39F6",
			163275354: "12iiJyC5GwNHoQcXazrGK32gWsVkjEzucpWdBuKMgUKFD3RFoQNYm8taE8MfSvmSQbVaNe2DuQcDKDAvaA67bdNUKe7pJ2bF",
			163275355: "12DEYEKQFZqAP3RbuCfFgSE2Yh58hyBdp47FwJ7LErkAK34MozUQzR8D41CfEryiVei1Gvf8bqAZcBasFLzJQR1K1j4wszZP",
			163275356: "1vwqK1hoWom2GyZv9QsasiFxwPiwqTHQdZ5SefJJBGMZYifr5thwaQtsC8DzgnvDxAdcFWxgZoJYDXh9omivqTWF7YYknmK",
			163275357: "1C9FwwxaXzpXLHoJcxeyWtMoGfGJQJ5hKQxLgZPNLhRNd7LPXnhuEuY7qxEQN89iAMuYVXUPf9REb1b3LcnTuZRgEpNwiZD",
			163275358: "13Y24AgSgBYi9sjtnNtDooLFi2kurhX6punvEKj75qRrtMDNrqxSMTMtDev69AHJ5sc9Ybb9Q5VK1gfpoishAVsUhfiJ6AmP",
			163275359: "1MkaKXhUq97yV73V7of1yoarVVAgfhwH1yTuojTwGwQTQQd2k6RxjQC6LDyhQvnTMHe3zEpJEoDFHvyticU7Tf5m4ayTU2N",
			163275360: "1rHFjq1UfG1eYmxGnv1cQ7ipChqX4BcmskXyPmCaGMfs5p6ncYRMKLcrpg1RV8ZjSGbffy8T63QRoBKdDHWjDVzrnd7yL6w",
			163275361: "16TSMUSBGi1mvYouvALMVUr6t94XtyBnjbPP5Xkkro15NDqYu2U6jHhWGyBNnpNQbebgr2FSRXr11Eq9KSD935aiSDQbeqZ",
			163275362: "12rFzDv9Z7vgsiReJqMgT9x9NTx6jd93RkGRcRymuD6otTHEFSB6RqaqBtHFJizY4tfULcrb1ZGoK3Sj9nmsKEvMc7QG8dHL",
			163275363: "12YswDF5qYuULX6BTiHM4gPdG2MiWCx1TzQp7h2oBpBw2RvCKPeuAt3PxXwfitdiUqEzvvAGeZYE3cMD3Uk9ZMAhxKoSugj7",
			163275364: "12QbaDwdDP7Y1PeP5nYJDjSY18x5WSpPjw4b9YRchAkJsjzbUfs2iCstLbZLH5vWGvvT9vCgQoWdMGLGpmYVQmso4HteniPa",
			163275365: "12fV3fabkvmEo49hXoygkxhteV46s1cjEpUuoMr6fwJbPJH3nsm4ofNBHoZj8zw7FAdy27HjTDywMFcV5yaSsMzC3zpQ9yoZ",
			163275366: "12iqbefziuhAzwsyRf7HiVLBzqgbcW5EVoqWQyj6iqKy5BVabsaP2aZn1rshq3vCZ4HRbyzpfxvauSzbZ1gmwKzBGobQLCG",
			163275367: "12fqNBMxMrTvTXLT5ZvG6PaUFbRFXxs82ES4GdNkCvntbK4dthcNXD1NL7WKAtgsoaWPm8azZgCN3XHkGQ1TCCGnT5QpAT44",
			163275368: "1hS8QZGbroa3X4AFc6rpLs5VPDktoqCLRrXv9Fqb1PFckfASYJ6eBSWE3iumP5pic52EJyjudBMpdkXnvtsFCnnCQ961KeL",
			163275369: "1346L8oxZGiSYAoFi3mm8GmQUcPhFsDPBDXrhDfteJssPwnkB81hoEXKwwMuPFA9FQbyMMMYF9zfhHxgP9aLqH3LvE1q2KRG",
			163275370: "12CyN7feFy9mnoC9QHHwFUgZEgAmghQG5nVNJzvWqP8qGrMYjFgP7hfzypUPt5PzxJs5Aqukobhq6GPXxoV17bUtbzMjGbPo",
			163275371: "12v6KwqVj4j3CFqqux1pk8Dse8BoroCEonbCYw75hmjkt5NppLpDDUKEmPHzvzmC3sFL4oPiEMF7N3z14W9BUaBaxqT54ej5",
			163275372: "1XPwruDEA64Fv9tMgvhhev7G9GmLw8mFjQDSdZwffKBRLen3GXBWqMmSRKNh4gTuLaZKRq4DQoVpBUiJgSP1v6a4bJv8Dg7",
			163275373: "12ucN4kCayCUuCtD5AGjRudxz37dGEt3ENWQXc5RZyerpZ2Hz3ZrqzC3YXUzCBMTBQLvYhWS5GkN6LUBPWUPr5EkxJVhy9kX",
			163275374: "12XoH5ApVf38mkbQRwgXW3VenfW4KXawir2mjjbt7KUdDiUCckKMKRj2wRARb1xPRAXUJrvmVWG8RPY2WwMNVbS9B5fHHGry",
			163275375: "12wNykBY8Evx1TypQbve6AZtaFk8EHuJACQM65vjfJAVEaQcL68nBWeRnow7MyCrfJ3Kx2GjYEvLKTq8hCbVX5RyZcovQRrd",
			163275376: "13E4xaSjPyjNPNUoqUKKdobdFN41UPRXmokVe2ahJq7T9KzPws7Q1au9LbJzX2Z1fwgUqQyuyk5oZfaiy4MkgMNh9hLdkjdQ",
			163275377: "1rhsuhhwc1Z8Tq9ZsPsV3X1SqyouDAhxJ2Mk1RYArab1mvLuXjqzcBGLMGkjkKRxPbehf9tynRx2BMPRhNfB5CHn4ZNfya8",
			163275378: "12pR2MmoBES1Np7eo25SRGQNMNFRbYbsFPAkRvyg6STQPXUcBDwyUdkGYEaV3ZDBJstaTwNABubwiuu7JpwCAvow9QnPUk2U",
			163275379: "1TEVmRy1NziqGLVpZ4f8vSJpVfkb8GFfNBEUz7weoYqaXYzWiUEPV8sRJgjRdpUgzBLC247NZ1N1KufbTyG4DBPQSNLrUbT",
			163275380: "14wuaDiVXEHbJT4z37TQhqHMCqMBnJCGRqYapx3AiJfVEXydCDdzet1Yi98NewrH6tAw5zx4VS75uWYQrk7JjbAYGMr6tL6",
			163275381: "12PcAy9oLEGzpvcDXVdTGXgdgTt95KuAbkopTZywT6zm8P74wdapKiYHWLn2KRNkhcMqLrAWFVsBLBFMS5dir2yZYpvKFeMv",
			163275382: "12mmRLASDda18D6RfGWRNNZHUMo9RhDeCvsioiwNSaQLG5qpTgULRSuMHmYPBpWkMERzPBFvi8qQWLzTBVBvAVLJKrbAMFha",
			163275383: "1j6NK4NkAGSk7ybENSrNdUX6pLe6MBv28tN4STuFhQqbLEzLFFE6NqpvscRjBhSxkeJnzGzTS3hUkRBqhmGGCguLJR4KxDY",
			163275384: "1d1DXDRvWof7V45P83k4hjCDk3VcmZeSeP2m2MzPjRrCNBGRkzoGeR6aBdT96i4ffrWx5CZXG2kf8bcGY5yx6P8Q8nbsDFD",
			163275385: "128HMLqu6MEvXmYujtBoC959dtLoTKmRrHhmYD5UqAcEu1PGoHX3dSh6f3GDf7rNuxYjSbW5vj44fbspCtyHU8ZfqoAa3xjd",
			163275386: "13K3J9z7eEjs2QfbuBgbfQgu1ejrajVg4bjJsgkUFyWNua5XqWW2otxYNE9duXBe4KJdmPv5KXnBBRDRM4CSh7bYYamCir6E",
			163275387: "13Gek6VGmapmFcLHDo8nMM49gC62sgNSY2PfzqhhUynmA29XGbeGVrZgaYURQiGskdUaJje71TvR1AGyGAzHENY2HAj5Ftey",
			163275388: "12Xzv92cdgguVCNzAFS7LANpXyXXdWAYDXFJssQgGK5236q8HPzW8RQKPmrEhVeP6A9fCvswfKihKWwKBURrzSaPZ7H5pnLm",
		},
	}
	tc4Chain := &mocks.Chain{}
	tc4Block := &mocksTypes.BlockInterface{}

	tc4Block.On("GetProposeTime").Return(int64(1632753910))
	tc4Block.On("GetProduceTime").Return(int64(1632753260))
	tc4Block.On("GetProducer").Return(shard0CommitteeString[6])
	tc4Block.On("GetProposer").Return(shard0CommitteeString[5])
	tc4Block.On("GetPrevHash").Return(common.Hash{}.NewHashFromStr2("2fc3ec987e4d0a56e344a16dbba524845f5aca6b699b0767735f0ad59926ea03"))
	tc4Block.On("GetAggregateRootHash").Return(common.Hash{}.NewHashFromStr2("698bed377b2d7c6c280924f6af6c46c7e84acddddc99268ac76e9393b3d3b05f"))

	type fields struct {
		logger                 common.Logger
		nextBlockFinalityProof map[string]map[int64]string
		chain                  Chain
	}
	type args struct {
		block           types.BlockInterface
		currentTimeSlot int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *FinalityProof
		want1  bool
	}{
		{
			name: "tc1: nil block",
			fields: fields{
				chain:                  tc1Chain,
				logger:                 logger,
				nextBlockFinalityProof: tc1FinalityProof,
			},
			args: args{
				block:           nil,
				currentTimeSlot: 163275391,
			},
			want:  NewFinalityProof(),
			want1: false,
		},
		{
			name: "tc2: reach max finality proof",
			fields: fields{
				chain:                  tc2Chain,
				logger:                 logger,
				nextBlockFinalityProof: tc2FinalityProof,
			},
			args: args{
				block:           tc2Block,
				currentTimeSlot: 163275391,
			},
			want:  NewFinalityProof(),
			want1: false,
		},
		{
			name: "tc3: missing proof",
			fields: fields{
				chain:                  tc3Chain,
				logger:                 logger,
				nextBlockFinalityProof: tc3FinalityProof,
			},
			args: args{
				block:           tc3Block,
				currentTimeSlot: 163275388,
			},
			want:  NewFinalityProof(),
			want1: false,
		},
		{
			name: "tc4: success add proof",
			fields: fields{
				chain:                  tc4Chain,
				logger:                 logger,
				nextBlockFinalityProof: tc4FinalityProof,
			},
			args: args{
				block:           tc4Block,
				currentTimeSlot: 163275388,
			},
			want: &FinalityProof{
				ReProposeHashSignature: []string{
					"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",
				},
			},
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ProposeRuleLemma2{
				logger:                 tt.fields.logger,
				nextBlockFinalityProof: tt.fields.nextBlockFinalityProof,
				chain:                  tt.fields.chain,
			}
			got, got1, _ := p.GetValidFinalityProof(tt.args.block, tt.args.currentTimeSlot)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetValidFinalityProof() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetValidFinalityProof() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
