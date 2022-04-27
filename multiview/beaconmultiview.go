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
	s.multiView.addView(v)
	s.FinalizeView(*s.GetExpectedFinalView().GetHash())
	return 0, nil
}
