CreateTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64) (jsonresult.CreateTransactionResult,error)
GetBalanceByPrivateKey(privateKey string) (uint64,error)
GetRewardAmount(paymentAddress string) (error)

