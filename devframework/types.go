package devframework

type Execute struct {
	sim          *SimulationEngine
	appliedChain []int
}

func (exec *Execute) GenerateBlock(args ...interface{}) {
	args = append(args, exec)
	exec.sim.GenerateBlock(args...)
}

func (sim *SimulationEngine) ApplyChain(chain_array ...int) *Execute {
	return &Execute{
		sim,
		chain_array,
	}
}
