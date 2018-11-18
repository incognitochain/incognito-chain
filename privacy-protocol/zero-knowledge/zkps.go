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
