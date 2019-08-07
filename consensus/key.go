package consensus

func (engine *Engine) LoadMiningKeys(keys string) error {
	return nil
}
func (engine *Engine) GetMiningPublicKey() (publickey string, keyType string) {
	return "", ""
}
func (engine *Engine) SignDataWithMiningKey(data []byte) (string, error) {
	return "", nil
}
func (engine *Engine) VerifyValidationData(data []byte, validationData string, consensusType string) error {
	return nil
}
