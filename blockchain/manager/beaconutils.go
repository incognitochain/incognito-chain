package manager

import "github.com/incognitochain/incognito-chain/instruction"

func buildAssignInstructionFromRandom(
	rI *instruction.RandomInstruction,
	bcE BeaconCommitteeEngine,
) []instruction.Instruction {
	res := []instruction.Instruction{}
	bc, shards := bcE.AssignCommitteeUsingRandomInstruction(rI.BtcNonce)
	aBC := instruction.NewAssignInstructionWithValue(instruction.BEACON_CHAIN_ID, bc)
	res = append(res, aBC)
	for sID, s := range shards {
		aS := instruction.NewAssignInstructionWithValue(int(sID), s)
		res = append(res, aS)
	}
	return res
}
