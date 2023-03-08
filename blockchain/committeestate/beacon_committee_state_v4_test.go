package committeestate

// func Test_getTotalLockingEpoch(t *testing.T) {
// 	type args struct {
// 		perf   uint64
// 		factor uint64
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want uint64
// 	}{
// 		{
// 			name: "a",
// 			args: args{
// 				perf:   100,
// 				factor: 1,
// 			},
// 			want: 19,
// 		},
// 		{
// 			name: "b",
// 			args: args{
// 				perf:   100,
// 				factor: 10,
// 			},
// 			want: 190,
// 		},
// 		{
// 			name: "c",
// 			args: args{
// 				perf:   100,
// 				factor: 3,
// 			},
// 			want: 19 * 3,
// 		},
// 		{
// 			name: "d",
// 			args: args{
// 				perf:   1000,
// 				factor: 3,
// 			},
// 			want: 10 * 3,
// 		},
// 		{
// 			name: "e",
// 			args: args{
// 				perf:   1000,
// 				factor: 10,
// 			},
// 			want: 10 * 10,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := getTotalLockingEpoch(tt.args.perf, tt.args.factor); got != tt.want {
// 				t.Errorf("getTotalLockingEpoch() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
