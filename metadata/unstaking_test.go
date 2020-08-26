package metadata_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/metadata/mocks"
)

const (
	SPLITTER = ","
)

var key1 = "121VhftSAygpEJZ6i9jGkEKLMQTKTiiHzeUfeuhpQCcLZtys8FazpWwytpHebkAwgCxvqgUUF13fcSMtp5dgV1YkbRMj3z42TW2EebzAaiGg2DkGPodckN2UsbqhVDibpMgJUHVkLXardemfLdgUqWGtymdxaaRyPM38BAZcLpo2pAjxKv5vG5Uh9zHMkn7ZHtdNHmBmhG8B46UeiGBXYTwhyMe9KGS83jCMPAoUwHhTEXj5qQh6586dHjVxwEkRzp7SKn9iG1FFWdJ97xEkP2ezAapNQ46quVrMggcHFvoZofs1xdd4o5vAmPKnPTZtGTKunFiTWGnpSG9L6r5QpcmapqvRrK5SiuFhNM5DqgzUeHBb7fTfoiWd2N29jkbTGSq8CPUSjx3zdLR9sZguvPdnAA8g25cFPGSZt8aEnFJoPRzM"
var key2 = "121VhftSAygpEJZ6i9jGkEqPGAXcmKffwMbzpwxnEfzJxen4oZKPukWAUBbqvV5xPnowZ2eQmAj2mEebG2oexebQPh1MPFC6vEZAk6i7AiRPrZmfaRrRVrBp4WXnVJmL3xK4wzTfkR2rZkhUmSZm112TTyhDNkDQSaBGJkexrPbryqUygazCA2eyo6LnK5qs7jz2RhhsWqUTQ3sQJUuFcYdf2pSnYwhqZqphDCSRizDHeysaua5L7LwS8fY7KZHhPgTuFjvUWWnWSRTmV8u1dTY5kcmMdDZsPiyN9WfqjgVoTFNALjFG8U4GMvzV3kKwVVjuPMsM2XqyPDVpdNQUgLnv2bJS8Tr22A9NgF1FQfWyAny1DYyY3N5H3tfCggsybzZXzrbYPPgokvEynac91y8hPkRdgKW1e7FHzuBnEisPuKzy"
var key3 = "121VhftSAygpEJZ6i9jGkGLcYhJBeaJTGY5aFjqQA2WwyxU69Utrviuy9AJ3ATkeEyigVGScQUZw22cD1HeFKiyASYAs82WEamujt3nefYA9FPhURBpRTn6jDmGKUdb4QNbs7HVCJkRRaL9aktg1yaQaZE8TJFg2UeE9tBqUdmvD8fy36aDCYM5W86jaTVCXeEJQWPxUunP2EEL3e283PJ8zqPeBkpoFvkvhB28Hk3oRDeCCTC7QhbaV18ayKeToYqAxoUMBBihanfA33ixeX1daeKpajLCgDZ6jrfphwdYwQbf7dMcZ2NVvQ1a5JUCTJUZypwgKRt8tnTAKCowt2L1KNGP4NJJZm61cfHAGbKRyG9QxCJgK2SdMKsKPVefZSc9LbVaB7VeBby5LHxvMoCD7bN7g1HYRp4BX9n1fZJUeEkVa"
var key4 = "121VhftSAygpEJZ6i9jGkDjJj7e2cfgQvrLsPsmLhGMmGD9U9Knffa1MZAw79EijnpueVfTStN2VYt5jRqEr2DTjVqzUinwHVKWH4Tg4szHUntiBdWeqzNC4E8iiwC9Y2KtcRr3hBkpfqvyuBvchigatrigRvFVWu8H2RQqjvopLL51DQ4LFD87L9Zgj9HhasMeyr6f37yirs47JgtGs4BM7EhhpM5zD3TCsFabPphtwDKnfuLMaGzoAw5fM8zEXvdLMuohk96oayjdYothncdtZom17DxB1Mmw535eEjxBwz9ELoZRKk3LYiheSd4xGN9QsxrT2WnZCTd8B5QktARte5S91QYvRMixKC8UEuovQhXt8jMZNkq7CmMeXoybfYdmNaAHuqbY1QeUT2AgaqPho4ay3z5eeKRhnB28H18RGWQ1L"

var incKey1, incKey2, incKey3, incKey4 *incognitokey.CommitteePublicKey

//initPublicKey init incognito public key for testing by base 58 string
func initPublicKey() {
	incKey1 = new(incognitokey.CommitteePublicKey)
	incKey2 = new(incognitokey.CommitteePublicKey)
	incKey3 = new(incognitokey.CommitteePublicKey)
	incKey4 = new(incognitokey.CommitteePublicKey)

	err := incKey1.FromBase58(key1)
	if err != nil {
		panic(err)
	}

	err = incKey2.FromBase58(key2)
	if err != nil {
		panic(err)
	}

	err = incKey3.FromBase58(key3)
	if err != nil {
		panic(err)
	}

	err = incKey4.FromBase58(key4)
	if err != nil {
		panic(err)
	}
}

func TestUnStakingMetadata_ValidateMetadataByItself(t *testing.T) {

	initPublicKey()

	type fields struct {
		MetadataBase       metadata.MetadataBase
		CommitteePublicKey string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "Wrong metadata type",
			fields: fields{
				MetadataBase: metadata.MetadataBase{
					Type: metadata.BeaconStakingMeta,
				},
				CommitteePublicKey: strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
			},
			want: false,
		},
		{
			name: "Invalid Format of Public Keys",
			fields: fields{
				MetadataBase: metadata.MetadataBase{
					Type: metadata.UnStakingMeta,
				},
				CommitteePublicKey: "123",
			},
			want: false,
		},
		{
			name: "Valid Input",
			fields: fields{
				MetadataBase: metadata.MetadataBase{
					Type: metadata.UnStakingMeta,
				},
				CommitteePublicKey: key1,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unStakingMetadata := &metadata.UnStakingMetadata{
				MetadataBase:       tt.fields.MetadataBase,
				CommitteePublicKey: tt.fields.CommitteePublicKey,
			}
			if got := unStakingMetadata.ValidateMetadataByItself(); got != tt.want {
				t.Errorf("UnStakingMetadata.ValidateMetadataByItself() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnStakingMetadata_ValidateTxWithBlockChain(t *testing.T) {

	initPublicKey()

	subtitutePublicKeys := []string{key1}

	unstakingMetaData := &metadata.UnStakingMetadata{
		MetadataBase: metadata.MetadataBase{
			Type: metadata.UnStakingMeta,
		},
	}

	stopAutoStakinggMetaData := &metadata.StopAutoStakingMetadata{
		MetadataBase: metadata.MetadataBase{
			Type: metadata.StopAutoStakingMeta,
		},
	}

	unstakingTx := &mocks.Transaction{}
	unstakingTx.On("GetMetadata").Return(unstakingMetaData)
	stopAutoStakingTx := &mocks.Transaction{}
	stopAutoStakingTx.On("GetMetadata").Return(stopAutoStakinggMetaData)

	beaconViewSubtituteCommitteesError := &mocks.BeaconViewRetriever{}
	beaconViewSubtituteCommitteesError.
		On("GetAllCommitteeValidatorCandidateFlattenListFromDatabase").
		Return([]string{}, errors.New("Can't find public key in list subtitute committees"))

	beaconViewNotFoundSubtituteCommittees := &mocks.BeaconViewRetriever{}
	beaconViewNotFoundSubtituteCommittees.
		On("GetAllCommitteeValidatorCandidateFlattenListFromDatabase").
		Return([]string{}, nil)

	beaconViewGetStakerInfoError := &mocks.BeaconViewRetriever{}
	beaconViewGetStakerInfoError.
		On("GetAllCommitteeValidatorCandidateFlattenListFromDatabase").
		Return(subtitutePublicKeys, nil)
	beaconViewGetStakerInfoError.
		On("GetStakerInfo", key1).
		Return(nil, false, errors.New("Can't Get Staker Info From Beacon Chain Database"))

	beaconViewNotFoundStakerInfo := &mocks.BeaconViewRetriever{}
	beaconViewNotFoundStakerInfo.
		On("GetAllCommitteeValidatorCandidateFlattenListFromDatabase").
		Return(subtitutePublicKeys, nil)
	beaconViewNotFoundStakerInfo.
		On("GetStakerInfo", key1).
		Return(nil, false, nil)

	supportShardView := &mocks.ShardViewRetriever{}
	supportShardView.
		On("GetShardID").
		Return(0)
	supportShardView.
		On("GetBeaconHeight").
		Return(100)

	shardViewGetStakingTxError := supportShardView
	shardViewGetStakingTxError.
		On("GetShardStakingTx", 0, 100).
		Return(nil, errors.New("Get Shard Staking Error"))

	stakingTx := map[string]string{
		key1: "12",
	}

	// stakingTxError := map[string]

	shardViewStakingTx := supportShardView
	shardViewStakingTx.
		On("GetShardStakingTx", 0, 100).
		Return(stakingTx, nil)

	beaconViewValidInput := &mocks.BeaconViewRetriever{}
	beaconViewValidInput.
		On("GetAllCommitteeValidatorCandidateFlattenListFromDatabase").
		Return(subtitutePublicKeys, nil)
	beaconViewValidInput.
		On("GetStakerInfo", key1).
		Return(nil, true, nil)

	type fields struct {
		MetadataBase       metadata.MetadataBase
		CommitteePublicKey string
	}
	type args struct {
		tx                  metadata.Transaction
		chainRetriever      metadata.ChainRetriever
		shardViewRetriever  metadata.ShardViewRetriever
		beaconViewRetriever metadata.BeaconViewRetriever
		shardID             byte
		transactionStateDB  *statedb.StateDB
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Invalid Type Metadata",
			fields: fields{
				MetadataBase: metadata.MetadataBase{
					Type: metadata.StopAutoStakingMeta,
				},
				CommitteePublicKey: key1,
			},
			args: args{
				tx: stopAutoStakingTx,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Get Subtitute Committees Public Key Error",
			fields: fields{
				MetadataBase: metadata.MetadataBase{
					Type: metadata.UnStakingMeta,
				},
				CommitteePublicKey: key1,
			},
			args: args{
				tx:                  unstakingTx,
				beaconViewRetriever: beaconViewSubtituteCommitteesError,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Public Key Is Not In Committees List Or Subtitutes List",
			fields: fields{
				MetadataBase: metadata.MetadataBase{
					Type: metadata.UnStakingMeta,
				},
				CommitteePublicKey: key1,
			},
			args: args{
				tx:                  unstakingTx,
				beaconViewRetriever: beaconViewNotFoundSubtituteCommittees,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Can't Get Staker Info In Beacon Chain",
			fields: fields{
				MetadataBase: metadata.MetadataBase{
					Type: metadata.UnStakingMeta,
				},
				CommitteePublicKey: key1,
			},
			args: args{
				tx:                  unstakingTx,
				beaconViewRetriever: beaconViewGetStakerInfoError,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Public Key Is Not In Staker Info List",
			fields: fields{
				MetadataBase: metadata.MetadataBase{
					Type: metadata.UnStakingMeta,
				},
				CommitteePublicKey: key1,
			},
			args: args{
				tx:                  unstakingTx,
				beaconViewRetriever: beaconViewNotFoundStakerInfo,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Can't Get Shard Staking Tx From Staking Tx Hash Of Staker Info",
			fields: fields{
				MetadataBase: metadata.MetadataBase{
					Type: metadata.UnStakingMeta,
				},
				CommitteePublicKey: key1,
			},
			args: args{
				tx:                  unstakingTx,
				beaconViewRetriever: beaconViewValidInput,
				shardViewRetriever:  shardViewGetStakingTxError,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Can't Find Public Key In Staking Tx",
			fields: fields{
				MetadataBase: metadata.MetadataBase{
					Type: metadata.UnStakingMeta,
				},
				CommitteePublicKey: "123",
			},
			args: args{
				tx:                  unstakingTx,
				beaconViewRetriever: beaconViewValidInput,
				shardViewRetriever:  shardViewStakingTx,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Format Of Staking Tx Hash Is Invalid",
			fields: fields{
				MetadataBase: metadata.MetadataBase{
					Type: metadata.UnStakingMeta,
				},
				CommitteePublicKey: key1,
			},
			args: args{
				tx:                  unstakingTx,
				beaconViewRetriever: beaconViewValidInput,
				shardViewRetriever:  shardViewStakingTx,
			},
			want:    false,
			wantErr: true,
		},
		// {
		// 	name: "Can't Get Transaction By Hash",
		// 	fields: fields{
		// 		MetadataBase: metadata.MetadataBase{
		// 			Type: metadata.UnStakingMeta,
		// 		},
		// 		CommitteePublicKey: key1,
		// 	},
		// 	args: args{
		// 		tx:                  unstakingTx,
		// 		beaconViewRetriever: beaconViewSubtituteCommitteesError,
		// 	},
		// 	want:    false,
		// 	wantErr: true,
		// },
		// {
		// 	name: "Send Of Staking Tx From Staker Info != Sender Of Transaction From Database",
		// 	fields: fields{
		// 		MetadataBase: metadata.MetadataBase{
		// 			Type: metadata.UnStakingMeta,
		// 		},
		// 		CommitteePublicKey: key1,
		// 	},
		// 	args: args{
		// 		tx:                  unstakingTx,
		// 		beaconViewRetriever: beaconViewSubtituteCommitteesError,
		// 	},
		// 	want:    false,
		// 	wantErr: true,
		// },
		// {
		// 	name: "Valid Input",
		// 	fields: fields{
		// 		MetadataBase: metadata.MetadataBase{
		// 			Type: metadata.UnStakingMeta,
		// 		},
		// 		CommitteePublicKey: key1,
		// 	},
		// 	args: args{
		// 		tx:                  unstakingTx,
		// 		beaconViewRetriever: beaconViewSubtituteCommitteesError,
		// 	},
		// 	want:    true,
		// 	wantErr: false,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unStakingMetadata := metadata.UnStakingMetadata{
				MetadataBase:       tt.fields.MetadataBase,
				CommitteePublicKey: tt.fields.CommitteePublicKey,
			}
			got, err := unStakingMetadata.ValidateTxWithBlockChain(tt.args.tx, tt.args.chainRetriever, tt.args.shardViewRetriever, tt.args.beaconViewRetriever, tt.args.shardID, tt.args.transactionStateDB)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnStakingMetadata.ValidateTxWithBlockChain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UnStakingMetadata.ValidateTxWithBlockChain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnStakingMetadata_ValidateSanityData(t *testing.T) {

	initPublicKey()

	privacyTx := mocks.Transaction{}
	normalTx := mocks.Transaction{}

	privacyTx.On("IsPrivacy").Return(true)
	normalTx.On("IsPrivacy").Return(false)

	type fields struct {
		MetadataBase       metadata.MetadataBase
		CommitteePublicKey string
	}
	type args struct {
		chainRetriever      metadata.ChainRetriever
		shardViewRetriever  metadata.ShardViewRetriever
		beaconViewRetriever metadata.BeaconViewRetriever
		beaconHeight        uint64
		tx                  metadata.Transaction
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		want1   bool
		wantErr bool
	}{
		{
			name: "Wrong Metadata Type",
			fields: fields{
				MetadataBase: metadata.MetadataBase{
					Type: metadata.BeaconStakingMeta,
				},
				CommitteePublicKey: key1,
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Invalid Format Committee Public Keys",
			fields: fields{
				MetadataBase: metadata.MetadataBase{
					Type: metadata.UnStakingMeta,
				},
				CommitteePublicKey: "123",
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Transaction Is Privacy",
			fields: fields{
				MetadataBase: metadata.MetadataBase{
					Type: metadata.UnStakingMeta,
				},
				CommitteePublicKey: key1,
			},
			args: args{
				tx: &privacyTx,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Valid Input",
			fields: fields{
				MetadataBase: metadata.MetadataBase{
					Type: metadata.UnStakingMeta,
				},
				CommitteePublicKey: key1,
			},
			args: args{
				tx: &normalTx,
			},
			want:    true,
			want1:   true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unStakingMetadata := metadata.UnStakingMetadata{
				MetadataBase:       tt.fields.MetadataBase,
				CommitteePublicKey: tt.fields.CommitteePublicKey,
			}
			got, got1, err := unStakingMetadata.ValidateSanityData(tt.args.chainRetriever, tt.args.shardViewRetriever, tt.args.beaconViewRetriever, tt.args.beaconHeight, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnStakingMetadata.ValidateSanityData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UnStakingMetadata.ValidateSanityData() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("UnStakingMetadata.ValidateSanityData() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestNewUnStakingMetadata(t *testing.T) {
	type args struct {
		unStakingType      int
		committeePublicKey string
	}
	tests := []struct {
		name    string
		args    args
		want    *metadata.UnStakingMetadata
		wantErr bool
	}{
		{
			name: "Unstaking Type Is Wrong",
			args: args{
				unStakingType:      metadata.BeaconStakingMeta,
				committeePublicKey: "keys",
			},
			wantErr: true,
		},
		{
			name: "Valid Input",
			args: args{
				unStakingType:      metadata.UnStakingMeta,
				committeePublicKey: "keys",
			},
			want: &metadata.UnStakingMetadata{
				MetadataBase: metadata.MetadataBase{
					Type: metadata.UnStakingMeta,
				},
				CommitteePublicKey: "keys",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := metadata.NewUnStakingMetadata(tt.args.unStakingType, tt.args.committeePublicKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewUnStakingMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUnStakingMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}
