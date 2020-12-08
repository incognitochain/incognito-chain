package signaturecounter

import (
	"github.com/incognitochain/incognito-chain/incognitokey"
	"reflect"
	"sync"
	"testing"
)

var (
	committeePublicKeys = []string{
		"121VhftSAygpEJZ6i9jGkEKLMQTKTiiHzeUfeuhpQCcLZtys8FazpWwytpHebkAwgCxvqgUUF13fcSMtp5dgV1YkbRMj3z42TW2EebzAaiGg2DkGPodckN2UsbqhVDibpMgJUHVkLXardemfLdgUqWGtymdxaaRyPM38BAZcLpo2pAjxKv5vG5Uh9zHMkn7ZHtdNHmBmhG8B46UeiGBXYTwhyMe9KGS83jCMPAoUwHhTEXj5qQh6586dHjVxwEkRzp7SKn9iG1FFWdJ97xEkP2ezAapNQ46quVrMggcHFvoZofs1xdd4o5vAmPKnPTZtGTKunFiTWGnpSG9L6r5QpcmapqvRrK5SiuFhNM5DqgzUeHBb7fTfoiWd2N29jkbTGSq8CPUSjx3zdLR9sZguvPdnAA8g25cFPGSZt8aEnFJoPRzM",
		"121VhftSAygpEJZ6i9jGkEqPGAXcmKffwMbzpwxnEfzJxen4oZKPukWAUBbqvV5xPnowZ2eQmAj2mEebG2oexebQPh1MPFC6vEZAk6i7AiRPrZmfaRrRVrBp4WXnVJmL3xK4wzTfkR2rZkhUmSZm112TTyhDNkDQSaBGJkexrPbryqUygazCA2eyo6LnK5qs7jz2RhhsWqUTQ3sQJUuFcYdf2pSnYwhqZqphDCSRizDHeysaua5L7LwS8fY7KZHhPgTuFjvUWWnWSRTmV8u1dTY5kcmMdDZsPiyN9WfqjgVoTFNALjFG8U4GMvzV3kKwVVjuPMsM2XqyPDVpdNQUgLnv2bJS8Tr22A9NgF1FQfWyAny1DYyY3N5H3tfCggsybzZXzrbYPPgokvEynac91y8hPkRdgKW1e7FHzuBnEisPuKzy",
		"121VhftSAygpEJZ6i9jGkGLcYhJBeaJTGY5aFjqQA2WwyxU69Utrviuy9AJ3ATkeEyigVGScQUZw22cD1HeFKiyASYAs82WEamujt3nefYA9FPhURBpRTn6jDmGKUdb4QNbs7HVCJkRRaL9aktg1yaQaZE8TJFg2UeE9tBqUdmvD8fy36aDCYM5W86jaTVCXeEJQWPxUunP2EEL3e283PJ8zqPeBkpoFvkvhB28Hk3oRDeCCTC7QhbaV18ayKeToYqAxoUMBBihanfA33ixeX1daeKpajLCgDZ6jrfphwdYwQbf7dMcZ2NVvQ1a5JUCTJUZypwgKRt8tnTAKCowt2L1KNGP4NJJZm61cfHAGbKRyG9QxCJgK2SdMKsKPVefZSc9LbVaB7VeBby5LHxvMoCD7bN7g1HYRp4BX9n1fZJUeEkVa",
		"121VhftSAygpEJZ6i9jGkDjJj7e2cfgQvrLsPsmLhGMmGD9U9Knffa1MZAw79EijnpueVfTStN2VYt5jRqEr2DTjVqzUinwHVKWH4Tg4szHUntiBdWeqzNC4E8iiwC9Y2KtcRr3hBkpfqvyuBvchigatrigRvFVWu8H2RQqjvopLL51DQ4LFD87L9Zgj9HhasMeyr6f37yirs47JgtGs4BM7EhhpM5zD3TCsFabPphtwDKnfuLMaGzoAw5fM8zEXvdLMuohk96oayjdYothncdtZom17DxB1Mmw535eEjxBwz9ELoZRKk3LYiheSd4xGN9QsxrT2WnZCTd8B5QktARte5S91QYvRMixKC8UEuovQhXt8jMZNkq7CmMeXoybfYdmNaAHuqbY1QeUT2AgaqPho4ay3z5eeKRhnB28H18RGWQ1L",
	}
	committeePublicKeys2 = []string{
		"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",
	}
	samplePenaltyRule = []Penalty{
		{
			MinPercent:   50,
			Time:         0,
			ForceUnstake: true,
		},
	}
	committeePublicKeyStructs  = []incognitokey.CommitteePublicKey{}
	committeePublicKeyStructs2 = []incognitokey.CommitteePublicKey{}
)

var _ = func() (_ struct{}) {
	committeePublicKeyStructs, _ = incognitokey.CommitteeBase58KeyListToStruct(committeePublicKeys)
	committeePublicKeyStructs2, _ = incognitokey.CommitteeBase58KeyListToStruct(committeePublicKeys2)
	return
}()

func TestSignatureCounter_AddMissingSignature(t *testing.T) {
	missingSignatureFull := make(map[string]MissingSignature)
	aggregatedMissingSignatureFull := make(map[string]uint)
	for _, v := range committeePublicKeys {
		missingSignatureFull[v] = NewMissingSignature()
		aggregatedMissingSignatureFull[v] = 0
	}

	type fields struct {
		missingSignature           map[string]MissingSignature
		aggregatedMissingSignature map[string]MissingSignature
	}
	type args struct {
		data       string
		committees []incognitokey.CommitteePublicKey
	}

	tests := []struct {
		name       string
		fields     fields
		args       args
		wantFields fields
		wantErr    bool
	}{
		{
			name: "invalid input 1",
			fields: fields{
				missingSignature: make(map[string]MissingSignature),
			},
			args: args{
				data:       "{\"ProducerBLSSig\":\"1I6pNHXngYdBKspO08xZvk3fasdklaw;dkl;alwkd;lawkdl;kawl;dkkAaRQ9VpD+GhmwfT2b8p3PIYzouW4q/BFDxinllrIwUqq+XpugEiDjdmpfsHCAA=\",\"ProducerBriSig\":null,\"ValidatiorsIdx\":[0,1,2],\"AggSig\":\"LuFMS0uCziQOC/AL83xZb0Mortu+3lvx5mZ/kCtyJWE=\",\"BridgeSig\":[\"\",\"\",\"\"]}",
				committees: committeePublicKeyStructs,
			},
			wantFields: fields{
				missingSignature: map[string]MissingSignature{},
			},
			wantErr: true,
		},
		{
			name: "invalid input 2",
			fields: fields{
				missingSignature: make(map[string]MissingSignature),
			},
			args: args{
				data:       "{\"ProducerBLSSig\":\"1I6pNHXngYdBKspO08xZvk3fkAaRQ9VpD+GhmwfT2b8p3PIYzouW4q/BFDxinllrIwUqq+XpugEiDjdmpfsHCAA=\",\"ProducerBriSig\":null,\"ValidatiorsIdx\":[0,1,2],\"AggSig\":\"LuFMS0uCziQOC/AL83xZb0Mortuawdawdawd+3lvx5mZ/kCtyJWE=\",\"BridgeSig\":[\"\",\"\",\"\"]}",
				committees: committeePublicKeyStructs,
			},
			wantFields: fields{
				missingSignature: map[string]MissingSignature{},
			},
			wantErr: true,
		},
		{
			name: "valid input, committee slot 3 miss 1 signature",
			fields: fields{
				missingSignature: missingSignatureFull,
			},
			args: args{
				data:       "{\"ProducerBLSSig\":\"1I6pNHXngYdBKspO08xZvk3fkAaRQ9VpD+GhmwfT2b8p3PIYzouW4q/BFDxinllrIwUqq+XpugEiDjdmpfsHCAA=\",\"ProducerBriSig\":null,\"ValidatiorsIdx\":[0,1,2],\"AggSig\":\"LuFMS0uCziQOC/AL83xZb0Mortu+3lvx5mZ/kCtyJWE=\",\"BridgeSig\":[\"\",\"\",\"\"]}",
				committees: committeePublicKeyStructs,
			},
			wantFields: fields{
				missingSignature: map[string]MissingSignature{
					committeePublicKeys[0]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
					committeePublicKeys[1]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
					committeePublicKeys[2]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
					committeePublicKeys[3]: MissingSignature{
						Missing: 1,
						Total:   1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid input, committee slot 3 miss 1 signature",
			fields: fields{
				missingSignature: missingSignatureFull,
			},
			args: args{
				data:       "{\"ProducerBLSSig\":\"4lEXt6Z5RwRJmG7vK/6q2pLwGc0EcWi3Pw2D+rYvwBM/3YwgDjElAnH8Qb2OrAX4Lx3APk0Wo3oHYp1eO9hj7gA=\",\"ProducerBriSig\":null,\"ValidatiorsIdx\":[0,1,2],\"AggSig\":\"B93JfdZq3Q110tbR4fC7BWQim3NYICJRG/DZ3xlHw04=\",\"BridgeSig\":[\"\",\"\",\"\"]}",
				committees: committeePublicKeyStructs,
			},
			wantFields: fields{
				missingSignature: map[string]MissingSignature{
					committeePublicKeys[0]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
					committeePublicKeys[1]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
					committeePublicKeys[2]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
					committeePublicKeys[3]: MissingSignature{
						Missing: 1,
						Total:   1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid input, committee slot 2 miss 1 signature",
			fields: fields{
				missingSignature: missingSignatureFull,
			},
			args: args{
				data:       "{\"ProducerBLSSig\":\"LGcjV69UWOBv90wEVFgeq8pMNRWXaxqVPr82g1wqWA5XMmbdq7TZzECtPJl8pCkrSyzQnGVduAVaODGQrykTNQE=\",\"ProducerBriSig\":null,\"ValidatiorsIdx\":[0,1,3],\"AggSig\":\"Flod04E7A67JW4uPp43RGGLJR6j5ZnS8ZMrmz7MdE/A=\",\"BridgeSig\":[\"\",\"\",\"\"]}",
				committees: committeePublicKeyStructs,
			},
			wantFields: fields{
				missingSignature: map[string]MissingSignature{
					committeePublicKeys[0]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
					committeePublicKeys[1]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
					committeePublicKeys[2]: MissingSignature{
						Missing: 1,
						Total:   1,
					},
					committeePublicKeys[3]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid input, committee slot 1 miss 1 signature",
			fields: fields{
				missingSignature: missingSignatureFull,
			},
			args: args{
				data:       "{\"ProducerBLSSig\":\"HrpGEaXOUzydou9S9YE96OD48dSAtgI3zzIC2eisytQJJhtj0MgEwqU9MP1HswRk87NW3msE8w7Uyi7C+npWogA=\",\"ProducerBriSig\":null,\"ValidatiorsIdx\":[0,2,3],\"AggSig\":\"HTraoh3hx22W3iRl3SB9a7kv+p1N+ESGodAp28yjRDk=\",\"BridgeSig\":[\"\",\"\",\"\"]}",
				committees: committeePublicKeyStructs,
			},
			wantFields: fields{
				missingSignature: map[string]MissingSignature{
					committeePublicKeys[0]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
					committeePublicKeys[1]: MissingSignature{
						Missing: 1,
						Total:   1,
					},
					committeePublicKeys[2]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
					committeePublicKeys[3]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid input, committee slot 1 miss 1 signature",
			fields: fields{
				missingSignature: missingSignatureFull,
			},
			args: args{
				data:       "{\"ProducerBLSSig\":\"uSpMynim78XpsufwR6imkWcNKT6c5wwz4Nyb1GR+d3FplCfBwSQXNCd3bCgNGhieBuwGqSg5C5KG+zThOpY4rAA=\",\"ProducerBriSig\":null,\"ValidatiorsIdx\":[0,2,3],\"AggSig\":\"ocFaeoEmrzq0Ivg1N5gAvkuW4xsyDnC+NQiDUnYqQPE=\",\"BridgeSig\":[\"\",\"\",\"\"]}",
				committees: committeePublicKeyStructs,
			},
			wantFields: fields{
				missingSignature: map[string]MissingSignature{
					committeePublicKeys[0]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
					committeePublicKeys[1]: MissingSignature{
						Missing: 1,
						Total:   1,
					},
					committeePublicKeys[2]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
					committeePublicKeys[3]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid input, committee slot 1 miss 1 signature",
			fields: fields{
				missingSignature: missingSignatureFull,
			},
			args: args{
				data:       "{\"ProducerBLSSig\":\"5fp+nanu4VJoVIU5ZpA+uRASzkrjJgZMZ5eZOfYY5kwRWfnhWW4HlZhZdJ+dw2nzVzoR0KTyiG4Hno+TfMPvewE=\",\"ProducerBriSig\":null,\"ValidatiorsIdx\":[0,2,3],\"AggSig\":\"idOzTlb8oEoL6VsZ7UsQdPiFVf8HUX4Pad+8xxlE1/0=\",\"BridgeSig\":[\"\",\"\",\"\"]}",
				committees: committeePublicKeyStructs,
			},
			wantFields: fields{
				missingSignature: map[string]MissingSignature{
					committeePublicKeys[0]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
					committeePublicKeys[1]: MissingSignature{
						Missing: 1,
						Total:   1,
					},
					committeePublicKeys[2]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
					committeePublicKeys[3]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MissingSignatureCounter{
				missingSignature: tt.fields.missingSignature,
				lock:             new(sync.RWMutex),
			}
			if err := s.AddMissingSignature(tt.args.data, tt.args.committees); (err != nil) != tt.wantErr {
				t.Errorf("AddMissingSignature() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				if !reflect.DeepEqual(s.missingSignature, tt.wantFields.missingSignature) {
					t.Errorf("AddMissingSignature() missingSignature = got %v, want %v", s.missingSignature, tt.wantFields.missingSignature)
				}
			}
			for k, v := range s.missingSignature {
				if v.Missing != 0 {
					aggregatedMissingSignatureFull[k] += 1
				}
			}
			for k, _ := range missingSignatureFull {
				missingSignatureFull[k] = NewMissingSignature()
			}
		})
	}
	wantAggregatedMissingSignature := map[string]uint{
		committeePublicKeys[1]: 3,
		committeePublicKeys[2]: 1,
		committeePublicKeys[3]: 2,
	}
	for wantK, wantV := range wantAggregatedMissingSignature {
		if gotV, ok := aggregatedMissingSignatureFull[wantK]; !ok {
			t.Errorf("aggregatedMissingSignatureFull missingSignature NOT FOUND want %v ", wantK)
		} else {
			if wantV != gotV {
				t.Errorf("aggregatedMissingSignatureFull number of missingSignature got = %+v, want = %v ", gotV, wantV)
			}
		}
	}
}

func TestSignatureCounter_AddMissingSignature2(t *testing.T) {
	missingSignature1 := make(map[string]MissingSignature)
	missingSignature1[committeePublicKeys[0]] = NewMissingSignature()
	missingSignature1[committeePublicKeys[2]] = NewMissingSignature()

	missingSignature2 := make(map[string]MissingSignature)
	missingSignature2[committeePublicKeys[1]] = NewMissingSignature()
	missingSignature2[committeePublicKeys[0]] = NewMissingSignature()
	missingSignature2[committeePublicKeys[2]] = NewMissingSignature()

	type fields struct {
		missingSignature           map[string]MissingSignature
		aggregatedMissingSignature map[string]MissingSignature
	}
	type args struct {
		data       string
		committees []incognitokey.CommitteePublicKey
	}

	tests := []struct {
		name       string
		fields     fields
		args       args
		wantFields fields
		wantErr    bool
	}{
		{
			name: "valid input, committee slot 1 miss 1 signature",
			fields: fields{
				missingSignature: missingSignature1,
			},
			args: args{
				data:       "{\"ProducerBLSSig\":\"HrpGEaXOUzydou9S9YE96OD48dSAtgI3zzIC2eisytQJJhtj0MgEwqU9MP1HswRk87NW3msE8w7Uyi7C+npWogA=\",\"ProducerBriSig\":null,\"ValidatiorsIdx\":[0,2,3],\"AggSig\":\"HTraoh3hx22W3iRl3SB9a7kv+p1N+ESGodAp28yjRDk=\",\"BridgeSig\":[\"\",\"\",\"\"]}",
				committees: committeePublicKeyStructs,
			},
			wantFields: fields{
				missingSignature: map[string]MissingSignature{
					committeePublicKeys[0]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
					committeePublicKeys[2]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid input, committee slot 1 miss 1 signature",
			fields: fields{
				missingSignature: missingSignature2,
			},
			args: args{
				data:       "{\"ProducerBLSSig\":\"uSpMynim78XpsufwR6imkWcNKT6c5wwz4Nyb1GR+d3FplCfBwSQXNCd3bCgNGhieBuwGqSg5C5KG+zThOpY4rAA=\",\"ProducerBriSig\":null,\"ValidatiorsIdx\":[0,2,3],\"AggSig\":\"ocFaeoEmrzq0Ivg1N5gAvkuW4xsyDnC+NQiDUnYqQPE=\",\"BridgeSig\":[\"\",\"\",\"\"]}",
				committees: committeePublicKeyStructs,
			},
			wantFields: fields{
				missingSignature: map[string]MissingSignature{
					committeePublicKeys[0]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
					committeePublicKeys[1]: MissingSignature{
						Missing: 1,
						Total:   1,
					},
					committeePublicKeys[2]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MissingSignatureCounter{
				missingSignature: tt.fields.missingSignature,
				lock:             new(sync.RWMutex),
			}
			if err := s.AddMissingSignature(tt.args.data, tt.args.committees); (err != nil) != tt.wantErr {
				t.Errorf("AddMissingSignature() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				if !reflect.DeepEqual(s.missingSignature, tt.wantFields.missingSignature) {
					t.Errorf("AddMissingSignature() missingSignature = got %v, want %v", s.missingSignature, tt.wantFields.missingSignature)
				}
			}
		})
	}
}

func TestSignatureCounter_AddMissingSignature3(t *testing.T) {
	missingSignature1 := make(map[string]MissingSignature)
	for _, v := range committeePublicKeys2 {
		missingSignature1[v] = NewMissingSignature()
	}
	type fields struct {
		missingSignature           map[string]MissingSignature
		aggregatedMissingSignature map[string]MissingSignature
	}
	type args struct {
		data       string
		committees []incognitokey.CommitteePublicKey
	}

	tests := []struct {
		name       string
		fields     fields
		args       args
		wantFields fields
		wantErr    bool
	}{
		{
			name: "valid input, committee slot 4 miss 1 signature",
			fields: fields{
				missingSignature: missingSignature1,
			},
			args: args{
				data:       "{\"ProducerBLSSig\":\"G7p8f8VNypDpy36jEdY93DvwEHltwfgTH+mMig/mqOwHXUXHW+htI/ZMSUa9L7mIv50sKTm9Muw993KfC4fYpgE=\",\"ProducerBriSig\":null,\"ValidatiorsIdx\":[0,1,2,3,5],\"AggSig\":\"IgVZK8tjtIcz1LPcvHekkYzcHsuoFh+2OOOPr8m3ch4=\",\"BridgeSig\":[\"\",\"\",\"\",\"\",\"\"]}",
				committees: committeePublicKeyStructs2,
			},
			wantFields: fields{
				missingSignature: map[string]MissingSignature{
					committeePublicKeys2[0]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
					committeePublicKeys2[1]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
					committeePublicKeys2[2]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
					committeePublicKeys2[3]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
					committeePublicKeys2[4]: MissingSignature{
						Missing: 1,
						Total:   1,
					},
					committeePublicKeys2[5]: MissingSignature{
						Missing: 0,
						Total:   1,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MissingSignatureCounter{
				missingSignature: tt.fields.missingSignature,
				lock:             new(sync.RWMutex),
			}
			if err := s.AddMissingSignature(tt.args.data, tt.args.committees); (err != nil) != tt.wantErr {
				t.Errorf("AddMissingSignature() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				if !reflect.DeepEqual(s.missingSignature, tt.wantFields.missingSignature) {
					t.Errorf("AddMissingSignature() missingSignature = got %v, want %v", s.missingSignature, tt.wantFields.missingSignature)
				}
			}
		})
	}
}

func TestSignatureCounter_GetAllSlashingPenalty(t *testing.T) {
	type fields struct {
		missingSignature map[string]MissingSignature
		penalties        []Penalty
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]Penalty
	}{
		{
			name: "no penalty",
			fields: fields{
				missingSignature: map[string]MissingSignature{
					committeePublicKeys[0]: MissingSignature{
						Missing: 49,
						Total:   100,
					},
					committeePublicKeys[1]: MissingSignature{
						Missing: 149,
						Total:   300,
					},
					committeePublicKeys[2]: MissingSignature{
						Missing: 239,
						Total:   480,
					},
					committeePublicKeys[3]: MissingSignature{
						Missing: 249,
						Total:   500,
					},
				},
				penalties: samplePenaltyRule,
			},
			want: map[string]Penalty{},
		},
		{
			name: "penalty range >= 50",
			fields: fields{
				missingSignature: map[string]MissingSignature{
					committeePublicKeys[0]: MissingSignature{
						Missing: 51,
						Total:   100,
					},
					committeePublicKeys[1]: MissingSignature{
						Missing: 149,
						Total:   300,
					},
					committeePublicKeys[2]: MissingSignature{
						Missing: 239,
						Total:   480,
					},
					committeePublicKeys[3]: MissingSignature{
						Missing: 250,
						Total:   500,
					},
				},
				penalties: samplePenaltyRule,
			},
			want: map[string]Penalty{
				committeePublicKeys[0]: samplePenaltyRule[0],
				committeePublicKeys[3]: samplePenaltyRule[0],
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := MissingSignatureCounter{
				missingSignature: tt.fields.missingSignature,
				penalties:        tt.fields.penalties,
				lock:             new(sync.RWMutex),
			}
			if got := s.GetAllSlashingPenalty(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllSlashingPenalty() = %v, want %v", got, tt.want)
			}
		})
	}
}
