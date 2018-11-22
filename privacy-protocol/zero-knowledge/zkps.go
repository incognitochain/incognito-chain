package zkp

type Proof interface {
}

type Witness interface {
}

// ZKProtocols interface
type ZKProtocols interface {
	SetWitness(witness Witness)
	Prove() (Proof, error)
	SetProof(proof Proof)
	Verify() bool
}


func Prove(){
	// CommitAll each component of coins being spent


	// Summing all commitments into one commitment and proving the knowledge of its openings
	// Proving one-out-of-N commitments is a commitment to the coins being spent
	// Proving that serial number is derived from the committed derivator
	// Proving that output values do not exceed v_max
	// Proving that sum of inputs equals sum of outputs
	// Proving ciphertexts encrypting for coins' details are well-formed
}
