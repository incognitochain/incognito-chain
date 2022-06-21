package coin

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/wallet"
)

// OTAReceiver holds the data necessary to send a coin to your receiver with privacy.
// It is somewhat equivalent in usage with PaymentAddress
type OTAReceiver struct {
	PublicKey *operation.Point
	TxRandom  TxRandom
}

// IsValid() checks the validity of this OTAReceiver (all referenced Points must be valid).
// Note that some sanity checks are already done when unmarshalling
func (recv OTAReceiver) IsValid() bool {
	_, err := recv.TxRandom.GetTxConcealRandomPoint()
	if err != nil {
		return false
	}
	_, err = recv.TxRandom.GetTxOTARandomPoint()
	if err != nil {
		return false
	}
	return recv.PublicKey.PointValid()
}

func (recv *OTAReceiver) FromAddress(addr key.PaymentAddress) error {
	if recv == nil {
		return fmt.Errorf("OTAReceiver not initialized")
	}

	targetShardID := common.GetShardIDFromLastByte(addr.Pk[len(addr.Pk)-1])
	otaRand := operation.RandomScalar()
	concealRand := operation.RandomScalar()

	index := uint32(0)
	publicOTA := addr.GetOTAPublicKey()
	if publicOTA == nil {
		return fmt.Errorf("missing public OTA in payment address")
	}
	publicSpend := addr.GetPublicSpend()
	rK := (&operation.Point{}).ScalarMult(publicOTA, otaRand)
	for i := MaxAttempts; i > 0; i-- {
		index++
		hash := operation.HashToScalar(append(rK.ToBytesS(), common.Uint32ToBytes(index)...))
		HrKG := (&operation.Point{}).ScalarMultBase(hash)
		publicKey := (&operation.Point{}).Add(HrKG, publicSpend)

		pkb := publicKey.ToBytesS()
		senderShardID, recvShardID, coinPrivacyType, _ := DeriveShardInfoFromCoin(pkb)
		if recvShardID == int(targetShardID) && senderShardID == int(targetShardID) && coinPrivacyType == PrivacyTypeMint {
			otaRandomPoint := (&operation.Point{}).ScalarMultBase(otaRand)
			concealRandomPoint := (&operation.Point{}).ScalarMultBase(concealRand)
			recv.PublicKey = publicKey
			recv.TxRandom = *NewTxRandom()
			recv.TxRandom.SetTxOTARandomPoint(otaRandomPoint)
			recv.TxRandom.SetTxConcealRandomPoint(concealRandomPoint)
			recv.TxRandom.SetIndex(index)
			return nil
		}
	}
	return fmt.Errorf("cannot generate OTAReceiver after %d attempts", MaxAttempts)
}

// FromString() returns a new OTAReceiver parsed from the input string,
// or error on failure
func (recv *OTAReceiver) FromString(data string) error {
	raw, _, err := base58.Base58Check{}.Decode(data)
	if err != nil {
		return err
	}
	err = recv.SetBytes(raw)
	if err != nil {
		return err
	}
	return nil
}

// String() marshals the OTAReceiver, then encodes it with base58
func (recv OTAReceiver) String() (string, error) {
	rawBytes, err := recv.Bytes()
	if err != nil {
		return "", err
	}
	return base58.Base58Check{}.NewEncode(rawBytes, common.ZeroByte), nil
}

func (recv OTAReceiver) Bytes() ([]byte, error) {
	rawBytes := []byte{byte(wallet.PrivateReceivingAddressType)}
	rawBytes = append(rawBytes, recv.PublicKey.ToBytesS()...)
	rawBytes = append(rawBytes, recv.TxRandom.Bytes()...)
	return rawBytes, nil
}

func (recv *OTAReceiver) SetBytes(b []byte) error {
	if len(b) == 0 {
		return fmt.Errorf("not enough bytes to parse ReceivingAddress")
	}
	if recv == nil {
		return fmt.Errorf("OTAReceiver not initialized")
	}
	keyType := b[0]
	switch keyType {
	case wallet.PrivateReceivingAddressType:
		buf := make([]byte, 32)
		copy(buf, b[1:33])
		pk, err := (&operation.Point{}).FromBytesS(buf)
		if err != nil {
			recv.PublicKey = nil
		} else {
			recv.PublicKey = pk
		}
		txr := NewTxRandom()
		// SetBytes() will perform length check
		err = txr.SetBytes(b[33:])
		if err != nil {
			return err
		}
		recv.TxRandom = *txr
		return nil
	default:
		return fmt.Errorf("unrecognized prefix for ReceivingAddress")
	}
}

func (recv OTAReceiver) MarshalJSON() ([]byte, error) {
	s, err := recv.String()
	if err != nil {
		return nil, err
	}
	return json.Marshal(s)
}

func (recv *OTAReceiver) UnmarshalJSON(raw []byte) error {
	var encodedString string
	err := json.Unmarshal(raw, &encodedString)
	if err != nil {
		return err
	}
	var temp OTAReceiver
	err = temp.FromString(encodedString)
	if err != nil {
		return err
	}
	*recv = temp
	return nil
}

func (recv OTAReceiver) GetShardID() byte {
	pkb := recv.PublicKey.ToBytes()
	lastByte := pkb[operation.Ed25519KeySize-1]
	shardID := common.GetShardIDFromLastByte(lastByte)
	return shardID
}
