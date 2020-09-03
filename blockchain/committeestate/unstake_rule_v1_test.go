package committeestate

import (
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
)

var key = "121VhftSAygpEJZ6i9jGkEKLMQTKTiiHzeUfeuhpQCcLZtys8FazpWwytpHebkAwgCxvqgUUF13fcSMtp5dgV1YkbRMj3z42TW2EebzAaiGg2DkGPodckN2UsbqhVDibpMgJUHVkLXardemfLdgUqWGtymdxaaRyPM38BAZcLpo2pAjxKv5vG5Uh9zHMkn7ZHtdNHmBmhG8B46UeiGBXYTwhyMe9KGS83jCMPAoUwHhTEXj5qQh6586dHjVxwEkRzp7SKn9iG1FFWdJ97xEkP2ezAapNQ46quVrMggcHFvoZofs1xdd4o5vAmPKnPTZtGTKunFiTWGnpSG9L6r5QpcmapqvRrK5SiuFhNM5DqgzUeHBb7fTfoiWd2N29jkbTGSq8CPUSjx3zdLR9sZguvPdnAA8g25cFPGSZt8aEnFJoPRzM"
var key2 = "121VhftSAygpEJZ6i9jGkEqPGAXcmKffwMbzpwxnEfzJxen4oZKPukWAUBbqvV5xPnowZ2eQmAj2mEebG2oexebQPh1MPFC6vEZAk6i7AiRPrZmfaRrRVrBp4WXnVJmL3xK4wzTfkR2rZkhUmSZm112TTyhDNkDQSaBGJkexrPbryqUygazCA2eyo6LnK5qs7jz2RhhsWqUTQ3sQJUuFcYdf2pSnYwhqZqphDCSRizDHeysaua5L7LwS8fY7KZHhPgTuFjvUWWnWSRTmV8u1dTY5kcmMdDZsPiyN9WfqjgVoTFNALjFG8U4GMvzV3kKwVVjuPMsM2XqyPDVpdNQUgLnv2bJS8Tr22A9NgF1FQfWyAny1DYyY3N5H3tfCggsybzZXzrbYPPgokvEynac91y8hPkRdgKW1e7FHzuBnEisPuKzy"
var key3 = "121VhftSAygpEJZ6i9jGkGLcYhJBeaJTGY5aFjqQA2WwyxU69Utrviuy9AJ3ATkeEyigVGScQUZw22cD1HeFKiyASYAs82WEamujt3nefYA9FPhURBpRTn6jDmGKUdb4QNbs7HVCJkRRaL9aktg1yaQaZE8TJFg2UeE9tBqUdmvD8fy36aDCYM5W86jaTVCXeEJQWPxUunP2EEL3e283PJ8zqPeBkpoFvkvhB28Hk3oRDeCCTC7QhbaV18ayKeToYqAxoUMBBihanfA33ixeX1daeKpajLCgDZ6jrfphwdYwQbf7dMcZ2NVvQ1a5JUCTJUZypwgKRt8tnTAKCowt2L1KNGP4NJJZm61cfHAGbKRyG9QxCJgK2SdMKsKPVefZSc9LbVaB7VeBby5LHxvMoCD7bN7g1HYRp4BX9n1fZJUeEkVa"
var key4 = "121VhftSAygpEJZ6i9jGkDjJj7e2cfgQvrLsPsmLhGMmGD9U9Knffa1MZAw79EijnpueVfTStN2VYt5jRqEr2DTjVqzUinwHVKWH4Tg4szHUntiBdWeqzNC4E8iiwC9Y2KtcRr3hBkpfqvyuBvchigatrigRvFVWu8H2RQqjvopLL51DQ4LFD87L9Zgj9HhasMeyr6f37yirs47JgtGs4BM7EhhpM5zD3TCsFabPphtwDKnfuLMaGzoAw5fM8zEXvdLMuohk96oayjdYothncdtZom17DxB1Mmw535eEjxBwz9ELoZRKk3LYiheSd4xGN9QsxrT2WnZCTd8B5QktARte5S91QYvRMixKC8UEuovQhXt8jMZNkq7CmMeXoybfYdmNaAHuqbY1QeUT2AgaqPho4ay3z5eeKRhnB28H18RGWQ1L"
var key5 = "121VhftSAygpEJZ6i9jGkS3D5FaqhzhSF79YFYKHLTHMY5erPhm5vT9VxMtFdWbUVfmhKvfKosXiUKsygyw8knbejNWineCFpx35KegXBbfnVv6AcE3KD4Rs46pDKrqDvWmpaPJoUDdiJeVPQQsFuTykMGs1txt14hhnWMWx9Bf8caDpxR3SKQY7PyHbEhRhdasL3eJC3X1y83PkzJistXPHFdoK4bszU5iE8EiMrXP5GiHTLLTyTpRxScg6AVnrFnmRzPsEMXJAz5zmwUkxwGNrj5iBi7ZJBHo5m3bTVYdQqTSBgVXSqAGZ6fPqAXPGkH6NfgGeZhXR33D3Q4JhEoBs4QWnr89gaVUDAwGXFoXEVfmGwGFFy7jPdLYKuc2u1yJ9YRa1MbSxcPLATui2wmN73UAFa6uSUdN71rCDHJEfCxpS"
var key6 = "121VhftSAygpEJZ6i9jGkQXi69eX7p8fmJosf8F4KEdBSqfh3cGxGMd6sGa4hfXTg9vxq16AN97mrqerdNM6ZUGBDgPAipbaGznaHSC8gE7gBpSrVKbNb93nwXSBHSBKFVC6MK5NAFN6bpK25YHrmC248FPr3VQMf9tfG57P5TTH7KWr4bn7v2YbTxNRkZFD9JwkTmwXAwEfWJ12rrc1kMDUkAkrSYYhmpykXTjK9wEBkKFA2z5rnw24cBVL9Tt6M2BEqUM3tuEoUfhiA6E6tdPAkYc7LusTjwikzpwRbVYi4cVMCmC7Dd2UccaA2iiotuyP85AYQSUaHzV2MaF2Cv7GtLqTMm6bRqvpetU1kpkunEnQmAuLVLG7QHPRVKdkX6wRYBE6uRcJ1FaejVbbrF3Tgyh6dsMhRVgEvvvocYPULcJ5"
var key7 = "121VhftSAygpEJZ6i9jGk68R6pmXasuHTdRxSeLvBa6dWdc7Mp7c9AxfP6Po9BAi7yRnmiffbEFvs6p6zLFRxwUV1gZLa4ijV7nhPHjxBmJW9vYwV6cJFv2VCN4P1ncnUPf75U8wFxt7AXBQ4o67rsqnHrWvifisURmZzqqaRSUsKAbgqvkmnb3GPcCdjGqFgiYkbwCf4QRWEPnCCdRKabbA2SHDo3bzxJS6CiQNXmKL9SRCrfm1aDcTMUrhPg4w2Gtx8YrQZpHDRYAhgigDgUHPLyLf4Gado342tNTBi9XwvyghJQ6i4PguGrqUvRns8kJ3mbouNWLBc8tQGi3NVN7vb9fmoNm4KSDc22RWESSDkUkj6pAqBiRtJvXjS24DqKTNwQU7FJWobc8a6Qudyxabb5TksrK6d4QirEW8CkX5ahnk"
var key8 = "121VhftSAygpEJZ6i9jGkAWwCGm383V8zyMqU2VbEsymfkv3tCPRcRFWtvuTeNVH4r8TDRAdHjaM2j5Nwvw6vqEr58seiM3SMgdDeZwkv942XhG1DmwdrvBPM5RyA3Na32DXRykeHqkAoGP7HbUfUQDZzwkVi3ufHnVEsEVM2CsBTFubBR5YREZVkC4L81a4Hb7BVQZ8yap1kGpZctkTdSCCyGMge2AfqyqvhQ7zn6UCw8aMNnajprw8hJCtuSLEQXA8MwYis6X9cRjKACxYQ9hzyKCvg19PSE7ntf9fXyLxTCmcvCHdNd7cAFrBiDKJHpzp9FVwARyNJF4jEKYmfFi599njpuSSyhQTqEanKg9JnWmp2TNENCEsZ8L9DjbUwbeEWs8uS4Skvx9HeG9itgHL2T3dWKFaisAfBS9YVqVpUnGL"

var incKey, incKey2, incKey3, incKey4, incKey5, incKey6, incKey7, incKey8 *incognitokey.CommitteePublicKey

//initPublicKey init incognito public key for testing by base 58 string
func initPublicKey() {
	incKey = new(incognitokey.CommitteePublicKey)
	incKey2 = new(incognitokey.CommitteePublicKey)
	incKey3 = new(incognitokey.CommitteePublicKey)
	incKey4 = new(incognitokey.CommitteePublicKey)
	incKey5 = new(incognitokey.CommitteePublicKey)
	incKey6 = new(incognitokey.CommitteePublicKey)
	incKey7 = new(incognitokey.CommitteePublicKey)
	incKey8 = new(incognitokey.CommitteePublicKey)

	err := incKey.FromBase58(key)
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

	err = incKey5.FromBase58(key5)
	if err != nil {
		panic(err)
	}

	err = incKey6.FromBase58(key6)
	if err != nil {
		panic(err)
	}

	err = incKey7.FromBase58(key7)
	if err != nil {
		panic(err)
	}

	err = incKey8.FromBase58(key8)
	if err != nil {
		panic(err)
	}
}

func TestBeaconCommitteeStateV1_processUnstakeInstruction(t *testing.T) {

	// Init data for testcases
	initStateDB()
	initPublicKey()

	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)
	rewardReceiverkey := incKey.GetIncKeyBase58()
	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})

	validSDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	hash, err := common.Hash{}.NewHashFromStr("123")
	statedb.StoreStakerInfoV1(
		validSDB,
		[]incognitokey.CommitteePublicKey{*incKey},
		map[string]privacy.PaymentAddress{
			rewardReceiverkey: paymentAddress,
		},
		map[string]bool{
			key: true,
		},
		map[string]common.Hash{
			key: *hash,
		},
	)
	committeePublicKeyWrongFormat := incognitokey.CommitteePublicKey{}
	committeePublicKeyWrongFormat.MiningPubKey = nil
	//

	type fields struct {
		beaconCommittee             []incognitokey.CommitteePublicKey
		beaconSubstitute            []incognitokey.CommitteePublicKey
		nextEpochShardCandidate     []incognitokey.CommitteePublicKey
		currentEpochShardCandidate  []incognitokey.CommitteePublicKey
		nextEpochBeaconCandidate    []incognitokey.CommitteePublicKey
		currentEpochBeaconCandidate []incognitokey.CommitteePublicKey
		shardCommittee              map[byte][]incognitokey.CommitteePublicKey
		shardSubstitute             map[byte][]incognitokey.CommitteePublicKey
		autoStake                   map[string]bool
		rewardReceiver              map[string]privacy.PaymentAddress
		stakingTx                   map[string]common.Hash
		mu                          *sync.RWMutex
	}
	type args struct {
		unstakeInstruction *instruction.UnstakeInstruction
		env                *BeaconCommitteeStateEnvironment
		committeeChange    *CommitteeChange
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *CommitteeChange
		want1   [][]string
		wantErr bool
	}{
		{
			name: "[Subtitute List] Invalid Format Of Committee Public Key In Unstake Instruction",
			fields: fields{
				nextEpochShardCandidate: []incognitokey.CommitteePublicKey{*incKey},
			},
			args: args{
				unstakeInstruction: &instruction.UnstakeInstruction{
					CommitteePublicKeys: []string{"123"},
				},
				env: &BeaconCommitteeStateEnvironment{
					substituteCandidates: []string{"123"},
				},
				committeeChange: &CommitteeChange{},
			},
			want:    &CommitteeChange{},
			wantErr: true,
		},
		{
			name: "[Subtitute List] Can't find staker info in database",
			fields: fields{
				nextEpochShardCandidate: []incognitokey.CommitteePublicKey{*incKey},
				autoStake: map[string]bool{
					key: true,
				},
				rewardReceiver: map[string]privacy.PaymentAddress{
					rewardReceiverkey: paymentAddress,
				},
				stakingTx: map[string]common.Hash{
					key: *hash,
				},
			},
			args: args{
				unstakeInstruction: &instruction.UnstakeInstruction{
					CommitteePublicKeys: []string{key2},
				},
				env: &BeaconCommitteeStateEnvironment{
					substituteCandidates: []string{key2},
					ConsensusStateDB:     sDB,
				},
				committeeChange: &CommitteeChange{},
			},
			want:    &CommitteeChange{},
			wantErr: true,
		},
		{
			name: "Valid Input Key In Subtitutes List",
			fields: fields{
				nextEpochShardCandidate: []incognitokey.CommitteePublicKey{*incKey},
				autoStake: map[string]bool{
					key: true,
				},
				rewardReceiver: map[string]privacy.PaymentAddress{
					rewardReceiverkey: paymentAddress,
				},
				stakingTx: map[string]common.Hash{
					key: *hash,
				},
			},
			args: args{
				committeeChange: &CommitteeChange{},
				unstakeInstruction: &instruction.UnstakeInstruction{
					CommitteePublicKeys: []string{key},
				},
				env: &BeaconCommitteeStateEnvironment{
					substituteCandidates: []string{key},
					ConsensusStateDB:     validSDB,
				},
			},
			want: &CommitteeChange{
				NextEpochShardCandidateRemoved: []incognitokey.CommitteePublicKey{*incKey},
			},
			want1: [][]string{
				[]string{
					instruction.RETURN_ACTION,
					key,
					"0",
					hash.String(),
					"100",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid Input Key In Validators List",
			fields: fields{
				currentEpochShardCandidate: []incognitokey.CommitteePublicKey{*incKey},
			},
			args: args{
				unstakeInstruction: &instruction.UnstakeInstruction{
					CommitteePublicKeys: []string{key},
				},
				env: &BeaconCommitteeStateEnvironment{
					allSubstituteCommittees: []string{key},
				},
				committeeChange: &CommitteeChange{},
			},
			want:    &CommitteeChange{},
			want1:   [][]string{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV1{
				beaconCommittee:             tt.fields.beaconCommittee,
				beaconSubstitute:            tt.fields.beaconSubstitute,
				nextEpochShardCandidate:     tt.fields.nextEpochShardCandidate,
				currentEpochShardCandidate:  tt.fields.currentEpochShardCandidate,
				nextEpochBeaconCandidate:    tt.fields.nextEpochBeaconCandidate,
				currentEpochBeaconCandidate: tt.fields.currentEpochBeaconCandidate,
				shardCommittee:              tt.fields.shardCommittee,
				shardSubstitute:             tt.fields.shardSubstitute,
				autoStake:                   tt.fields.autoStake,
				rewardReceiver:              tt.fields.rewardReceiver,
				stakingTx:                   tt.fields.stakingTx,
				mu:                          tt.fields.mu,
			}
			got, got1, err := b.processUnstakeInstruction(tt.args.unstakeInstruction, tt.args.env, tt.args.committeeChange)

			sDB.ClearObjects() // Clear Object From StateDB

			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeStateV1.processUnstakeInstruction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV1.processUnstakeInstruction() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("BeaconCommitteeStateV1.processUnstakeInstruction() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestBeaconCommitteeStateV1_getSubtituteCandidates(t *testing.T) {

	initPublicKey()

	type fields struct {
		beaconCommittee             []incognitokey.CommitteePublicKey
		beaconSubstitute            []incognitokey.CommitteePublicKey
		nextEpochShardCandidate     []incognitokey.CommitteePublicKey
		currentEpochShardCandidate  []incognitokey.CommitteePublicKey
		nextEpochBeaconCandidate    []incognitokey.CommitteePublicKey
		currentEpochBeaconCandidate []incognitokey.CommitteePublicKey
		shardCommittee              map[byte][]incognitokey.CommitteePublicKey
		shardSubstitute             map[byte][]incognitokey.CommitteePublicKey
		autoStake                   map[string]bool
		rewardReceiver              map[string]privacy.PaymentAddress
		stakingTx                   map[string]common.Hash
		mu                          *sync.RWMutex
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{
			name: "Valid Input",
			fields: fields{
				nextEpochShardCandidate: []incognitokey.CommitteePublicKey{
					*incKey,
				},
			},
			want:    []string{key},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV1{
				beaconCommittee:             tt.fields.beaconCommittee,
				beaconSubstitute:            tt.fields.beaconSubstitute,
				nextEpochShardCandidate:     tt.fields.nextEpochShardCandidate,
				currentEpochShardCandidate:  tt.fields.currentEpochShardCandidate,
				nextEpochBeaconCandidate:    tt.fields.nextEpochBeaconCandidate,
				currentEpochBeaconCandidate: tt.fields.currentEpochBeaconCandidate,
				shardCommittee:              tt.fields.shardCommittee,
				shardSubstitute:             tt.fields.shardSubstitute,
				autoStake:                   tt.fields.autoStake,
				rewardReceiver:              tt.fields.rewardReceiver,
				stakingTx:                   tt.fields.stakingTx,
				mu:                          tt.fields.mu,
			}
			got, err := b.getSubstituteCandidates()
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeStateV1.getSubstituteCandidates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV1.getSubstituteCandidates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBeaconCommitteeStateV1_getAllSubstituteCommittees(t *testing.T) {

	initPublicKey()

	type fields struct {
		beaconCommittee             []incognitokey.CommitteePublicKey
		beaconSubstitute            []incognitokey.CommitteePublicKey
		nextEpochShardCandidate     []incognitokey.CommitteePublicKey
		currentEpochShardCandidate  []incognitokey.CommitteePublicKey
		nextEpochBeaconCandidate    []incognitokey.CommitteePublicKey
		currentEpochBeaconCandidate []incognitokey.CommitteePublicKey
		shardCommittee              map[byte][]incognitokey.CommitteePublicKey
		shardSubstitute             map[byte][]incognitokey.CommitteePublicKey
		autoStake                   map[string]bool
		rewardReceiver              map[string]privacy.PaymentAddress
		stakingTx                   map[string]common.Hash
		mu                          *sync.RWMutex
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{
			name: "Valid Input",
			fields: fields{
				currentEpochShardCandidate: []incognitokey.CommitteePublicKey{
					*incKey,
				},
			},
			want:    []string{key},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV1{
				beaconCommittee:             tt.fields.beaconCommittee,
				beaconSubstitute:            tt.fields.beaconSubstitute,
				nextEpochShardCandidate:     tt.fields.nextEpochShardCandidate,
				currentEpochShardCandidate:  tt.fields.currentEpochShardCandidate,
				nextEpochBeaconCandidate:    tt.fields.nextEpochBeaconCandidate,
				currentEpochBeaconCandidate: tt.fields.currentEpochBeaconCandidate,
				shardCommittee:              tt.fields.shardCommittee,
				shardSubstitute:             tt.fields.shardSubstitute,
				autoStake:                   tt.fields.autoStake,
				rewardReceiver:              tt.fields.rewardReceiver,
				stakingTx:                   tt.fields.stakingTx,
				mu:                          tt.fields.mu,
			}
			got, err := b.getAllSubstituteCommittees()
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeStateV1.getValidators() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV1.getValidators() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBeaconCommitteeStateV1_processUnstakeChange(t *testing.T) {
	initStateDB()
	initPublicKey()

	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)
	rewardReceiver := incKey.GetIncKeyBase58()
	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})
	hash, err := common.Hash{}.NewHashFromStr("123")

	type fields struct {
		beaconCommittee             []incognitokey.CommitteePublicKey
		beaconSubstitute            []incognitokey.CommitteePublicKey
		nextEpochShardCandidate     []incognitokey.CommitteePublicKey
		currentEpochShardCandidate  []incognitokey.CommitteePublicKey
		nextEpochBeaconCandidate    []incognitokey.CommitteePublicKey
		currentEpochBeaconCandidate []incognitokey.CommitteePublicKey
		shardCommittee              map[byte][]incognitokey.CommitteePublicKey
		shardSubstitute             map[byte][]incognitokey.CommitteePublicKey
		autoStake                   map[string]bool
		rewardReceiver              map[string]privacy.PaymentAddress
		stakingTx                   map[string]common.Hash
		mu                          *sync.RWMutex
	}
	type args struct {
		committeeChange *CommitteeChange
		env             *BeaconCommitteeStateEnvironment
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *CommitteeChange
		wantErr bool
	}{
		{
			name:   "Invalid Format Of Public Key In List Unstake Of CommitteeChange",
			fields: fields{},
			args: args{
				committeeChange: &CommitteeChange{
					Unstake: []string{"123"},
				},
				env: &BeaconCommitteeStateEnvironment{},
			},
			want: &CommitteeChange{
				Unstake: []string{"123"},
			},
			wantErr: true,
		},
		{
			name:   "Error When Store Staker Info",
			fields: fields{},
			args: args{
				committeeChange: &CommitteeChange{
					Unstake: []string{key},
				},
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB: sDB,
				},
			},
			want: &CommitteeChange{
				Unstake: []string{key},
			},
			wantErr: true,
		},
		{
			name: "Valid Input",
			fields: fields{
				autoStake: map[string]bool{
					key: true,
				},
				rewardReceiver: map[string]privacy.PaymentAddress{
					rewardReceiver: paymentAddress,
				},
				stakingTx: map[string]common.Hash{
					key: *hash,
				},
			},
			args: args{
				committeeChange: &CommitteeChange{
					Unstake: []string{key},
				},
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB: sDB,
				},
			},
			want: &CommitteeChange{
				Unstake: []string{key},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV1{
				beaconCommittee:             tt.fields.beaconCommittee,
				beaconSubstitute:            tt.fields.beaconSubstitute,
				nextEpochShardCandidate:     tt.fields.nextEpochShardCandidate,
				currentEpochShardCandidate:  tt.fields.currentEpochShardCandidate,
				nextEpochBeaconCandidate:    tt.fields.nextEpochBeaconCandidate,
				currentEpochBeaconCandidate: tt.fields.currentEpochBeaconCandidate,
				shardCommittee:              tt.fields.shardCommittee,
				shardSubstitute:             tt.fields.shardSubstitute,
				autoStake:                   tt.fields.autoStake,
				rewardReceiver:              tt.fields.rewardReceiver,
				stakingTx:                   tt.fields.stakingTx,
				mu:                          tt.fields.mu,
			}
			got, err := b.processUnstakeChange(tt.args.committeeChange, tt.args.env)

			sDB.ClearObjects() // Clear objects for next test case

			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeStateV1.processUnstakeChange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV1.processUnstakeChange() = %v, want %v", got, tt.want)
			}
		})
	}
}
