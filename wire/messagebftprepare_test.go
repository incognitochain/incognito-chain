package wire

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	peer "github.com/libp2p/go-libp2p-peer"
)

func TestMessageBFTAgree_Hash(t *testing.T) {
	type fields struct {
		BlkHash    common.Hash
		Ri         []byte
		Pubkey     string
		ContentSig string
		Timestamp  int64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &MessageBFTAgree{
				BlkHash:    tt.fields.BlkHash,
				Ri:         tt.fields.Ri,
				Pubkey:     tt.fields.Pubkey,
				ContentSig: tt.fields.ContentSig,
				Timestamp:  tt.fields.Timestamp,
			}
			if got := msg.Hash(); got != tt.want {
				t.Errorf("MessageBFTAgree.Hash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageBFTAgree_MessageType(t *testing.T) {
	type fields struct {
		BlkHash    common.Hash
		Ri         []byte
		Pubkey     string
		ContentSig string
		Timestamp  int64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &MessageBFTAgree{
				BlkHash:    tt.fields.BlkHash,
				Ri:         tt.fields.Ri,
				Pubkey:     tt.fields.Pubkey,
				ContentSig: tt.fields.ContentSig,
				Timestamp:  tt.fields.Timestamp,
			}
			if got := msg.MessageType(); got != tt.want {
				t.Errorf("MessageBFTAgree.MessageType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageBFTAgree_MaxPayloadLength(t *testing.T) {
	type fields struct {
		BlkHash    common.Hash
		Ri         []byte
		Pubkey     string
		ContentSig string
		Timestamp  int64
	}
	type args struct {
		pver int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &MessageBFTAgree{
				BlkHash:    tt.fields.BlkHash,
				Ri:         tt.fields.Ri,
				Pubkey:     tt.fields.Pubkey,
				ContentSig: tt.fields.ContentSig,
				Timestamp:  tt.fields.Timestamp,
			}
			if got := msg.MaxPayloadLength(tt.args.pver); got != tt.want {
				t.Errorf("MessageBFTAgree.MaxPayloadLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageBFTAgree_JsonSerialize(t *testing.T) {
	type fields struct {
		BlkHash    common.Hash
		Ri         []byte
		Pubkey     string
		ContentSig string
		Timestamp  int64
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &MessageBFTAgree{
				BlkHash:    tt.fields.BlkHash,
				Ri:         tt.fields.Ri,
				Pubkey:     tt.fields.Pubkey,
				ContentSig: tt.fields.ContentSig,
				Timestamp:  tt.fields.Timestamp,
			}
			got, err := msg.JsonSerialize()
			if (err != nil) != tt.wantErr {
				t.Errorf("MessageBFTAgree.JsonSerialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MessageBFTAgree.JsonSerialize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageBFTAgree_JsonDeserialize(t *testing.T) {
	type fields struct {
		BlkHash    common.Hash
		Ri         []byte
		Pubkey     string
		ContentSig string
		Timestamp  int64
	}
	type args struct {
		jsonStr string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &MessageBFTAgree{
				BlkHash:    tt.fields.BlkHash,
				Ri:         tt.fields.Ri,
				Pubkey:     tt.fields.Pubkey,
				ContentSig: tt.fields.ContentSig,
				Timestamp:  tt.fields.Timestamp,
			}
			if err := msg.JsonDeserialize(tt.args.jsonStr); (err != nil) != tt.wantErr {
				t.Errorf("MessageBFTAgree.JsonDeserialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMessageBFTAgree_SetSenderID(t *testing.T) {
	type fields struct {
		BlkHash    common.Hash
		Ri         []byte
		Pubkey     string
		ContentSig string
		Timestamp  int64
	}
	type args struct {
		senderID peer.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &MessageBFTAgree{
				BlkHash:    tt.fields.BlkHash,
				Ri:         tt.fields.Ri,
				Pubkey:     tt.fields.Pubkey,
				ContentSig: tt.fields.ContentSig,
				Timestamp:  tt.fields.Timestamp,
			}
			if err := msg.SetSenderID(tt.args.senderID); (err != nil) != tt.wantErr {
				t.Errorf("MessageBFTAgree.SetSenderID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMessageBFTAgree_SignMsg(t *testing.T) {
	type fields struct {
		BlkHash    common.Hash
		Ri         []byte
		Pubkey     string
		ContentSig string
		Timestamp  int64
	}
	type args struct {
		keySet *incognitokey.KeySet
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &MessageBFTAgree{
				BlkHash:    tt.fields.BlkHash,
				Ri:         tt.fields.Ri,
				Pubkey:     tt.fields.Pubkey,
				ContentSig: tt.fields.ContentSig,
				Timestamp:  tt.fields.Timestamp,
			}
			if err := msg.SignMsg(tt.args.keySet); (err != nil) != tt.wantErr {
				t.Errorf("MessageBFTAgree.SignMsg() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMessageBFTAgree_VerifyMsgSanity(t *testing.T) {
	type fields struct {
		BlkHash    common.Hash
		Ri         []byte
		Pubkey     string
		ContentSig string
		Timestamp  int64
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &MessageBFTAgree{
				BlkHash:    tt.fields.BlkHash,
				Ri:         tt.fields.Ri,
				Pubkey:     tt.fields.Pubkey,
				ContentSig: tt.fields.ContentSig,
				Timestamp:  tt.fields.Timestamp,
			}
			if err := msg.VerifyMsgSanity(); (err != nil) != tt.wantErr {
				t.Errorf("MessageBFTAgree.VerifyMsgSanity() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
