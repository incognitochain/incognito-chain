package pedersen

import "github.com/incognitochain/incognito-chain/privacy/operation"

const (
	PedersenPrivateKeyIndex = byte(0x00)
	PedersenValueIndex      = byte(0x01)
	PedersenSndIndex        = byte(0x02)
	PedersenShardIDIndex    = byte(0x03)
	PedersenRandomnessIndex = byte(0x04)
)

type Point = operation.Point
type Scalar = operation.Scalar

var pedCom = NewPedersenParams()
