package blockchain

import (
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
