package blockchain

import (
	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/blockchain/committeestate/mocks"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"reflect"
	"testing"
)

var (
	keys = []string{
		"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",
	}
	incognitoKeys []incognitokey.CommitteePublicKey
)

var _ = func() (_ struct{}) {
	incognitoKeys, _ = incognitokey.CommitteeBase58KeyListToStruct(keys)
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func TestBeaconBestState_CalculateExpectedTotalBlock(t *testing.T) {
	type fields struct {
		NumberOfShardBlock map[byte]uint
	}
	tests := []struct {
		name   string
		fields fields
		want   map[byte]uint
	}{
		{
			name: "moderate different between shards",
			fields: fields{
				NumberOfShardBlock: map[byte]uint{
					0: 100,
					1: 150,
					2: 200,
					3: 200,
					4: 250,
					5: 300,
					6: 330,
					7: 340,
				},
			},
			want: map[byte]uint{
				0: 233,
				1: 233,
				2: 233,
				3: 233,
				4: 250,
				5: 300,
				6: 330,
				7: 340,
			},
		},
		{
			name: "big different between shards",
			fields: fields{
				NumberOfShardBlock: map[byte]uint{
					0: 10,
					1: 0,
					2: 20,
					3: 200,
					4: 250,
					5: 300,
					6: 330,
					7: 340,
				},
			},
			want: map[byte]uint{
				0: 181,
				1: 181,
				2: 181,
				3: 200,
				4: 250,
				5: 300,
				6: 330,
				7: 340,
			},
		},
		{
			name: "only one shard big different compare to shards",
			fields: fields{
				NumberOfShardBlock: map[byte]uint{
					0: 10,
					1: 300,
					2: 280,
					3: 290,
					4: 310,
					5: 300,
					6: 330,
					7: 340,
				},
			},
			want: map[byte]uint{
				0: 270,
				1: 300,
				2: 280,
				3: 290,
				4: 310,
				5: 300,
				6: 330,
				7: 340,
			},
		},
		{
			name: "0 all shard",
			fields: fields{
				NumberOfShardBlock: map[byte]uint{
					0: 0,
					1: 0,
					2: 0,
					3: 0,
					4: 0,
					5: 0,
					6: 0,
					7: 0,
				},
			},
			want: map[byte]uint{
				0: 0,
				1: 0,
				2: 0,
				3: 0,
				4: 0,
				5: 0,
				6: 0,
				7: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconBestState{
				NumberOfShardBlock: tt.fields.NumberOfShardBlock,
			}
			if got := b.CalculateExpectedTotalBlock(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CalculateExpectedTotalBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBeaconBestState_GetExpectedTotalBlock(t *testing.T) {
	type fields struct {
		NumberOfShardBlock    map[byte]uint
		beaconCommitteeEngine committeestate.BeaconCommitteeEngine
	}

	mockCommittee := map[byte][]incognitokey.CommitteePublicKey{
		0: incognitoKeys[0:2],
		1: incognitoKeys[2:4],
		2: incognitoKeys[4:6],
		3: incognitoKeys[6:8],
		4: incognitoKeys[8:10],
		5: incognitoKeys[10:12],
		6: incognitoKeys[12:14],
		7: incognitoKeys[14:16],
	}

	mockEngine1 := &mocks.BeaconCommitteeEngine{}
	mockEngine1.On("GetShardCommittee").Return(mockCommittee).Times(4)

	tests := []struct {
		name   string
		fields fields
		want   map[string]uint
	}{
		{
			name: "moderate different between shards",
			fields: fields{
				NumberOfShardBlock: map[byte]uint{
					0: 100,
					1: 150,
					2: 200,
					3: 200,
					4: 250,
					5: 300,
					6: 330,
					7: 340,
				},
				beaconCommitteeEngine: mockEngine1,
			},
			want: map[string]uint{
				keys[0]:  233,
				keys[1]:  233,
				keys[2]:  233,
				keys[3]:  233,
				keys[4]:  233,
				keys[5]:  233,
				keys[6]:  233,
				keys[7]:  233,
				keys[8]:  250,
				keys[9]:  250,
				keys[10]: 300,
				keys[11]: 300,
				keys[12]: 330,
				keys[13]: 330,
				keys[14]: 340,
				keys[15]: 340,
			},
		},
		{
			name: "big different between shards",
			fields: fields{
				NumberOfShardBlock: map[byte]uint{
					0: 10,
					1: 0,
					2: 20,
					3: 200,
					4: 250,
					5: 300,
					6: 330,
					7: 340,
				},
				beaconCommitteeEngine: mockEngine1,
			},
			want: map[string]uint{
				keys[0]:  181,
				keys[1]:  181,
				keys[2]:  181,
				keys[3]:  181,
				keys[4]:  181,
				keys[5]:  181,
				keys[6]:  200,
				keys[7]:  200,
				keys[8]:  250,
				keys[9]:  250,
				keys[10]: 300,
				keys[11]: 300,
				keys[12]: 330,
				keys[13]: 330,
				keys[14]: 340,
				keys[15]: 340,
			},
		},
		{
			name: "only one shard big different compare to shards",
			fields: fields{
				NumberOfShardBlock: map[byte]uint{
					0: 10,
					1: 300,
					2: 280,
					3: 290,
					4: 310,
					5: 300,
					6: 330,
					7: 340,
				},
				beaconCommitteeEngine: mockEngine1,
			},
			want: map[string]uint{
				keys[0]:  270,
				keys[1]:  270,
				keys[2]:  300,
				keys[3]:  300,
				keys[4]:  280,
				keys[5]:  280,
				keys[6]:  290,
				keys[7]:  290,
				keys[8]:  310,
				keys[9]:  310,
				keys[10]: 300,
				keys[11]: 300,
				keys[12]: 330,
				keys[13]: 330,
				keys[14]: 340,
				keys[15]: 340,
			},
		},
		{
			name: "0 all shard",
			fields: fields{
				NumberOfShardBlock: map[byte]uint{
					0: 0,
					1: 0,
					2: 0,
					3: 0,
					4: 0,
					5: 0,
					6: 0,
					7: 0,
				},
				beaconCommitteeEngine: mockEngine1,
			},
			want: map[string]uint{
				keys[0]:  0,
				keys[1]:  0,
				keys[2]:  0,
				keys[3]:  0,
				keys[4]:  0,
				keys[5]:  0,
				keys[6]:  0,
				keys[7]:  0,
				keys[8]:  0,
				keys[9]:  0,
				keys[10]: 0,
				keys[11]: 0,
				keys[12]: 0,
				keys[13]: 0,
				keys[14]: 0,
				keys[15]: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconBestState{
				NumberOfShardBlock:    tt.fields.NumberOfShardBlock,
				beaconCommitteeEngine: tt.fields.beaconCommitteeEngine,
			}
			if got := b.GetExpectedTotalBlock(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetExpectedTotalBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}
