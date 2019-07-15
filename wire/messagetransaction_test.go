package wire

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/incognitokey"
	peer "github.com/libp2p/go-libp2p-peer"
)

func TestMessageTx_Hash(t *testing.T) {
	tests := []struct {
		name string
		msg  *MessageTx
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.msg.Hash(); got != tt.want {
				t.Errorf("MessageTx.Hash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageTx_MessageType(t *testing.T) {
	tests := []struct {
		name string
		msg  *MessageTx
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.msg.MessageType(); got != tt.want {
				t.Errorf("MessageTx.MessageType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageTx_MaxPayloadLength(t *testing.T) {
	type args struct {
		pver int
	}
	tests := []struct {
		name string
		msg  *MessageTx
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.msg.MaxPayloadLength(tt.args.pver); got != tt.want {
				t.Errorf("MessageTx.MaxPayloadLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageTx_JsonSerialize(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MessageTx
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.msg.JsonSerialize()
			if (err != nil) != tt.wantErr {
				t.Errorf("MessageTx.JsonSerialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MessageTx.JsonSerialize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageTx_JsonDeserialize(t *testing.T) {
	type args struct {
		jsonStr string
	}
	tests := []struct {
		name    string
		msg     *MessageTx
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.msg.JsonDeserialize(tt.args.jsonStr); (err != nil) != tt.wantErr {
				t.Errorf("MessageTx.JsonDeserialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMessageTx_SetSenderID(t *testing.T) {
	type args struct {
		senderID peer.ID
	}
	tests := []struct {
		name    string
		msg     *MessageTx
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.msg.SetSenderID(tt.args.senderID); (err != nil) != tt.wantErr {
				t.Errorf("MessageTx.SetSenderID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMessageTx_SignMsg(t *testing.T) {
	type args struct {
		in0 *incognitokey.KeySet
	}
	tests := []struct {
		name    string
		msg     *MessageTx
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.msg.SignMsg(tt.args.in0); (err != nil) != tt.wantErr {
				t.Errorf("MessageTx.SignMsg() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMessageTx_VerifyMsgSanity(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MessageTx
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.msg.VerifyMsgSanity(); (err != nil) != tt.wantErr {
				t.Errorf("MessageTx.VerifyMsgSanity() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
