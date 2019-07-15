package wire

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/incognitokey"
	peer "github.com/libp2p/go-libp2p-peer"
)

func TestMessageGetAddr_Hash(t *testing.T) {
	tests := []struct {
		name string
		msg  *MessageGetAddr
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.msg.Hash(); got != tt.want {
				t.Errorf("MessageGetAddr.Hash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageGetAddr_MessageType(t *testing.T) {
	tests := []struct {
		name string
		msg  *MessageGetAddr
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.msg.MessageType(); got != tt.want {
				t.Errorf("MessageGetAddr.MessageType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageGetAddr_MaxPayloadLength(t *testing.T) {
	type args struct {
		pver int
	}
	tests := []struct {
		name string
		msg  *MessageGetAddr
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.msg.MaxPayloadLength(tt.args.pver); got != tt.want {
				t.Errorf("MessageGetAddr.MaxPayloadLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageGetAddr_JsonSerialize(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MessageGetAddr
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.msg.JsonSerialize()
			if (err != nil) != tt.wantErr {
				t.Errorf("MessageGetAddr.JsonSerialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MessageGetAddr.JsonSerialize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageGetAddr_JsonDeserialize(t *testing.T) {
	type args struct {
		jsonStr string
	}
	tests := []struct {
		name    string
		msg     *MessageGetAddr
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.msg.JsonDeserialize(tt.args.jsonStr); (err != nil) != tt.wantErr {
				t.Errorf("MessageGetAddr.JsonDeserialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMessageGetAddr_SetSenderID(t *testing.T) {
	type args struct {
		senderID peer.ID
	}
	tests := []struct {
		name    string
		msg     *MessageGetAddr
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.msg.SetSenderID(tt.args.senderID); (err != nil) != tt.wantErr {
				t.Errorf("MessageGetAddr.SetSenderID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMessageGetAddr_SignMsg(t *testing.T) {
	type args struct {
		in0 *incognitokey.KeySet
	}
	tests := []struct {
		name    string
		msg     *MessageGetAddr
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.msg.SignMsg(tt.args.in0); (err != nil) != tt.wantErr {
				t.Errorf("MessageGetAddr.SignMsg() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMessageGetAddr_VerifyMsgSanity(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MessageGetAddr
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.msg.VerifyMsgSanity(); (err != nil) != tt.wantErr {
				t.Errorf("MessageGetAddr.VerifyMsgSanity() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
