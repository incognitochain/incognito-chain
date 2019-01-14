package privacy

// PedersenCommitment represents the parameters for the commitment
type PublicParams struct {
	G []*EllipticPoint // generators

	// 5 points: for Pedersen commitment's generators
	// G[0]: public key
	// G[1]: Value
	// G[2]: SNDerivator
	// G[3]: ShardID
	// G[4]: Randomness

	// 128 points: for multi-range proof
}


// newPedersenCommitment creates new generators
func newPublicParams() PublicParams {
	var params PublicParams
	const capacity = 133 // fixed value = 5+128
	params.G = make([]*EllipticPoint, capacity, capacity)
	params.G[0] = new(EllipticPoint)
	params.G[0].X, params.G[0].Y = Curve.Params().Gx, Curve.Params().Gy

	for i := 1; i < len(params.G); i++ {
		params.G[i] = params.G[0].Hash(i)
	}
	return params
}

var PubParams = newPublicParams()


/***** Pedersen commitment params *****/


