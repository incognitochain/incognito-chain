package voting

//type GOVProposalData struct {
//	GOVParams       params.GOVParams
//	ExecuteDuration int32
//	Explaination    string
//}
//
//type DCBProposalData struct {
//	DCBParams       params.DCBParams
//	ExecuteDuration int32
//	Explaination    string
//}
//
//func (DCBProposalData DCBProposalData) Validate() bool {
//	return DCBProposalData.DCBParams.Validate() && ValidateExplaination(DCBProposalData.Explaination)
//}
//
//func (GOVProposalData GOVProposalData) Validate() bool {
//	return GOVProposalData.GOVParams.Validate() && ValidateExplaination(GOVProposalData.Explaination)
//}
//
//func ValidateExplaination(explaination string) bool {
//	return true
//}
//
//func (data DCBProposalData) Hash() *common.Hash {
//	record := string(common.ToBytes(data.DCBParams.Hash()))
//	record += data.Explaination
//	hash := common.DoubleHashH([]byte(record))
//	return &hash
//}
//
//func (GOVProposalData GOVProposalData) Hash() *common.Hash {
//	record := string(common.ToBytes(GOVProposalData.GOVParams.Hash()))
//	record += GOVProposalData.Explaination
//	hash := common.DoubleHashH([]byte(record))
//	return &hash
//}
