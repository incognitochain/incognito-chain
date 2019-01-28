package ppos

import (
	"time"

	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/wire"
)

func (self *Engine) StartSwap() {
	Logger.log.Info("Consensus engine START SWAP")

	self.cSwapSig = make(chan swapSig)
	self.cQuitSwap = make(chan struct{})
	self.cSwapChain = make(chan byte)

	for {
		select {
		case <-self.cQuitSwap:
			{
				Logger.log.Info("Consensus engine STOP SWAP")
				return
			}
		case chainID := <-self.cSwapChain:
			{
				Logger.log.Infof("Consensus engine swap %d START", chainID)

				allSigReceived := make(chan struct{})
				retryTime := 0

				committee := self.GetCommittee()

				requesterPbk := base58.Base58Check{}.Encode(self.config.ProducerKeySet.PaymentAddress.Pk, byte(0x00))

				if common.IndexOfStr(requesterPbk, committee) < 0 {
					continue
				}

				committeeCandidateList := self.config.BlockChain.GetCommitteeCandidateList()
				nextProducerPbk := ""
				for _, committeeCandidatePbk := range committeeCandidateList {
					peerIDs := self.config.Server.GetPeerIDsFromPublicKey(committeeCandidatePbk)
					if len(peerIDs) == 0 {
						continue
					}
					nextProducerPbk = committeeCandidatePbk
				}
				//if producerPbk == "" {
				//	//TODO for testing
				//	producerPbk = "1q4iCdtqb67DcNYyCE8FvMZKrDRE8KHW783VoYm5LXvds7vpsi"
				//}
				if nextProducerPbk == "" {
					continue
				}

				if common.IndexOfStr(nextProducerPbk, committee) >= 0 {
					continue
				}

				signatureMap := make(map[string]string)
				lockTime := time.Now().Unix()
				reqSigMsg, _ := wire.MakeEmptyMessage(wire.CmdSwapRequest)
				reqSigMsg.(*wire.MessageSwapRequest).LockTime = lockTime
				reqSigMsg.(*wire.MessageSwapRequest).Requester = requesterPbk
				reqSigMsg.(*wire.MessageSwapRequest).ChainID = chainID
				reqSigMsg.(*wire.MessageSwapRequest).Candidate = nextProducerPbk
			BeginSwap:
			// Collect signatures of other validators
				cancel := make(chan struct{})
				go func() {
					for {
						select {
						case <-cancel:
							return
						case swapSig := <-self.cSwapSig:
							if common.IndexOfStr(swapSig.Validator, committee) >= 0 && swapSig.Validator == requesterPbk {
								// verify signature
								rawBytes := reqSigMsg.(*wire.MessageSwapRequest).GetMsgByte()
								err := cashec.ValidateDataB58(swapSig.Validator, swapSig.SwapSig, rawBytes)
								if err != nil {
									continue
								}
								Logger.log.Info("SWAP validate signature ok from ", swapSig.Validator, nextProducerPbk)
								signatureMap[swapSig.Validator] = swapSig.SwapSig
								if len(signatureMap) >= common.TotalValidators/2 {
									close(allSigReceived)
									return
								}
							}
						case <-time.After(common.MaxBlockSigWaitTime * time.Second * 5):
							return
						}
					}
				}()

				// Request signatures from other validators
				go func() {
					sigStr, err := self.signData(reqSigMsg.(*wire.MessageSwapRequest).GetMsgByte())
					if err != nil {
						Logger.log.Infof("Request swap sign error", err)
						return
					}
					reqSigMsg.(*wire.MessageSwapRequest).RequesterSig = sigStr

					for idx := 0; idx < common.TotalValidators; idx++ {
						if committee[idx] != requesterPbk {
							go func(validator string) {
								peerIDs := self.config.Server.GetPeerIDsFromPublicKey(validator)
								if len(peerIDs) > 0 {
									for _, peerID := range peerIDs {
										Logger.log.Infof("Request swap to %s %s", peerID, validator)
										self.config.Server.PushMessageToPeer(reqSigMsg, peerID)
									}
								} else {
									Logger.log.Error("Validator's peer not found!", validator)
								}
							}(committee[idx])
						}
					}
				}()

				// Wait for signatures of other validators
				select {
				case <-allSigReceived:
					Logger.log.Info("Validator signatures: ", signatureMap)
				case <-time.After(common.MaxBlockSigWaitTime * time.Second):

					close(cancel)
					if retryTime == 5 {
						continue
					}
					retryTime++
					Logger.log.Infof("Start finalizing swap... %d time", retryTime)
					goto BeginSwap
				}

				Logger.log.Infof("SWAP DONE")

				committeeV := make([]string, common.TotalValidators)
				copy(committeeV, self.GetCommittee())

				err := self.updateCommittee(nextProducerPbk, chainID)
				if err != nil {
					Logger.log.Errorf("Consensus update committee is error", err)
					continue
				}

				// broadcast message for update new committee list
				swapUpdMsg, _ := wire.MakeEmptyMessage(wire.CmdSwapUpdate)
				swapUpdMsg.(*wire.MessageSwapUpdate).LockTime = lockTime
				swapUpdMsg.(*wire.MessageSwapUpdate).Requester = requesterPbk
				swapUpdMsg.(*wire.MessageSwapUpdate).ChainID = chainID
				swapUpdMsg.(*wire.MessageSwapUpdate).Candidate = nextProducerPbk
				swapUpdMsg.(*wire.MessageSwapUpdate).Signatures = signatureMap

				self.config.Server.PushMessageToAll(reqSigMsg)

				Logger.log.Infof("Consensus engine swap %d END", chainID)
				continue
			}
		}
	}
}
