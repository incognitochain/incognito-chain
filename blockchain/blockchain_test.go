package blockchain

import (
	"testing"
)

func TestBlockChain_GetCurrentEpochLength(t *testing.T) {
	type fields struct {
		config Config
	}
	type args struct {
		beaconHeight uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint64
	}{
		{
			name: "< break point",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 299,
			},
			want: 100,
		}, {
			name: "= break point",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 300,
			},
			want: 100,
		},
		{
			name: "> break point",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 301,
			},
			want: 350,
		},
		{
			name: "> break point",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 302,
			},
			want: 350,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &BlockChain{
				config: tt.fields.config,
			}
			if got := bc.GetCurrentEpochLength(tt.args.beaconHeight); got != tt.want {
				t.Errorf("GetCurrentEpochLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockChain_GetEpochByHeight(t *testing.T) {
	type fields struct {
		config Config
	}
	type args struct {
		beaconHeight uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint64
	}{
		{
			name: "< break point",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 299,
			},
			want: 3,
		},
		{
			name: "= break point",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 300,
			},
			want: 3,
		},
		{
			name: "> break point",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 301,
			},
			want: 4,
		},
		{
			name: "> break point 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 650,
			},
			want: 4,
		},
		{
			name: "> break point 3",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 651,
			},
			want: 5,
		},
		{
			name: "> break point 4",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 1000,
			},
			want: 5,
		},
		{
			name: "> break point 5",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 1001,
			},
			want: 6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &BlockChain{
				config: tt.fields.config,
			}
			if got := bc.GetEpochByHeight(tt.args.beaconHeight); got != tt.want {
				t.Errorf("GetEpochByHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockChain_GetEpochNextHeight(t *testing.T) {
	type fields struct {
		config Config
	}
	type args struct {
		beaconHeight uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint64
		want1  bool
	}{
		{
			name: "< break point 1",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 299,
			},
			want:  3,
			want1: false,
		},
		{
			name: "< break point 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 2,
			},
			want:  1,
			want1: false,
		},
		{
			name: "< break point 3",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 1,
			},
			want:  1,
			want1: false,
		},
		{
			name: "< break point 4",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 100,
			},
			want:  2,
			want1: true,
		},
		{
			name: "< break point 5",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             20,
						EpochV2:           50,
						EpochV2BreakPoint: 10,
					},
				},
			},
			args: args{
				beaconHeight: 20,
			},
			want:  2,
			want1: true,
		},
		{
			name: "= break point 1",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 299,
			},
			want:  3,
			want1: false,
		},
		{
			name: "= break point 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 300,
			},
			want:  4,
			want1: true,
		},
		{
			name: "= break point 3",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 301,
			},
			want:  4,
			want1: false,
		},
		{
			name: "> break point",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 301,
			},
			want:  4,
			want1: false,
		},
		{
			name: "> break point 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 650,
			},
			want:  5,
			want1: true,
		},
		{
			name: "> break point 3",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 649,
			},
			want:  4,
			want1: false,
		},
		{
			name: "> break point 4",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 651,
			},
			want:  5,
			want1: false,
		},
		{
			name: "> break point 5",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 1000,
			},
			want:  6,
			want1: true,
		},
		{
			name: "> break point 6",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 1001,
			},
			want:  6,
			want1: false,
		},
		{
			name: "> break point 7",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 999,
			},
			want:  5,
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &BlockChain{
				config: tt.fields.config,
			}
			got, got1 := bc.GetEpochNextHeight(tt.args.beaconHeight)
			if got != tt.want {
				t.Errorf("GetEpochNextHeight() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetEpochNextHeight() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestBlockChain_GetRandomTimeOfCurrentEpoch(t *testing.T) {
	type fields struct {
		config Config
	}
	type args struct {
		epoch uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint64
	}{
		{
			name: "< break point 1",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				epoch: 2,
			},
			want: 150,
		},
		{
			name: "< break point 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				epoch: 3,
			},
			want: 250,
		},
		{
			name: "= break point",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				epoch: 4,
			},
			want: 300 + 175,
		},
		{
			name: "> break point",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				epoch: 5,
			},
			want: 300 + 350 + 175,
		},
		{
			name: "> break point 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				epoch: 6,
			},
			want: 300 + 350*2 + 175,
		},
		{
			name: "> break point 3",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				epoch: 7,
			},
			want: 300 + 350*3 + 175,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &BlockChain{
				config: tt.fields.config,
			}
			if got := bc.GetRandomTimeInEpoch(tt.args.epoch); got != tt.want {
				t.Errorf("GetRandomTimeInEpoch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockChain_GetFirstBeaconHeightInEpoch(t *testing.T) {
	type fields struct {
		config Config
	}
	type args struct {
		epoch uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint64
	}{
		{
			name: "< break point 1",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				epoch: 2,
			},
			want: 101,
		},
		{
			name: "< break point 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				epoch: 3,
			},
			want: 201,
		},
		{
			name: "= break point",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				epoch: 4,
			},
			want: 301,
		},
		{
			name: "> break point",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				epoch: 5,
			},
			want: 300 + 350 + 1,
		},
		{
			name: "> break point 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				epoch: 6,
			},
			want: 300 + 350*2 + 1,
		},
		{
			name: "> break point 3",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				epoch: 7,
			},
			want: 300 + 350*3 + 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &BlockChain{
				config: tt.fields.config,
			}
			if got := bc.GetFirstBeaconHeightInEpoch(tt.args.epoch); got != tt.want {
				t.Errorf("GetFirstBeaconHeightInEpoch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockChain_GetLastBeaconHeightInEpoch(t *testing.T) {
	type fields struct {
		config Config
	}
	type args struct {
		epoch uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint64
	}{
		{
			name: "< break point 1",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				epoch: 2,
			},
			want: 200,
		},
		{
			name: "< break point 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				epoch: 3,
			},
			want: 300,
		},
		{
			name: "= break point",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				epoch: 4,
			},
			want: 650,
		},
		{
			name: "> break point",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				epoch: 5,
			},
			want: 300 + 350 + 350,
		},
		{
			name: "> break point 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				epoch: 6,
			},
			want: 300 + 350*2 + 350,
		},
		{
			name: "> break point 3",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				epoch: 7,
			},
			want: 300 + 350*3 + 350,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &BlockChain{
				config: tt.fields.config,
			}
			if got := bc.GetLastBeaconHeightInEpoch(tt.args.epoch); got != tt.want {
				t.Errorf("GetLastBeaconHeightInEpoch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockChain_GetBeaconBlockOrderInEpoch(t *testing.T) {
	type fields struct {
		config Config
	}
	type args struct {
		beaconHeight uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint64
		want1  uint64
	}{
		{
			name: "< break point 1",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 170,
			},
			want:  70,
			want1: 30,
		},
		{
			name: "< break point 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 264,
			},
			want:  64,
			want1: 36,
		},
		{
			name: "= break point 1",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 300,
			},
			want:  0,
			want1: 350,
		},
		{
			name: "= break point 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 301,
			},
			want:  1,
			want1: 349,
		},
		{
			name: "= break point 3",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 402,
			},
			want:  102,
			want1: 248,
		},
		{
			name: "> break point",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 734,
			},
			want:  84,
			want1: 266,
		},
		{
			name: "> break point 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 888,
			},
			want:  238,
			want1: 112,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &BlockChain{
				config: tt.fields.config,
			}
			got, got1 := bc.GetBeaconBlockOrderInEpoch(tt.args.beaconHeight)
			if got != tt.want {
				t.Errorf("GetBeaconBlockOrderInEpoch() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetBeaconBlockOrderInEpoch() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestBlockChain_IsGreaterThanRandomTime(t *testing.T) {
	type fields struct {
		config Config
	}
	type args struct {
		beaconHeight uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "< break point 1",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 150,
			},
			want: true,
		},
		{
			name: "< break point 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 149,
			},
			want: false,
		},
		{
			name: "< break point 3",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 250,
			},
			want: true,
		},
		{
			name: "< break point 4",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 249,
			},
			want: false,
		},
		{
			name: "= break point 1",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 474,
			},
			want: false,
		},
		{
			name: "= break point 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 475,
			},
			want: true,
		},
		{
			name: "= break point 3",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 476,
			},
			want: false,
		},
		{
			name: "= break point 4",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 402,
			},
			want: false,
		},
		{
			name: "> break point",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 734,
			},
			want: false,
		},
		{
			name: "> break point 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 888,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &BlockChain{
				config: tt.fields.config,
			}
			if got := bc.IsGreaterThanRandomTime(tt.args.beaconHeight); got != tt.want {
				t.Errorf("IsGreaterThanRandomTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockChain_IsFirstBeaconHeightInEpoch(t *testing.T) {
	type fields struct {
		config Config
	}
	type args struct {
		beaconHeight uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "< break point 1",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 100,
			},
			want: false,
		},
		{
			name: "< break point 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 101,
			},
			want: true,
		},
		{
			name: "< break point 3",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 102,
			},
			want: false,
		},
		{
			name: "< break point 4",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 200,
			},
			want: false,
		},
		{
			name: "< break point 5",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 201,
			},
			want: true,
		},
		{
			name: "= break point 1",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 300,
			},
			want: false,
		},
		{
			name: "= break point 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 301,
			},
			want: true,
		},
		{
			name: "= break point 3",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 302,
			},
			want: false,
		},
		{
			name: "> break point",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 650,
			},
			want: false,
		},
		{
			name: "> break point 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 651,
			},
			want: true,
		},
		{
			name: "> break point 3",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 1000,
			},
			want: false,
		},
		{
			name: "> break point 4",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 1001,
			},
			want: true,
		},
		{
			name: "> break point 5",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 1002,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &BlockChain{
				config: tt.fields.config,
			}
			if got := bc.IsFirstBeaconHeightInEpoch(tt.args.beaconHeight); got != tt.want {
				t.Errorf("IsFirstBeaconHeightInEpoch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockChain_IsEqualToRandomTime(t *testing.T) {
	type fields struct {
		config Config
	}
	type args struct {
		beaconHeight uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "< break point 1",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 150,
			},
			want: true,
		},
		{
			name: "< break point 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 149,
			},
			want: false,
		},
		{
			name: "< break point 3",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 151,
			},
			want: false,
		},
		{
			name: "< break point 4",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 50,
			},
			want: true,
		},
		{
			name: "< break point 5",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 51,
			},
			want: false,
		},
		{
			name: "= break point 1",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 475,
			},
			want: true,
		},
		{
			name: "= break point 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 474,
			},
			want: false,
		},
		{
			name: "= break point 3",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 476,
			},
			want: false,
		},
		{
			name: "> break point",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 825,
			},
			want: true,
		},
		{
			name: "> break point 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 824,
			},
			want: false,
		},
		{
			name: "> break point 3",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
						RandomTime:        50,
						RandomTimeV2:      175,
					},
				},
			},
			args: args{
				beaconHeight: 826,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &BlockChain{
				config: tt.fields.config,
			}
			if got := bc.IsEqualToRandomTime(tt.args.beaconHeight); got != tt.want {
				t.Errorf("IsEqualToRandomTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockChain_IsLastBeaconHeightInEpoch(t *testing.T) {
	type fields struct {
		config Config
	}
	type args struct {
		beaconHeight uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "< break point 1",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 100,
			},
			want: true,
		},
		{
			name: "< break point 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 101,
			},
			want: false,
		},
		{
			name: "< break point 3",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 99,
			},
			want: false,
		},
		{
			name: "< break point 4",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 200,
			},
			want: true,
		},
		{
			name: "< break point 5",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 199,
			},
			want: false,
		},
		{
			name: "= break point 1",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 300,
			},
			want: true,
		},
		{
			name: "= break point 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 299,
			},
			want: false,
		},
		{
			name: "= break point 3",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 301,
			},
			want: false,
		},
		{
			name: "> break point",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 650,
			},
			want: true,
		},
		{
			name: "> break point 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 651,
			},
			want: false,
		},
		{
			name: "> break point 3",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 1000,
			},
			want: true,
		},
		{
			name: "> break point 4",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 999,
			},
			want: false,
		},
		{
			name: "> break point 5",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						Epoch:             100,
						EpochV2:           350,
						EpochV2BreakPoint: 4,
					},
				},
			},
			args: args{
				beaconHeight: 1001,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &BlockChain{
				config: tt.fields.config,
			}
			if got := bc.IsLastBeaconHeightInEpoch(tt.args.beaconHeight); got != tt.want {
				t.Errorf("IsLastBeaconHeightInEpoch() = %v, want %v", got, tt.want)
			}
		})
	}
}
