package ppos

import (
	"time"

	"github.com/ninjadotorg/cash/cashec"
	"github.com/ninjadotorg/cash/common"
	"github.com/ninjadotorg/cash/common/base58"
	"github.com/ninjadotorg/cash/wire"
)

func (self *Engine) StartSwap() error {
	Logger.log.Info("Consensus engine START SWAP")

	self.cSwapSig = make(chan swapSig)
	self.cQuitSwap = make(chan struct{})
	self.cSwapChain = make(chan byte)

	go func() {
		for {
			select {
			case <-time.After(10 * time.Second):
				Logger.log.Info("Consensus engine SWAP TIMER")
				self.cSwapChain <- byte(10)
				continue
			}
		}
	}()

	for {
		select {
		case <-self.cQuitSwap:
			{
				Logger.log.Info("Consensus engine STOP SWAP")
				return nil
			}
		case chainID := <-self.cSwapChain:
			{
				Logger.log.Infof("Consensus engine swap %d START", chainID)

				allSigReceived := make(chan struct{})
				retryTime := 0

				committee := self.GetCommittee()

				requesterPbk := base58.Base58Check{}.Encode(self.config.ValidatorKeySet.SpublicKey, byte(0x00))

				if common.IndexOfStr(requesterPbk, committee) < 0 {
					continue
				}

				committeeCandidateList := self.config.BlockChain.GetCommitteeCandidateList()
				sealerPbk := common.EmptyString
				for _, committeeCandidatePbk := range committeeCandidateList {
					peerIDs := self.config.Server.GetPeerIDsFromPublicKey(committeeCandidatePbk)
					if len(peerIDs) == 0 {
						continue
					}
					sealerPbk = committeeCandidatePbk
				}
				if sealerPbk == common.EmptyString {
					//TODO for testing
					sealerPbk = "1q4iCdtqb67DcNYyCE8FvMZKrDRE8KHW783VoYm5LXvds7vpsi"
				}
				if sealerPbk == common.EmptyString {
					continue
				}

				if common.IndexOfStr(sealerPbk, committee) >= 0 {
					continue
				}

				signatureMap := make(map[string]string)
				lockTime := time.Now().Unix()
				reqSigMsg, _ := wire.MakeEmptyMessage(wire.CmdSwapRequest)
				reqSigMsg.(*wire.MessageSwapRequest).LockTime = lockTime
				reqSigMsg.(*wire.MessageSwapRequest).Requester = requesterPbk
				reqSigMsg.(*wire.MessageSwapRequest).ChainID = chainID
				reqSigMsg.(*wire.MessageSwapRequest).Candidate = sealerPbk
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
								Logger.log.Info("SWAP validate signature ok from ", swapSig.Validator, sealerPbk)
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
					//blocksig wait time exceeded -> get a new committee list and retry
					//Logger.log.Error(errExceedSigWaitTime)

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

				err := self.updateCommittee(sealerPbk, chainID)
				if err != nil {
					Logger.log.Errorf("Consensus update committee is error", err)
					continue
				}

				// broadcast message for update new committee list
				swapUpdMsg, _ := wire.MakeEmptyMessage(wire.CmdSwapUpdate)
				swapUpdMsg.(*wire.MessageSwapUpdate).LockTime = lockTime
				swapUpdMsg.(*wire.MessageSwapUpdate).Requester = requesterPbk
				swapUpdMsg.(*wire.MessageSwapUpdate).ChainID = chainID
				swapUpdMsg.(*wire.MessageSwapUpdate).Candidate = sealerPbk
				swapUpdMsg.(*wire.MessageSwapUpdate).Signatures = signatureMap

				self.config.Server.PushMessageToAll(reqSigMsg)

				Logger.log.Infof("Consensus engine swap %d END", chainID)
				continue
			}
		}
	}
	return nil
}
