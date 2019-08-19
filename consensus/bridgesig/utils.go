package bridgesig

import (
	"errors"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
)

func DecodeECDSASig(sigStr string) (
	v byte,
	r string,
	s string,
	err error,
) {
	sig, ver, errDecode := base58.Base58Check{}.Decode(sigStr)
	if (len(sig) != CBridgeSigSz) || (ver != common.ZeroByte) || (errDecode != nil) {
		err = errors.New("Wrong input")
		return
	}
	v = byte(sig[64] + 27)
	r = hexutil.Encode(sig[:32])
	s = hexutil.Encode(sig[32:64])
	return
}
