package consensustypes

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/bridgesig"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type FinalityProof struct {
	ReProposeHashSignature []string
}

func NewFinalityProof() *FinalityProof {
	return &FinalityProof{}
}

func (f *FinalityProof) AddProof(reProposeHashSig string) {
	f.ReProposeHashSignature = append(f.ReProposeHashSignature, reProposeHashSig)
}

func (f *FinalityProof) GetProofByIndex(index int) (string, error) {
	if index < 0 || index >= len(f.ReProposeHashSignature) {
		return "", fmt.Errorf("Proof index %+v, is not valid. Number of Proof %+v", index, len(f.ReProposeHashSignature))
	}
	proof := f.ReProposeHashSignature[index]
	if proof == "" {
		return "", fmt.Errorf("invalid proof zero length")
	}
	return f.ReProposeHashSignature[index], nil
}

func (f *FinalityProof) Verify(
	previousBlockHash common.Hash,
	producer string,
	beginTimeSlot int64,
	proposers []string,
	rootHash common.Hash,
) error {

	for i := 0; i < len(f.ReProposeHashSignature); i++ {
		reProposer := proposers[i]
		reProposeTimeSlot := beginTimeSlot + int64(i)
		sig := f.ReProposeHashSignature[i]

		isValid, err := VerifyReProposeHashSignature(
			sig,
			previousBlockHash,
			producer,
			beginTimeSlot,
			reProposer,
			reProposeTimeSlot,
			rootHash,
		)
		if err != nil {
			return fmt.Errorf("verification failed verifyFinalityProof "+
				"Re-ProposeTimeSlot %+v, ReProposer %+v, error %+v",
				reProposeTimeSlot, reProposer, err)
		}
		if !isValid {
			return fmt.Errorf("invalid Signature verifyFinalityProof "+
				"Re-ProposeTimeSlot %+v, ReProposer %+v", reProposeTimeSlot, reProposer)
		}
	}

	return nil
}

//previousblockhash, producerTimeslot, Producer, proposerTimeslot, Proposer roothash
type ReProposeBlockInfo struct {
	PreviousBlockHash common.Hash
	Producer          string
	ProducerTimeSlot  int64
	Proposer          string
	ProposerTimeSlot  int64
	RootHash          common.Hash
}

func CreateReProposeHashSignature(privateKey []byte, block types.BlockInterface) (string, error) {

	reProposeBlockInfo := NewReProposeBlockInfo(
		block.GetPrevHash(),
		block.GetProducer(),
		common.CalculateTimeSlot(block.GetProduceTime()),
		block.GetProposer(),
		common.CalculateTimeSlot(block.GetProposeTime()),
		block.GetAggregateRootHash(),
	)

	return reProposeBlockInfo.Sign(privateKey)
}

func VerifyReProposeHashSignature(
	sig string,
	previousBlockHash common.Hash,
	producerBase58 string,
	producerTimeSlot int64,
	proposerBase58 string,
	proposerTimeSlot int64,
	rootHash common.Hash,
) (bool, error) {

	proposer := incognitokey.CommitteePublicKey{}

	_ = proposer.FromString(proposerBase58)
	publicKey := proposer.MiningPubKey[common.BridgeConsensus]

	reProposeBlockInfo := NewReProposeBlockInfo(
		previousBlockHash,
		producerBase58,
		producerTimeSlot,
		proposerBase58,
		proposerTimeSlot,
		rootHash,
	)

	return reProposeBlockInfo.VerifySignature(sig, publicKey)
}

func VerifyReProposeHashSignatureFromBlock(sig string, block types.BlockInterface) (bool, error) {
	return VerifyReProposeHashSignature(
		sig,
		block.GetPrevHash(),
		block.GetProducer(),
		common.CalculateTimeSlot(block.GetProduceTime()),
		block.GetProposer(),
		common.CalculateTimeSlot(block.GetProposeTime()),
		block.GetAggregateRootHash(),
	)
}

func NewReProposeBlockInfo(previousBlockHash common.Hash, producer string, producerTimeSlot int64, proposer string, proposerTimeSlot int64, rootHash common.Hash) *ReProposeBlockInfo {
	return &ReProposeBlockInfo{PreviousBlockHash: previousBlockHash, Producer: producer, ProducerTimeSlot: producerTimeSlot, Proposer: proposer, ProposerTimeSlot: proposerTimeSlot, RootHash: rootHash}
}

func (r ReProposeBlockInfo) Hash() common.Hash {
	data, _ := json.Marshal(&r)
	return common.HashH(data)
}

func (r ReProposeBlockInfo) Sign(privateKey []byte) (string, error) {

	hash := r.Hash()

	sig, err := bridgesig.Sign(privateKey, hash.Bytes())
	if err != nil {
		return "", err
	}

	sigBase58 := base58.Base58Check{}.Encode(sig, common.Base58Version)

	return sigBase58, nil
}

func (r ReProposeBlockInfo) VerifySignature(sigBase58 string, publicKey []byte) (bool, error) {

	hash := r.Hash()
	sig, _, _ := base58.Base58Check{}.Decode(sigBase58)

	isValid, err := bridgesig.Verify(publicKey, hash.Bytes(), sig)
	if err != nil {
		return false, err
	}

	return isValid, nil
}
