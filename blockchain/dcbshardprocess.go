package blockchain

func (bc *BlockChain) StoreMetadataInstructions(inst []string, shardID byte) error {
	if len(inst) < 2 {
		return nil // Not error, just not stability instruction
	}
	switch inst[0] {
	default:
		return nil
	}
	return nil
}
