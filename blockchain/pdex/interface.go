package pdex

type State interface {
	Version() uint
	Clone() State
	Update(StateEnvironment) ([][]string, error)
	Upgrade(StateEnvironment) State
}
