package constantpos

type Layershard struct {
	cQuit                 chan struct{}
	Committee             CommitteeStruct
	CurrentShard          byte
	started               bool
	protocol              *BFTProtocol
	knownChainsHeight     []int
	validatedChainsHeight []int
}

func (self *Layershard) Start() {

}
func (self *Layershard) Stop() {

}
