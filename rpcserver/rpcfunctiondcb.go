package rpcserver

import (
	"encoding/hex"

	"github.com/constant-money/constant-chain/privacy"
)

func ListPaymentAddressToListString(addresses []privacy.PaymentAddress) []string {
	res := make([]string, 0)
	for _, i := range addresses {
		pk := hex.EncodeToString(i.Pk)
		res = append(res, pk)
	}
	return res
}

func (rpcServer RpcServer) handleGetConstantCirculating(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	type result struct {
		Total uint64
	}
	return result{Total: uint64(0)}, nil
}
