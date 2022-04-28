package multiview

type BeaconMultiView struct {
	*multiView
}

func NewBeaconMultiView() *BeaconMultiView {
	bv := &BeaconMultiView{}
	bv.multiView = NewMultiView()
	return bv
}

func (s *BeaconMultiView) AddView(v View) (int, error) {
	added := s.multiView.addView(v)
	err := s.FinalizeView(*s.GetExpectedFinalView().GetHash())
	res := 0
	if added {
		res = 1
	}
	return res, err
}
