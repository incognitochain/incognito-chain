package jsonresult

type ListLoanResponseApproved struct {
	Approvers map[string][]string `json:"Approvers"`
	Approved  map[string]bool     `json:"Approved"`
}

type ListLoanResponseRejected struct {
	Rejectors map[string][]string `json:"Rejectors"`
	Rejected  map[string]bool     `json:"Rejected"`
}

type LoanPaymentInfo struct {
	Principle uint64 `json:"Principle"`
	Interest  uint64 `json:"Interest"`
	Deadline  uint64 `json:"Deadline"`
}

type ListLoanPaymentInfo struct {
	Info map[string]LoanPaymentInfo `json:"Info"`
}
