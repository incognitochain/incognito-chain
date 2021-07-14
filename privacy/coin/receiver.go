package coin

import (
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/wallet"
)

type OTAReceiver struct {
	PublicKey 	operation.Point
	TxRandom 	TxRandom
}

func ParseReceiver(data string) (*OTAReceiver, error) {
	result := &OTAReceiver{}
	raw, _, err := base58.Base58Check{}.Decode(data)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, errors.New("Not enough bytes to parse ReceivingAddress")
	}
	keyType := raw[0]
	switch keyType {
	case wallet.PrivateReceivingAddressType:
		buf := make([]byte, 32)
		copy(buf, raw[1:33])
		pk, err := (&operation.Point{}).FromBytesS(buf)
		if err != nil {
			return nil, err
		}
		result.PublicKey = *pk
		txr := NewTxRandom()
		err = txr.SetBytes(raw[33:])
		if err != nil {
			return nil, err
		}
		result.TxRandom = *txr
		return result, nil
	default:
		return nil, errors.New("Unrecognized prefix for ReceivingAddress")
	}
}

func (recv OTAReceiver) String() (string, error) {
	rawBytes := []byte{byte(wallet.PrivateReceivingAddressType)}
	rawBytes = append(rawBytes, recv.PublicKey.ToBytesS()...)
	rawBytes = append(rawBytes, recv.TxRandom.Bytes()...)
	return base58.Base58Check{}.NewEncode(rawBytes, common.ZeroByte), nil
}
