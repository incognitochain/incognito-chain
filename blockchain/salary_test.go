package blockchain

import (
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
	"testing"
)

func Test_getNoBlkPerYear(t *testing.T) {
	type args struct {
		blockCreationTimeSeconds uint64
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{
			name: "40s",
			args: args{blockCreationTimeSeconds: 40},
			want: 788940,
		},
		{
			name: "10s",
			args: args{blockCreationTimeSeconds: 10},
			want: 3155760,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getNoBlkPerYear(tt.args.blockCreationTimeSeconds); got != tt.want {
				t.Errorf("getNoBlkPerYear() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockChain_getRewardAmount(t *testing.T) {
	numberOfBlockPerYear := getNoBlkPerYear(40)
	type fields struct {
		config Config
	}
	type args struct {
		blkHeight uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint64
	}{
		{
			name: "Mainnet year 1",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				1,
			},
			want: 1386666000,
		},
		{
			name: "Mainnet year 1",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear,
			},
			want: 1386666000,
		},
		{
			name: "Mainnet year 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear + 1,
			},
			want: 1261866060,
		},
		{
			name: "Mainnet year 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear * 2,
			},
			want: 1261866060,
		},
		{
			name: "Mainnet year 3",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear*2 + 1,
			},
			want: 1148298114,
		},
		{
			name: "Mainnet year 3",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear * 3,
			},
			want: 1148298114,
		},
		{
			name: "Mainnet year 4",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear*3 + 1,
			},
			want: 1044951283,
		},
		{
			name: "Mainnet year 4",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear * 4,
			},
			want: 1044951283,
		},
		{
			name: "Mainnet year 5",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear*4 + 1,
			},
			want: 950905667,
		},
		{
			name: "Mainnet year 5",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear * 5,
			},
			want: 950905667,
		},
		{
			name: "Mainnet year 6",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear*5 + 1,
			},
			want: 865324156,
		},
		{
			name: "Mainnet year 6",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear * 6,
			},
			want: 865324156,
		},
		{
			name: "Mainnet year 7",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear*6 + 1,
			},
			want: 787444981,
		},
		{
			name: "Mainnet year 7",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear * 7,
			},
			want: 787444981,
		},
		{
			name: "Mainnet year 8",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear*7 + 1,
			},
			want: 716574932,
		},
		{
			name: "Mainnet year 8",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear * 8,
			},
			want: 716574932,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockchain := &BlockChain{
				config: tt.fields.config,
			}
			if got := blockchain.getRewardAmount(tt.args.blkHeight); got != tt.want {
				t.Errorf("getRewardAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_splitReward(t *testing.T) {
	type args struct {
		totalReward               *map[common.Hash]uint64
		numberOfActiveShards      int
		devPercent                int
		isSplitRewardForCustodian bool
		percentCustodianRewards   uint64
	}
	totalRewardYear1 := make(map[common.Hash]uint64)
	totalRewardYear1[common.PRVCoinID] = 8751970
	beaconRewardYear1 := make(map[common.Hash]uint64)
	beaconRewardYear1[common.PRVCoinID] = 1575354
	daoRewardYear1 := make(map[common.Hash]uint64)
	daoRewardYear1[common.PRVCoinID] = 875197
	custodianRewardYear1 := make(map[common.Hash]uint64)
	shardRewardYear1 := make(map[common.Hash]uint64)
	shardRewardYear1[common.PRVCoinID] = 6301419

	totalRewardYear2 := make(map[common.Hash]uint64)
	totalRewardYear2[common.PRVCoinID] = 7964293
	beaconRewardYear2 := make(map[common.Hash]uint64)
	beaconRewardYear2[common.PRVCoinID] = 1449501
	daoRewardYear2 := make(map[common.Hash]uint64)
	daoRewardYear2[common.PRVCoinID] = 716786
	custodianRewardYear2 := make(map[common.Hash]uint64)
	shardRewardYear2 := make(map[common.Hash]uint64)
	shardRewardYear2[common.PRVCoinID] = 5798006
	tests := []struct {
		name    string
		args    args
		want    *map[common.Hash]uint64
		want1   *map[common.Hash]uint64
		want2   *map[common.Hash]uint64
		want3   *map[common.Hash]uint64
		wantErr bool
	}{
		{
			name: "year 1",
			args: args{
				totalReward:               &totalRewardYear1,
				numberOfActiveShards:      8,
				devPercent:                10,
				isSplitRewardForCustodian: false,
				percentCustodianRewards:   0,
			},
			want:  &beaconRewardYear1,
			want1: &daoRewardYear1,
			want2: &custodianRewardYear1,
			want3: &shardRewardYear1,
		},
		{
			name: "year 2",
			args: args{
				totalReward:               &totalRewardYear2,
				numberOfActiveShards:      8,
				devPercent:                9,
				isSplitRewardForCustodian: false,
				percentCustodianRewards:   0,
			},
			want:  &beaconRewardYear2,
			want1: &daoRewardYear2,
			want2: &custodianRewardYear2,
			want3: &shardRewardYear2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, err := splitReward(tt.args.totalReward, tt.args.numberOfActiveShards, tt.args.devPercent, tt.args.isSplitRewardForCustodian, tt.args.percentCustodianRewards)
			if (err != nil) != tt.wantErr {
				t.Errorf("splitReward() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitReward() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("splitReward() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("splitReward() got2 = %v, want %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(tt.args.totalReward, tt.want3) {
				t.Errorf("splitReward() totalReward = %v, want %v", tt.args.totalReward, tt.want3)
			}
		})
	}
}

func Test_getPercentForIncognitoDAO(t *testing.T) {
	type args struct {
		blockHeight uint64
		blkPerYear  uint64
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "year 1",
			args: args{
				blockHeight: 788940,
				blkPerYear:  788940,
			},
			want: 10,
		},
		{
			name: "year 2-1",
			args: args{
				blockHeight: 788941,
				blkPerYear:  788940,
			},
			want: 9,
		}, {
			name: "year 2-2",
			args: args{
				blockHeight: 1577880,
				blkPerYear:  788940,
			},
			want: 9,
		},
		{
			name: "year 3-1",
			args: args{
				blockHeight: 1577881,
				blkPerYear:  788940,
			},
			want: 8,
		},
		{
			name: "year 3-2",
			args: args{
				blockHeight: 2366820,
				blkPerYear:  788940,
			},
			want: 8,
		},
		{
			name: "year 4-1",
			args: args{
				blockHeight: 2366821,
				blkPerYear:  788940,
			},
			want: 7,
		},
		{
			name: "year 4-2",
			args: args{
				blockHeight: 3155760,
				blkPerYear:  788940,
			},
			want: 7,
		},
		{
			name: "year 5-1",
			args: args{
				blockHeight: 3155761,
				blkPerYear:  788940,
			},
			want: 6,
		},
		{
			name: "year 5-2",
			args: args{
				blockHeight: 3944700,
				blkPerYear:  788940,
			},
			want: 6,
		},
		{
			name: "year 6-1",
			args: args{
				blockHeight: 3944701,
				blkPerYear:  788940,
			},
			want: 5,
		},
		{
			name: "year 6-2",
			args: args{
				blockHeight: 4733640,
				blkPerYear:  788940,
			},
			want: 5,
		},
		{
			name: "year 7",
			args: args{
				blockHeight: 5522580,
				blkPerYear:  788940,
			},
			want: 4,
		},
		{
			name: "year 8",
			args: args{
				blockHeight: 6311520,
				blkPerYear:  788940,
			},
			want: 3,
		},
		{
			name: "year 9",
			args: args{
				blockHeight: 7100460,
				blkPerYear:  788940,
			},
			want: 3,
		},
		{
			name: "year 10",
			args: args{
				blockHeight: 7889400,
				blkPerYear:  788940,
			},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPercentForIncognitoDAO(tt.args.blockHeight, tt.args.blkPerYear); got != tt.want {
				t.Errorf("getPercentForIncognitoDAO() = %v, want %v", got, tt.want)
			}
		})
	}
}
