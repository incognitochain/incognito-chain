package instruction

import "github.com/incognitochain/incognito-chain/incognitokey"

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

// func TestCommitteeStateInstruction_ToString(t *testing.T) {

// 	type fields struct {
// 		SwapInstructions          []*SwapInstruction
// 		StakeInstructions         []*StakeInstruction
// 		AssignInstructions        []*AssignInstruction
// 		StopAutoStakeInstructions []*StopAutoStakeInstruction
// 	}

// 	type args struct {
// 		action string
// 	}

// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 		want   [][]string
// 	}{
// 		{
// 			name:   "Swap Instruction",
// 			fields: fields{
// 				SwapInstructions:          []*SwapInstruction{
// 					&SwapInstruction{
// 						InPublicKeys:        nil,
// 						InPublicKeyStructs:  nil,
// 						OutPublicKeys:       nil,
// 						OutPublicKeyStructs: nil,
// 						ChainID:             0,
// 						PunishedPublicKeys:  "",
// 						NewRewardReceivers:  nil,
// 						IsReplace:           false,
// 					},
// 					&SwapInstruction{
// 						InPublicKeys:        nil,
// 						InPublicKeyStructs:  nil,
// 						OutPublicKeys:       nil,
// 						OutPublicKeyStructs: nil,
// 						ChainID:             0,
// 						PunishedPublicKeys:  "",
// 						NewRewardReceivers:  nil,
// 						IsReplace:           false,
// 					},
// 				},
// 				StakeInstructions:         nil,
// 				AssignInstructions:        nil,
// 				StopAutoStakeInstructions: nil,
// 			},
// 			args:   args{SWAP_ACTION},
// 			want:   nil,
// 		},
// 		{
// 			name:   "Stake Instruction",
// 			fields: fields{
// 				SwapInstructions:          nil,
// 				StakeInstructions:         nil,
// 				AssignInstructions:        nil,
// 				StopAutoStakeInstructions: nil,
// 			},
// 			args:   args{STAKE_ACTION},
// 			want:   nil,
// 		},
// 		{
// 			name:   "Assign Instruction",
// 			fields: fields{
// 				SwapInstructions:          nil,
// 				StakeInstructions:         nil,
// 				AssignInstructions:        nil,
// 				StopAutoStakeInstructions: nil,
// 			},
// 			args:   args{ASSIGN_ACTION},
// 			want:   nil,
// 		},
// 		{
// 			name:   "Stop Auto Staking Instruction",
// 			fields: fields{
// 				SwapInstructions:          nil,
// 				StakeInstructions:         nil,
// 				AssignInstructions:        nil,
// 				StopAutoStakeInstructions: nil,
// 			},
// 			args:   args{STOP_AUTO_STAKE_ACTION},
// 			want:   nil,
// 		},
// 		{
// 			name:   "Invalid Instruction",
// 			fields: fields{
// 				SwapInstructions:          nil,
// 				StakeInstructions:         nil,
// 				AssignInstructions:        nil,
// 				StopAutoStakeInstructions: nil,
// 			},
// 			args:   args{""},
// 			want:   nil,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			i := &CommitteeStateInstruction{
// 				SwapInstructions:          tt.fields.SwapInstructions,
// 				StakeInstructions:         tt.fields.StakeInstructions,
// 				AssignInstructions:        tt.fields.AssignInstructions,
// 				StopAutoStakeInstructions: tt.fields.StopAutoStakeInstructions,
// 			}
// 			if got := i.ToString(tt.args.action); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("ToString() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

//func TestCommitteeStateInstruction_ValidateAndFilterStakeInstructionsV1(t *testing.T) {
//	type fields struct {
//		SwapInstructions          []*SwapInstruction
//		StakeInstructions         []*StakeInstruction
//		AssignInstructions        []*AssignInstruction
//		StopAutoStakeInstructions []*StopAutoStakeInstruction
//	}
//	type args struct {
//		v *ViewEnvironment
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			i := &CommitteeStateInstruction{
//				SwapInstructions:          tt.fields.SwapInstructions,
//				StakeInstructions:         tt.fields.StakeInstructions,
//				AssignInstructions:        tt.fields.AssignInstructions,
//				StopAutoStakeInstructions: tt.fields.StopAutoStakeInstructions,
//			}
//		})
//	}
//}

// func TestImportCommitteeStateInstruction(t *testing.T) {
// 	type args struct {
// 		instructions [][]string
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want *CommitteeStateInstruction
// 	}{
// 		{},
// 		{},
// 		{},
// 		{},
// 		{},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := ImportCommitteeStateInstruction(tt.args.instructions); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("ImportCommitteeStateInstruction() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
