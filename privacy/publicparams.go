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

func setPedersenParams() PedersenCommitment {
	var pcm PedersenCommitment
	const capacity = 5 // fixed value
	pcm.G = make([]*EllipticPoint, capacity, capacity)

	for i := 0; i < capacity; i++ {
		pcm.G[i] = new(EllipticPoint)
		pcm.G[i].Set(PubParams.G[i].X, PubParams.G[i].Y)
	}
	return pcm
}

var PedCom = setPedersenParams()

/***** Bullet proof params *****/

// BulletproofParams includes all generator for aggregated range proof
type BulletproofParams struct {
	G []*EllipticPoint
	H []*EllipticPoint
}

func setBulletproofParams() BulletproofParams {
	var gen BulletproofParams
	const capacity = 64 // fixed value
	gen.G = make([]*EllipticPoint, capacity, capacity)
	gen.H = make([]*EllipticPoint, capacity, capacity)

	for i := 0; i < capacity; i++ {
		gen.G[i] = new(EllipticPoint)
		gen.G[i].Set(PubParams.G[5 + i].X, PubParams.G[5 + i].Y)

		gen.H[i] = new(EllipticPoint)
		gen.H[i].Set(PubParams.G[69 + i].X, PubParams.G[69 + i].Y)
	}
	return gen
}

var AggParam = setBulletproofParams()

