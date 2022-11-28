package signaturecounter

import (
	"reflect"
	"sync"
	"testing"

	"github.com/incognitochain/incognito-chain/incognitokey"
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
			MinPercent: 50,
			// Time:         0,
			ForceUnstake: true,
		},
	}
	committeePublicKeyStructs  = []incognitokey.CommitteePublicKey{}
	committeePublicKeyStructs2 = []incognitokey.CommitteePublicKey{}

	keys = []string{
		"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",
	}
	incognitoKeys []incognitokey.CommitteePublicKey
)

var _ = func() (_ struct{}) {
	incognitoKeys, _ = incognitokey.CommitteeBase58KeyListToStruct(keys)
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
						Missing:     0,
						ActualTotal: 1,
					},
					committeePublicKeys[1]: MissingSignature{
						Missing:     0,
						ActualTotal: 1,
					},
					committeePublicKeys[2]: MissingSignature{
						Missing:     0,
						ActualTotal: 1,
					},
					committeePublicKeys[3]: MissingSignature{
						Missing:     1,
						ActualTotal: 1,
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
						Missing:     0,
						ActualTotal: 1,
					},
					committeePublicKeys[1]: MissingSignature{
						Missing:     0,
						ActualTotal: 1,
					},
					committeePublicKeys[2]: MissingSignature{
						Missing:     0,
						ActualTotal: 1,
					},
					committeePublicKeys[3]: MissingSignature{
						Missing:     1,
						ActualTotal: 1,
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
						Missing:     0,
						ActualTotal: 1,
					},
					committeePublicKeys[1]: MissingSignature{
						Missing:     0,
						ActualTotal: 1,
					},
					committeePublicKeys[2]: MissingSignature{
						Missing:     1,
						ActualTotal: 1,
					},
					committeePublicKeys[3]: MissingSignature{
						Missing:     0,
						ActualTotal: 1,
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
						Missing:     0,
						ActualTotal: 1,
					},
					committeePublicKeys[1]: MissingSignature{
						Missing:     1,
						ActualTotal: 1,
					},
					committeePublicKeys[2]: MissingSignature{
						Missing:     0,
						ActualTotal: 1,
					},
					committeePublicKeys[3]: MissingSignature{
						Missing:     0,
						ActualTotal: 1,
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
						Missing:     0,
						ActualTotal: 1,
					},
					committeePublicKeys[1]: MissingSignature{
						Missing:     1,
						ActualTotal: 1,
					},
					committeePublicKeys[2]: MissingSignature{
						Missing:     0,
						ActualTotal: 1,
					},
					committeePublicKeys[3]: MissingSignature{
						Missing:     0,
						ActualTotal: 1,
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
						Missing:     0,
						ActualTotal: 1,
					},
					committeePublicKeys[1]: MissingSignature{
						Missing:     1,
						ActualTotal: 1,
					},
					committeePublicKeys[2]: MissingSignature{
						Missing:     0,
						ActualTotal: 1,
					},
					committeePublicKeys[3]: MissingSignature{
						Missing:     0,
						ActualTotal: 1,
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
			if err := s.AddMissingSignature(tt.args.data, 5, tt.args.committees); (err != nil) != tt.wantErr {
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
						Missing:     0,
						ActualTotal: 1,
					},
					committeePublicKeys[2]: MissingSignature{
						Missing:     0,
						ActualTotal: 1,
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
						Missing:     0,
						ActualTotal: 1,
					},
					committeePublicKeys[1]: MissingSignature{
						Missing:     1,
						ActualTotal: 1,
					},
					committeePublicKeys[2]: MissingSignature{
						Missing:     0,
						ActualTotal: 1,
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
			if err := s.AddMissingSignature(tt.args.data, 5, tt.args.committees); (err != nil) != tt.wantErr {
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
						Missing:     0,
						ActualTotal: 1,
					},
					committeePublicKeys2[1]: MissingSignature{
						Missing:     0,
						ActualTotal: 1,
					},
					committeePublicKeys2[2]: MissingSignature{
						Missing:     0,
						ActualTotal: 1,
					},
					committeePublicKeys2[3]: MissingSignature{
						Missing:     0,
						ActualTotal: 1,
					},
					committeePublicKeys2[4]: MissingSignature{
						Missing:     1,
						ActualTotal: 1,
					},
					committeePublicKeys2[5]: MissingSignature{
						Missing:     0,
						ActualTotal: 1,
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
			if err := s.AddMissingSignature(tt.args.data, 5, tt.args.committees); (err != nil) != tt.wantErr {
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
						Missing:     49,
						ActualTotal: 100,
					},
					committeePublicKeys[1]: MissingSignature{
						Missing:     149,
						ActualTotal: 300,
					},
					committeePublicKeys[2]: MissingSignature{
						Missing:     239,
						ActualTotal: 480,
					},
					committeePublicKeys[3]: MissingSignature{
						Missing:     249,
						ActualTotal: 500,
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
						Missing:     51,
						ActualTotal: 100,
					},
					committeePublicKeys[1]: MissingSignature{
						Missing:     149,
						ActualTotal: 300,
					},
					committeePublicKeys[2]: MissingSignature{
						Missing:     239,
						ActualTotal: 480,
					},
					committeePublicKeys[3]: MissingSignature{
						Missing:     250,
						ActualTotal: 500,
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
			if got := s.GetAllSlashingPenaltyWithActualTotalBlock(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllSlashingPenaltyWithActualTotalBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMissingSignatureCounter_GetAllSlashingPenaltyWithExpectedTotalBlock(t *testing.T) {
	type fields struct {
		missingSignature map[string]MissingSignature
		penalties        []Penalty
		lock             *sync.RWMutex
	}
	type args struct {
		expectedTotalBlocks map[string]uint
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]Penalty
	}{
		{
			name: "no penalty",
			fields: fields{
				missingSignature: map[string]MissingSignature{
					keys[0]: {
						Missing:     0,
						ActualTotal: 100,
					},
					keys[1]: {
						Missing:     49,
						ActualTotal: 100,
					},
					keys[2]: {
						Missing:     149,
						ActualTotal: 300,
					},
					keys[3]: {
						Missing:     99,
						ActualTotal: 200,
					},
				},
				penalties: samplePenaltyRule,
				lock:      new(sync.RWMutex),
			},
			args: args{
				expectedTotalBlocks: map[string]uint{
					keys[0]: 100,
					keys[1]: 100,
					keys[2]: 300,
					keys[3]: 200,
				},
			},
			want: map[string]Penalty{},
		},
		{
			name: "not in list",
			fields: fields{
				missingSignature: map[string]MissingSignature{
					keys[0]: {
						Missing:     0,
						ActualTotal: 100,
					},
					keys[1]: {
						Missing:     49,
						ActualTotal: 100,
					},
					keys[2]: {
						Missing:     149,
						ActualTotal: 300,
					},
					keys[3]: {
						Missing:     99,
						ActualTotal: 200,
					},
					keys[4]: {
						Missing: 99,
					},
				},
				penalties: samplePenaltyRule,
				lock:      new(sync.RWMutex),
			},
			args: args{
				expectedTotalBlocks: map[string]uint{
					keys[0]: 100,
					keys[1]: 100,
					keys[2]: 300,
					keys[3]: 200,
				},
			},
			want: map[string]Penalty{},
		},
		{
			name: "not record missing signature",
			fields: fields{
				missingSignature: map[string]MissingSignature{
					keys[0]: {
						Missing:     0,
						ActualTotal: 100,
					},
					keys[1]: {
						Missing:     49,
						ActualTotal: 100,
					},
					keys[2]: {
						Missing:     149,
						ActualTotal: 300,
					},
					keys[3]: {
						Missing:     99,
						ActualTotal: 200,
					},
					keys[4]: {
						Missing: 99,
					},
				},
				penalties: samplePenaltyRule,
				lock:      new(sync.RWMutex),
			},
			args: args{
				expectedTotalBlocks: map[string]uint{
					keys[0]: 100,
					keys[1]: 100,
					keys[2]: 300,
					keys[3]: 200,
					keys[5]: 200,
				},
			},
			want: map[string]Penalty{
				keys[5]: samplePenaltyRule[0],
			},
		},
		{
			name: "missing more signature than over expected",
			fields: fields{
				missingSignature: map[string]MissingSignature{
					keys[0]: {
						Missing:     0,
						ActualTotal: 100,
					},
					keys[1]: {
						Missing:     49,
						ActualTotal: 100,
					},
					keys[2]: {
						Missing:     149,
						ActualTotal: 300,
					},
					keys[3]: {
						Missing:     99,
						ActualTotal: 200,
					},
					keys[4]: {
						Missing:     99,
						ActualTotal: 100,
					},
					keys[6]: {
						Missing:     0,
						ActualTotal: 200,
					},
				},
				penalties: samplePenaltyRule,
				lock:      new(sync.RWMutex),
			},
			args: args{
				expectedTotalBlocks: map[string]uint{
					keys[0]: 100,
					keys[1]: 100,
					keys[2]: 300,
					keys[3]: 200,
					keys[5]: 200,
					keys[6]: 180,
				},
			},
			want: map[string]Penalty{
				keys[5]: samplePenaltyRule[0],
			},
		},
		{
			name: "missing more signature more than 50% of expected",
			fields: fields{
				missingSignature: map[string]MissingSignature{
					keys[0]: {
						Missing:     0,
						ActualTotal: 100,
					},
					keys[1]: {
						Missing:     49,
						ActualTotal: 100,
					},
					keys[2]: {
						Missing:     149,
						ActualTotal: 300,
					},
					keys[3]: {
						Missing:     99,
						ActualTotal: 200,
					},
					keys[4]: {
						Missing: 99,
					},
					keys[6]: {
						Missing:     0,
						ActualTotal: 200,
					},
					keys[7]: {
						Missing:     100,
						ActualTotal: 200,
					},
				},
				penalties: samplePenaltyRule,
				lock:      new(sync.RWMutex),
			},
			args: args{
				expectedTotalBlocks: map[string]uint{
					keys[0]: 100,
					keys[1]: 100,
					keys[2]: 300,
					keys[3]: 200,
					keys[5]: 200,
					keys[6]: 180,
					keys[7]: 200,
				},
			},
			want: map[string]Penalty{
				keys[5]: samplePenaltyRule[0],
				keys[7]: samplePenaltyRule[0],
			},
		},
		{
			name: "produce no block in an epoch",
			fields: fields{
				missingSignature: map[string]MissingSignature{
					keys[0]: {
						Missing:     0,
						ActualTotal: 100,
					},
					keys[1]: {
						Missing:     49,
						ActualTotal: 100,
					},
					keys[2]: {
						Missing:     149,
						ActualTotal: 300,
					},
					keys[3]: {
						Missing:     99,
						ActualTotal: 200,
					},
					keys[4]: {
						Missing: 99,
					},
					keys[6]: {
						Missing:     0,
						ActualTotal: 200,
					},
					keys[7]: {
						Missing:     100,
						ActualTotal: 200,
					},
					keys[8]: {
						Missing:     0,
						ActualTotal: 0,
					},
				},
				penalties: samplePenaltyRule,
				lock:      new(sync.RWMutex),
			},
			args: args{
				expectedTotalBlocks: map[string]uint{
					keys[0]: 100,
					keys[1]: 100,
					keys[2]: 300,
					keys[3]: 200,
					keys[5]: 200,
					keys[6]: 180,
					keys[7]: 200,
					keys[8]: 100,
				},
			},
			want: map[string]Penalty{
				keys[5]: samplePenaltyRule[0],
				keys[7]: samplePenaltyRule[0],
				keys[8]: samplePenaltyRule[0],
			},
		},
		{
			name: "mix cases",
			fields: fields{
				missingSignature: map[string]MissingSignature{
					keys[0]: {
						Missing:     0,
						ActualTotal: 100,
					},
					keys[1]: {
						Missing:     49,
						ActualTotal: 100,
					},
					keys[2]: {
						Missing:     149,
						ActualTotal: 300,
					},
					keys[3]: {
						Missing:     99,
						ActualTotal: 200,
					},
					keys[4]: {
						Missing: 99,
					},
					keys[6]: {
						Missing:     0,
						ActualTotal: 200,
					},
					keys[7]: {
						Missing:     100,
						ActualTotal: 200,
					},
					keys[8]: {
						Missing:     0,
						ActualTotal: 0,
					},
					keys[9]: {
						Missing:     15,
						ActualTotal: 30,
					},
					keys[10]: {
						Missing:     0,
						ActualTotal: 30,
					},
					keys[11]: {
						Missing:     0,
						ActualTotal: 101,
					},
				},
				penalties: samplePenaltyRule,
				lock:      new(sync.RWMutex),
			},
			args: args{
				expectedTotalBlocks: map[string]uint{
					keys[0]:  100,
					keys[1]:  100,
					keys[2]:  300,
					keys[3]:  200,
					keys[5]:  200,
					keys[6]:  180,
					keys[7]:  200,
					keys[8]:  100,
					keys[9]:  100,
					keys[10]: 100,
					keys[11]: 100,
				},
			},
			want: map[string]Penalty{
				keys[5]:  samplePenaltyRule[0],
				keys[7]:  samplePenaltyRule[0],
				keys[8]:  samplePenaltyRule[0],
				keys[9]:  samplePenaltyRule[0],
				keys[10]: samplePenaltyRule[0],
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := MissingSignatureCounter{
				missingSignature: tt.fields.missingSignature,
				penalties:        tt.fields.penalties,
				lock:             tt.fields.lock,
			}
			if got := s.GetAllSlashingPenaltyWithExpectedTotalBlock(tt.args.expectedTotalBlocks); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllSlashingPenaltyWithExpectedTotalBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}
