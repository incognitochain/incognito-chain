package zkp

//func TestPKOneOfMany1(t *testing.T) {
//	witness := new(PKOneOfManyWitness)
//
//	indexIsZero := 23
//
//	// list of commitments
//	commitments := make([]*privacy.EllipticPoint, 32)
//	SNDerivators := make([]*big.Int, 32)
//	randoms := make([]*big.Int, 32)
//
//	for i := 0; i < 32; i++ {
//		SNDerivators[i] = privacy.RandInt()
//		randoms[i] = privacy.RandInt()
//		commitments[i] = privacy.PedCom.CommitAtIndex(SNDerivators[i], randoms[i], privacy.SND)
//	}
//
//	// create Commitment to zero at indexIsZero
//	SNDerivators[indexIsZero] = big.NewInt(0)
//	commitments[indexIsZero] = privacy.PedCom.CommitAtIndex(SNDerivators[indexIsZero], randoms[indexIsZero], privacy.SND)
//
//	witness.Set(commitments, nil, randoms[indexIsZero], uint64(indexIsZero), privacy.SND)
//	//start := time.Now()
//	proof, err := witness.Prove()
//
//	if err != nil {
//		fmt.Println(err)
//	}
//	res := proof.Verify()
//
//	//end := time.Now()
//	//fmt.Printf("%v_+_\n", end.Sub(start))
//	fmt.Println(res)
//	//return res
//}
