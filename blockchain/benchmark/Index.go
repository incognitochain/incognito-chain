package benchmark

func GetInitTransaction() []string{
	Shard0 := []string{}
	Shard1 := []string{}
	Shard0 = append(Shard0, InitTxsShard0...)
	//Shard0 = append(Shard0, InitTxsShard0_1...)
	//Shard0 = append(Shard0, InitTxsShard0_2...)
	//Shard0 = append(Shard0, InitTxsShard0_3...)
	//Shard0 = append(Shard0, InitTxsShard0_4...)
	//Shard0 = append(Shard0, InitTxsShard0_5...)
	//Shard0 = append(Shard0, InitTxsShard0_6...)
	//Shard0 = append(Shard0, InitTxsShard0_7...)
	//Shard0 = append(Shard0, InitTxsShard0_8...)
	//Shard0 = append(Shard0, InitTxsShard0_9...)
	Shard1 = append(Shard1, InitTxsShard1...)
	//Shard1 = append(Shard1, InitTxsShard1_1...)
	//Shard1 = append(Shard1, InitTxsShard1_2...)
	//Shard1 = append(Shard1, InitTxsShard1_3...)
	//Shard1 = append(Shard1, InitTxsShard1_4...)
	//Shard1 = append(Shard1, InitTxsShard1_5...)
	//Shard1 = append(Shard1, InitTxsShard1_6...)
	//Shard1 = append(Shard1, InitTxsShard1_7...)
	//Shard1 = append(Shard1, InitTxsShard1_8...)
	//Shard1 = append(Shard1, InitTxsShard1_9...)
	return append(Shard0, Shard1...)
}

