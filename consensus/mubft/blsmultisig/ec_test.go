package blsmultisig

import "testing"

func TestCurve(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	NewCurve()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			NewCurve()
		})
	}
}
