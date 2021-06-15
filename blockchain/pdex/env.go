package pdex

type StateEnvBuilder interface {
	Build() StateEnvironment
}

func NewStateEnvBuilder() StateEnvBuilder {
	return &stateEnvironment{}
}

type StateEnvironment interface {
}

type stateEnvironment struct {
}

func (env *stateEnvironment) Build() StateEnvironment {
	return env
}
