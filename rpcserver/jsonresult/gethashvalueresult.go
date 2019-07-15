package jsonresult

type HashValueDetail struct {
	IsBlock       bool `json:"IsBlock"`
	IsBeaconBlock bool `json:"IsBeaconBlock"`
	IsTransaction bool `json:"IsTransaction"`
}
