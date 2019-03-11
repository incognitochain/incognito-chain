package wire

/*import (
	"encoding/hex"
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/big0t/constant-chain/transaction"
)

const (
	MaxTxRegisterationPayload = 4000000 // 4 Mb
)

type MessageRegistration struct {
	Transaction metadata.Transaction
}

func (msg MessageRegistration) MessageType() string {
	return CmdRegisteration
}

func (msg MessageRegistration) MaxPayloadLength(pver int) int {
	return MaxTxRegisterationPayload
}

func (msg MessageRegistration) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg MessageRegistration) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), msg)
	return err
}

func (msg MessageRegistration) SetSenderID(senderID peer.ID) error {
	return nil
}*/
