package blsbft

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/bridgesig"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type ProposeBlockInfo struct {
	block                   types.BlockInterface
	receiveTime             time.Time
	committees              []incognitokey.CommitteePublicKey
	signingCommittees       []incognitokey.CommitteePublicKey
	userKeySet              []signatureschemes2.MiningKey
	votes                   map[string]*BFTVote //pk->BFTVote
	isValid                 bool
	hasNewVote              bool
	isVoted                 bool
	isCommitted             bool
	validVotes              int
	errVotes                int
	proposerSendVote        bool
	proposerMiningKeyBase58 string
	lastValidateTime        time.Time

	reProposeHashSignature string
	isValidLemma2Proof     bool
	reProposeBlockInfo     ReProposeBlockInfo
	finalityProof          FinalityProof
	finalityData           []ReProposeBlockInfo
}

//NewProposeBlockInfoValue : new propose block info
func newProposeBlockForProposeMsg(
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	signingCommittes []incognitokey.CommitteePublicKey,
	userKeySet []signatureschemes2.MiningKey,
	proposerMiningKeyBase58 string,
) *ProposeBlockInfo {
	return &ProposeBlockInfo{
		block:                   block,
		votes:                   make(map[string]*BFTVote),
		committees:              incognitokey.DeepCopy(committees),
		signingCommittees:       incognitokey.DeepCopy(signingCommittes),
		userKeySet:              signatureschemes2.DeepCopyMiningKeyArray(userKeySet),
		proposerMiningKeyBase58: proposerMiningKeyBase58,
	}
}

func (proposeBlockInfo *ProposeBlockInfo) addBlockInfo(
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	signingCommittes []incognitokey.CommitteePublicKey,
	userKeySet []signatureschemes2.MiningKey,
	validVotes, errVotes int,
) {
	proposeBlockInfo.block = block
	proposeBlockInfo.committees = incognitokey.DeepCopy(committees)
	proposeBlockInfo.signingCommittees = incognitokey.DeepCopy(signingCommittes)
	proposeBlockInfo.userKeySet = signatureschemes2.DeepCopyMiningKeyArray(userKeySet)
	proposeBlockInfo.validVotes = validVotes
	proposeBlockInfo.errVotes = errVotes
}

func newBlockInfoForVoteMsg() *ProposeBlockInfo {
	return &ProposeBlockInfo{
		votes:      make(map[string]*BFTVote),
		hasNewVote: true,
	}
}

type FinalityProof struct {
	ReProposeHashSignature []string
}

func NewFinalityProof() *FinalityProof {
	return &FinalityProof{}
}

func (f *FinalityProof) AddProof(reproposeHash string) {
	f.ReProposeHashSignature = append(f.ReProposeHashSignature, reproposeHash)
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

func createReProposeHashSignature(privateKey []byte, block types.BlockInterface) (string, error) {

	reProposeBlockInfo := newReProposeBlockInfo(
		block.GetPrevHash(),
		block.GetProducer(),
		block.GetProduceTime(),
		block.GetProposer(),
		block.GetProposeTime(),
		block.GetRootHash(),
	)

	return reProposeBlockInfo.Sign(privateKey)
}

func newReProposeBlockInfo(previousBlockHash common.Hash, producer string, producerTimeSlot int64, proposer string, proposerTimeSlot int64, rootHash common.Hash) *ReProposeBlockInfo {
	return &ReProposeBlockInfo{PreviousBlockHash: previousBlockHash, Producer: producer, ProducerTimeSlot: producerTimeSlot, Proposer: proposer, ProposerTimeSlot: proposerTimeSlot, RootHash: rootHash}
}

func (r ReProposeBlockInfo) Hash() common.Hash {
	data, _ := json.Marshal(r)
	return common.HashH(data)
}

func (r ReProposeBlockInfo) Sign(privateKey []byte) (string, error) {

	hash := r.Hash()

	sig, err := bridgesig.Sign(privateKey, hash.Bytes())
	if err != nil {
		return "", err
	}

	return string(sig), nil
}

func (r ReProposeBlockInfo) VerifySignature(sig string, publicKey []byte) (bool, error) {

	hash := r.Hash()

	isValid, err := bridgesig.Verify(publicKey, hash.Bytes(), []byte(sig))
	if err != nil {
		return false, err
	}

	return isValid, nil
}
