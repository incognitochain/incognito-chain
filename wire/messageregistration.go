package wire

/*import (
	"encoding/hex"
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/transaction"
)

const (
	MaxTxRegisterationPayload = 4000000 // 4 Mb
)

type MessageRegistration struct {
	Transaction metadata.Transaction
}

func (self MessageRegistration) MessageType() string {
	return CmdRegisteration
}

func (self MessageRegistration) MaxPayloadLength(pver int) int {
	return MaxTxRegisterationPayload
}

func (self MessageRegistration) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageRegistration) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), self)
	return err
}

func (self MessageRegistration) SetSenderID(senderID peer.ID) error {
	return nil
}*/
