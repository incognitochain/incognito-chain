package peerv2

import (
	"reflect"
	"testing"
)

func TestDiscoverHighWay(t *testing.T) {
	type args struct {
		discoverPeerAddress string
		shardsStr           []string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string][]string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	// x, err := DiscoverHighWay("0.0.0.0:9330", []string{"all"})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DiscoverHighWay(tt.args.discoverPeerAddress, tt.args.shardsStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("DiscoverHighWay() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DiscoverHighWay() = %v, want %v", got, tt.want)
			}
		})
	}
}
