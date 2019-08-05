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
func (engine *Engine) VerifyDataWithMiningKey(data []byte, sig string, publicKey string, publicKeyType string) error {
	return nil
}
