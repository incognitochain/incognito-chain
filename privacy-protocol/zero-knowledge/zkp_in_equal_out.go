package zkp


// PKMaxValue is a protocol for Zero-knowledge Proof of Knowledge of max value is 2^64-1
// include witnesses: commitedValue, r []byte
type PKInEqualOutProtocol struct{
	//witnesses [][]byte
}

// PKOneOfManyProof contains proof's value
type PKInEqualOutProof struct {
	//commitments [][]byte
	//proofZeroOneCommitments []*PKComZeroOneProof
	//PKComZeroOneProtocol *PKComZeroOneProtocol
}

func (pro * PKInEqualOutProtocol) Prove(inComs [][]byte, outComs [][]byte) (*PKInEqualOutProof, error) {

	proof := new(PKInEqualOutProof)


	return proof, nil
}



