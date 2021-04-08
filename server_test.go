package main

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/peer"
	"github.com/incognitochain/incognito-chain/wire"
	"testing"
)

func TestServer_OnFinishSync(t *testing.T) {
	type fields struct {
		blockChain *blockchain.BlockChain
	}
	type args struct {
		p   *peer.PeerConn
		msg *wire.MessageFinishSync
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		//TODO: @hung add testcase
		// Testcase 1: beaconBestState.SyncValidator > 0, finish sync validator = 0
		// Testcase 2: beaconBestState.SyncValidator = 0, finish sync validator > 0
		// Testcase 3: beaconBestState.SyncValidator > 0, finish sync validator > 0, but no 100% identical
		// Testcase 4: beaconBestState.SyncValidator > 0, finish sync validator > 0, 100% identical
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverObj := &Server{
				blockChain: tt.fields.blockChain,
			}
			serverObj.OnFinishSync(tt.args.p, tt.args.msg)
		})
	}
}
