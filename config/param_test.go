package config

import "testing"

func Test_verifyParam(t *testing.T) {
	type args struct {
		p *param
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "shard committee size",
			args: args{
				p: &param{
					CommitteeSize: committeeSize{
						MaxShardCommitteeSize: 5,
						MinShardCommitteeSize: 6,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "beacon committee size",
			args: args{
				p: &param{
					CommitteeSize: committeeSize{
						MaxShardCommitteeSize:  6,
						MinShardCommitteeSize:  5,
						MaxBeaconCommitteeSize: 5,
						MinBeaconCommitteeSize: 6,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "min committee size and fixed validator",
			args: args{
				p: &param{
					CommitteeSize: committeeSize{
						MaxShardCommitteeSize:            6,
						MinShardCommitteeSize:            5,
						MaxBeaconCommitteeSize:           6,
						MinBeaconCommitteeSize:           5,
						NumberOfFixedShardBlockValidator: 6,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "min shard committee size invalid",
			args: args{
				p: &param{
					CommitteeSize: committeeSize{
						MaxShardCommitteeSize:            6,
						MinShardCommitteeSize:            3,
						MaxBeaconCommitteeSize:           6,
						MinBeaconCommitteeSize:           5,
						NumberOfFixedShardBlockValidator: 3,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "min beacon committee size invalid",
			args: args{
				p: &param{
					CommitteeSize: committeeSize{
						MaxShardCommitteeSize:            6,
						MinShardCommitteeSize:            5,
						MaxBeaconCommitteeSize:           6,
						MinBeaconCommitteeSize:           3,
						NumberOfFixedShardBlockValidator: 3,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid init committee size",
			args: args{
				p: &param{
					CommitteeSize: committeeSize{
						MaxShardCommitteeSize:            6,
						MinShardCommitteeSize:            5,
						MaxBeaconCommitteeSize:           6,
						MinBeaconCommitteeSize:           5,
						NumberOfFixedShardBlockValidator: 5,
						InitShardCommitteeSize:           4,
						ShardCommitteeSizeKeyListV2:      3,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid init committee size",
			args: args{
				p: &param{
					CommitteeSize: committeeSize{
						MaxShardCommitteeSize:            6,
						MinShardCommitteeSize:            5,
						MaxBeaconCommitteeSize:           6,
						MinBeaconCommitteeSize:           5,
						NumberOfFixedShardBlockValidator: 5,
						InitShardCommitteeSize:           4,
						ShardCommitteeSizeKeyListV2:      4,
						InitBeaconCommitteeSize:          4,
						BeaconCommitteeSizeKeyListV2:     3,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid random time and epoch",
			args: args{
				p: &param{
					CommitteeSize: committeeSize{
						MaxShardCommitteeSize:            6,
						MinShardCommitteeSize:            5,
						MaxBeaconCommitteeSize:           6,
						MinBeaconCommitteeSize:           5,
						NumberOfFixedShardBlockValidator: 5,
						InitShardCommitteeSize:           4,
						ShardCommitteeSizeKeyListV2:      4,
						InitBeaconCommitteeSize:          4,
						BeaconCommitteeSizeKeyListV2:     4,
					},
					EpochParam: epochParam{
						RandomTime:           100,
						NumberOfBlockInEpoch: 100,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid beacon creation",
			args: args{
				p: &param{
					CommitteeSize: committeeSize{
						MaxShardCommitteeSize:            6,
						MinShardCommitteeSize:            5,
						MaxBeaconCommitteeSize:           6,
						MinBeaconCommitteeSize:           5,
						NumberOfFixedShardBlockValidator: 5,
						InitShardCommitteeSize:           4,
						ShardCommitteeSizeKeyListV2:      4,
						InitBeaconCommitteeSize:          4,
						BeaconCommitteeSizeKeyListV2:     4,
					},
					EpochParam: epochParam{
						RandomTime:           99,
						NumberOfBlockInEpoch: 100,
					},
					BlockTime: blockTime{
						MaxBeaconBlockCreation: 10,
						MinBeaconBlockInterval: 11,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid shard creation",
			args: args{
				p: &param{
					CommitteeSize: committeeSize{
						MaxShardCommitteeSize:            6,
						MinShardCommitteeSize:            5,
						MaxBeaconCommitteeSize:           6,
						MinBeaconCommitteeSize:           5,
						NumberOfFixedShardBlockValidator: 5,
						InitShardCommitteeSize:           4,
						ShardCommitteeSizeKeyListV2:      4,
						InitBeaconCommitteeSize:          4,
						BeaconCommitteeSizeKeyListV2:     4,
					},
					EpochParam: epochParam{
						RandomTime:           99,
						NumberOfBlockInEpoch: 100,
					},
					BlockTime: blockTime{
						MaxBeaconBlockCreation: 10,
						MinBeaconBlockInterval: 10,
						MinShardBlockInterval:  10,
						MaxShardBlockCreation:  9,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "pass all",
			args: args{
				p: &param{
					CommitteeSize: committeeSize{
						MaxShardCommitteeSize:            6,
						MinShardCommitteeSize:            5,
						MaxBeaconCommitteeSize:           6,
						MinBeaconCommitteeSize:           5,
						NumberOfFixedShardBlockValidator: 5,
						InitShardCommitteeSize:           4,
						ShardCommitteeSizeKeyListV2:      4,
						InitBeaconCommitteeSize:          4,
						BeaconCommitteeSizeKeyListV2:     4,
					},
					EpochParam: epochParam{
						RandomTime:           99,
						NumberOfBlockInEpoch: 100,
					},
					BlockTime: blockTime{
						MaxBeaconBlockCreation: 10,
						MinBeaconBlockInterval: 10,
						MinShardBlockInterval:  10,
						MaxShardBlockCreation:  10,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := verifyParam(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("verifyParam() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
