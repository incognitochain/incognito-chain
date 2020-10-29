package devframework

func (sim *SimulationEngine) API_CreateTransaction(args ...interface{}) (string, error) {
	var sender string
	var receivers = make(map[string]interface{})
	for i, arg := range args {
		if i == 0 {
			sender = arg.(Account).PrivateKey
		} else {
			switch arg.(type) {
			default:
				if i%2 == 1 {
					amount, ok := args[i+1].(int)
					if !ok {
						amountF64 := args[i+1].(float64)
						amount = int(amountF64)
					}
					receivers[arg.(Account).PaymentAddress] = float64(amount)
				}
			}
		}
	}

	res, err := sim.rpc_createtransaction(sender, receivers, 1, 1)
	if err != nil {
		return "", err
	}
	sim.InjectTx(res.Base58CheckData)
	return res.TxID, nil
}
