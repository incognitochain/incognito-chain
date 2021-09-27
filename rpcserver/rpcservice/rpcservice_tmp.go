package rpcservice

//type TmpTxInfo struct {
//	Version int8
//	Type    string
//	Inputs  map[uint64]jsonresult.ICoinInfo
//	Outputs map[uint64]jsonresult.ICoinInfo
//}
//
//// GetCoinInfoByHashes returns a list of input/output coin information for a list of hashes.
//func (txService TxService) GetCoinInfoByHashes(txHashList []string, tokenID string) (map[string]TmpTxInfo, *RPCError) {
//	res := make(map[string]TmpTxInfo)
//	for _, txHashStr := range txHashList {
//		tmpTxInfo := TmpTxInfo{}
//		if _, ok := res[txHashStr]; ok {
//			continue
//		}
//		txHash, err := common.Hash{}.NewHashFromStr(txHashStr)
//		if err != nil {
//			return nil, NewRPCError(RPCInvalidParamsError, fmt.Errorf("tx hash %v is invalid", txHashStr))
//		}
//		Logger.log.Infof("Get Transaction By Hash %+v", *txHash)
//		_, _, _, _, tx, err := txService.BlockChain.GetTransactionByHash(*txHash)
//		if err != nil {
//			return nil, NewRPCError(RPCInternalError, fmt.Errorf("tx hash %v has not been found or has not been confirmed", txHashStr))
//		}
//		tmpTxInfo.Version = tx.GetVersion()
//		tmpTxInfo.Type = tx.GetType()
//		if tokenID == common.PRVIDStr {
//
//		}
//
//
//
//
//		txBytes, err := json.Marshal(tx)
//		if err != nil {
//			return nil, NewRPCError(UnexpectedError, fmt.Errorf("cannot marshal tx %v", txHashStr))
//		}
//		res[txHashStr] = base58.Base58Check{}.Encode(txBytes, common.ZeroByte)
//	}
//	return res, nil
//}
