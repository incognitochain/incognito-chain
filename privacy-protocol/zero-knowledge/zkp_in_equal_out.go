package zkp


// PKMaxValue is a protocol for Zero-knowledge Proof of Knowledge of max value is 2^64-1
// include Witness: CommitedValue, r []byte
type PKInEqualOutProtocol struct{
	//Witness [][]byte
}

// PKOneOfManyProof contains Proof's value
type PKInEqualOutProof struct {
	//commitments [][]byte
	//proofZeroOneCommitments []*PKComZeroOneProof
	//PKComZeroOneProtocol *PKComZeroOneProtocol
}

func (pro * PKInEqualOutProtocol) Prove(inComs [][]byte, outComs [][]byte) (*PKInEqualOutProof, error) {

	proof := new(PKInEqualOutProof)


	return proof, nil
}



