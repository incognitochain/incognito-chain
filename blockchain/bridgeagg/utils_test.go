package bridgeagg

import (
	"testing"
)

func TestCalculateActualAmount(t *testing.T) {
	type args struct {
		x        uint64
		y        uint64
		deltaX   uint64
		operator byte
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{
			name: "Cannot recognize operator",
			args: args{
				operator: 3,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "y == 0, first shield",
			args: args{
				deltaX:   100,
				operator: AddOperator,
			},
			want:    100,
			wantErr: false,
		},
		{
			name: "y == 0, second shield",
			args: args{
				x:        100,
				deltaX:   300,
				operator: AddOperator,
			},
			want:    300,
			wantErr: false,
		},
		{
			name: "y != 0, first shield",
			args: args{
				y:        100,
				deltaX:   100,
				operator: AddOperator,
			},
			want:    200,
			wantErr: false,
		},
		{
			name: "y != 0, second shield",
			args: args{
				y:        100,
				deltaX:   100,
				x:        1000,
				operator: AddOperator,
			},
			want:    110,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateActualAmount(tt.args.x, tt.args.y, tt.args.deltaX, tt.args.operator)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateActualAmount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CalculateActualAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEstimateActualAmountByBurntAmount(t *testing.T) {
	type args struct {
		x           uint64
		y           uint64
		burntAmount uint64
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{
			name: "burntAmount == 0",
			args: args{
				burntAmount: 0,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "y == 0, burntAmount > x",
			args: args{
				burntAmount: 10,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "y == 0, burntAmount < x",
			args: args{
				x:           150,
				burntAmount: 100,
			},
			want:    100,
			wantErr: false,
		},
		{
			name: "y != 0, burntAmount <= x",
			args: args{
				x:           1000,
				y:           100,
				burntAmount: 111,
			},
			want:    100,
			wantErr: false,
		},
		{
			name: "y != 0, burntAmount <= x - example 2",
			args: args{
				x:           50000000,
				y:           1000000,
				burntAmount: 300,
			},
			want:    294,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EstimateActualAmountByBurntAmount(tt.args.x, tt.args.y, tt.args.burntAmount)
			if (err != nil) != tt.wantErr {
				t.Errorf("EstimateActualAmountByBurntAmount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EstimateActualAmountByBurntAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}
