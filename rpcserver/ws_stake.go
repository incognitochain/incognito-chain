package rpcserver

import (
	"errors"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/pubsub"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"reflect"
)

func (wsServer *WsServer) handleSubcribeShardCandidateByPublickey(params interface{}, subcription string, cResult chan RpcSubResult, closeChan <-chan struct{}) {
	Logger.log.Info("Handle Subcribe Shard Candidate By Pubkey", params, subcription)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		err := NewRPCError(RPCInvalidParamsError, errors.New("Methods should only contain 1 params"))
		cResult <- RpcSubResult{Error: err}
		return
	}
	candidate, ok := arrayParams[0].(string)
	if !ok {
		err := NewRPCError(RPCInvalidParamsError, errors.New("Invalid Public Key"))
		cResult <- RpcSubResult{Error: err}
	}
	// try to get candidate from beacon beststate
	candidates := wsServer.config.BlockChain.BestState.Beacon.GetShardCandidate()
	if common.IndexOfStr(candidate, candidates) > -1 {
		cResult <- RpcSubResult{Result: true, Error: nil}
		return
	}
	subId, subChan, err := wsServer.config.PubSubManager.RegisterNewSubscriber(pubsub.NewBeaconBlockTopic)
	if err != nil {
		err := NewRPCError(SubcribeError, err)
		cResult <- RpcSubResult{Error: err}
		return
	}
	defer func() {
		Logger.log.Info("Finish Subscribe Candidate Role", candidate)
		wsServer.config.PubSubManager.Unsubscribe(pubsub.NewBeaconBlockTopic, subId)
		close(cResult)
	}()
	for {
		select {
		case msg := <-subChan:
			{
				_, ok := msg.Value.(*blockchain.BeaconBlock)
				if !ok {
					Logger.log.Errorf("Wrong Message Type from Pubsub Manager, wanted *blockchain.BeaconBlock, have %+v", reflect.TypeOf(msg.Value))
					continue
				}
				// try to get candidate from beacon beststate
				candidates := wsServer.config.BlockChain.BestState.Beacon.GetShardCandidate()
				if common.IndexOfStr(candidate, candidates) > -1 {
					cResult <- RpcSubResult{Result: true, Error: nil}
					return
				} else {
					continue
				}
			}
		case <-closeChan:
			{
				cResult <- RpcSubResult{Result: jsonresult.UnsubcribeResult{Message: "Unsubscribe Shard Candidate " + candidate}}
				return
			}
		}
	}
}

func (wsServer *WsServer) handleSubcribeShardPendingValidatorByPublickey(params interface{}, subcription string, cResult chan RpcSubResult, closeChan <-chan struct{}) {
	Logger.log.Info("Handle Subcribe Shard Validator By Pubkey", params, subcription)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		err := NewRPCError(RPCInvalidParamsError, errors.New("Methods should only contain 1 params"))
		cResult <- RpcSubResult{Error: err}
		return
	}
	validator, ok := arrayParams[0].(string)
	if !ok {
		err := NewRPCError(RPCInvalidParamsError, errors.New("Invalid Public Key"))
		cResult <- RpcSubResult{Error: err}
	}
	// try to get validator from beacon beststate
	allValidators := wsServer.config.BlockChain.BestState.Beacon.GetShardPendingValidator()
	for _, shardValidators := range allValidators {
		if common.IndexOfStr(validator, shardValidators) > -1 {
			cResult <- RpcSubResult{Result: true, Error: nil}
			return
		}
	}
	subId, subChan, err := wsServer.config.PubSubManager.RegisterNewSubscriber(pubsub.NewBeaconBlockTopic)
	if err != nil {
		err := NewRPCError(SubcribeError, err)
		cResult <- RpcSubResult{Error: err}
		return
	}
	defer func() {
		Logger.log.Info("Finish Subscribe Validator Role", validator)
		wsServer.config.PubSubManager.Unsubscribe(pubsub.NewBeaconBlockTopic, subId)
		close(cResult)
	}()
	for {
		select {
		case msg := <-subChan:
			{
				_, ok := msg.Value.(*blockchain.BeaconBlock)
				if !ok {
					Logger.log.Errorf("Wrong Message Type from Pubsub Manager, wanted *blockchain.BeaconBlock, have %+v", reflect.TypeOf(msg.Value))
					continue
				}
				// try to get validator from beacon beststate
				allValidators := wsServer.config.BlockChain.BestState.Beacon.GetShardPendingValidator()
				for _, shardValidators := range allValidators {
					if common.IndexOfStr(validator, shardValidators) > -1 {
						cResult <- RpcSubResult{Result: true, Error: nil}
						return
					}
				}
			}
		case <-closeChan:
			{
				cResult <- RpcSubResult{Result: jsonresult.UnsubcribeResult{Message: "Unsubscribe Shard Validator " + validator}}
				return
			}
		}
	}
}

func (wsServer *WsServer) handleSubcribeShardCommitteeByPublickey(params interface{}, subcription string, cResult chan RpcSubResult, closeChan <-chan struct{}) {
	Logger.log.Info("Handle Subcribe Shard Committee By Pubkey", params, subcription)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		err := NewRPCError(RPCInvalidParamsError, errors.New("Methods should only contain 1 params"))
		cResult <- RpcSubResult{Error: err}
		return
	}
	committee, ok := arrayParams[0].(string)
	if !ok {
		err := NewRPCError(RPCInvalidParamsError, errors.New("Invalid Public Key"))
		cResult <- RpcSubResult{Error: err}
	}
	// try to get committee from beacon beststate
	allCommittees := wsServer.config.BlockChain.BestState.Beacon.GetShardCommittee()
	for _, shardCommittees := range allCommittees {
		if common.IndexOfStr(committee, shardCommittees) > -1 {
			cResult <- RpcSubResult{Result: true, Error: nil}
			return
		}
	}
	subId, subChan, err := wsServer.config.PubSubManager.RegisterNewSubscriber(pubsub.NewBeaconBlockTopic)
	if err != nil {
		err := NewRPCError(SubcribeError, err)
		cResult <- RpcSubResult{Error: err}
		return
	}
	defer func() {
		Logger.log.Info("Finish Subscribe Committee Role", committee)
		wsServer.config.PubSubManager.Unsubscribe(pubsub.NewBeaconBlockTopic, subId)
		close(cResult)
	}()
	for {
		select {
		case msg := <-subChan:
			{
				_, ok := msg.Value.(*blockchain.BeaconBlock)
				if !ok {
					Logger.log.Errorf("Wrong Message Type from Pubsub Manager, wanted *blockchain.BeaconBlock, have %+v", reflect.TypeOf(msg.Value))
					continue
				}
				// try to get committee from beacon beststate
				allCommittees := wsServer.config.BlockChain.BestState.Beacon.GetShardCommittee()
				for _, shardCommittees := range allCommittees {
					if common.IndexOfStr(committee, shardCommittees) > -1 {
						cResult <- RpcSubResult{Result: true, Error: nil}
						return
					}
				}
			}
		case <-closeChan:
			{
				cResult <- RpcSubResult{Result: jsonresult.UnsubcribeResult{Message: "Unsubscribe Shard Committee " + committee}}
				return
			}
		}
	}
}

func (wsServer *WsServer) handleSubcribeBeaconCandidateByPublickey(params interface{}, subcription string, cResult chan RpcSubResult, closeChan <-chan struct{}) {
	Logger.log.Info("Handle Subcribe Pending Transaction", params, subcription)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		err := NewRPCError(RPCInvalidParamsError, errors.New("Methods should only contain 1 params"))
		cResult <- RpcSubResult{Error: err}
		return
	}
	candidate, ok := arrayParams[0].(string)
	if !ok {
		err := NewRPCError(RPCInvalidParamsError, errors.New("Invalid Public Key"))
		cResult <- RpcSubResult{Error: err}
	}
	// try to get candidate from beacon beststate
	candidates := wsServer.config.BlockChain.BestState.Beacon.GetBeaconCandidate()
	if common.IndexOfStr(candidate, candidates) > -1 {
		cResult <- RpcSubResult{Result: true, Error: nil}
		return
	}
	subId, subChan, err := wsServer.config.PubSubManager.RegisterNewSubscriber(pubsub.NewBeaconBlockTopic)
	if err != nil {
		err := NewRPCError(SubcribeError, err)
		cResult <- RpcSubResult{Error: err}
		return
	}
	defer func() {
		Logger.log.Info("Finish Subscribe Candidate Role", candidate)
		wsServer.config.PubSubManager.Unsubscribe(pubsub.NewBeaconBlockTopic, subId)
		close(cResult)
	}()
	for {
		select {
		case msg := <-subChan:
			{
				_, ok := msg.Value.(*blockchain.BeaconBlock)
				if !ok {
					Logger.log.Errorf("Wrong Message Type from Pubsub Manager, wanted *blockchain.BeaconBlock, have %+v", reflect.TypeOf(msg.Value))
					continue
				}
				// try to get candidate from beacon beststate
				candidates := wsServer.config.BlockChain.BestState.Beacon.GetBeaconCandidate()
				if common.IndexOfStr(candidate, candidates) > -1 {
					cResult <- RpcSubResult{Result: true, Error: nil}
					return
				}
			}
		case <-closeChan:
			{
				cResult <- RpcSubResult{Result: jsonresult.UnsubcribeResult{Message: "Unsubscribe Beacon Candidate " + candidate}}
				return
			}
		}
	}
}

func (wsServer *WsServer) handleSubcribeBeaconPendingValidatorByPublickey(params interface{}, subcription string, cResult chan RpcSubResult, closeChan <-chan struct{}) {
	Logger.log.Info("Handle Subcribe Pending Transaction", params, subcription)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		err := NewRPCError(RPCInvalidParamsError, errors.New("Methods should only contain 1 params"))
		cResult <- RpcSubResult{Error: err}
		return
	}
	validator, ok := arrayParams[0].(string)
	if !ok {
		err := NewRPCError(RPCInvalidParamsError, errors.New("Invalid Public Key"))
		cResult <- RpcSubResult{Error: err}
	}
	// try to get validator from beacon beststate
	validators := wsServer.config.BlockChain.BestState.Beacon.GetBeaconPendingValidator()
	if common.IndexOfStr(validator, validators) > -1 {
		cResult <- RpcSubResult{Result: true, Error: nil}
		return
	}
	subId, subChan, err := wsServer.config.PubSubManager.RegisterNewSubscriber(pubsub.NewBeaconBlockTopic)
	if err != nil {
		err := NewRPCError(SubcribeError, err)
		cResult <- RpcSubResult{Error: err}
		return
	}
	defer func() {
		Logger.log.Info("Finish Subscribe Pending Validator Role", validator)
		wsServer.config.PubSubManager.Unsubscribe(pubsub.NewBeaconBlockTopic, subId)
		close(cResult)
	}()
	for {
		select {
		case msg := <-subChan:
			{
				_, ok := msg.Value.(*blockchain.BeaconBlock)
				if !ok {
					Logger.log.Errorf("Wrong Message Type from Pubsub Manager, wanted *blockchain.BeaconBlock, have %+v", reflect.TypeOf(msg.Value))
					continue
				}
				// try to get validator from beacon beststate
				validators := wsServer.config.BlockChain.BestState.Beacon.GetBeaconPendingValidator()
				if common.IndexOfStr(validator, validators) > -1 {
					cResult <- RpcSubResult{Result: true, Error: nil}
					return
				}
			}
		case <-closeChan:
			{
				cResult <- RpcSubResult{Result: jsonresult.UnsubcribeResult{Message: "Unsubscribe Beacon Validator " + validator}}
				return
			}
		}
	}
}

func (wsServer *WsServer) handleSubcribeBeaconCommitteeByPublickey(params interface{}, subcription string, cResult chan RpcSubResult, closeChan <-chan struct{}) {
	Logger.log.Info("Handle Subcribe Pending Transaction", params, subcription)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		err := NewRPCError(RPCInvalidParamsError, errors.New("Methods should only contain 1 params"))
		cResult <- RpcSubResult{Error: err}
		return
	}
	committee, ok := arrayParams[0].(string)
	if !ok {
		err := NewRPCError(RPCInvalidParamsError, errors.New("Invalid Public Key"))
		cResult <- RpcSubResult{Error: err}
	}
	// try to get candidate from beacon beststate
	committees := wsServer.config.BlockChain.BestState.Beacon.GetBeaconCommittee()
	if common.IndexOfStr(committee, committees) > -1 {
		cResult <- RpcSubResult{Result: true, Error: nil}
		return
	}
	subId, subChan, err := wsServer.config.PubSubManager.RegisterNewSubscriber(pubsub.NewBeaconBlockTopic)
	if err != nil {
		err := NewRPCError(SubcribeError, err)
		cResult <- RpcSubResult{Error: err}
		return
	}
	defer func() {
		Logger.log.Info("Finish Subscribe Committee Role", committee)
		wsServer.config.PubSubManager.Unsubscribe(pubsub.NewBeaconBlockTopic, subId)
		close(cResult)
	}()
	for {
		select {
		case msg := <-subChan:
			{
				_, ok := msg.Value.(*blockchain.BeaconBlock)
				if !ok {
					Logger.log.Errorf("Wrong Message Type from Pubsub Manager, wanted *blockchain.BeaconBlock, have %+v", reflect.TypeOf(msg.Value))
					continue
				}
				// try to get committee from beacon beststate
				committees := wsServer.config.BlockChain.BestState.Beacon.GetBeaconCommittee()
				if common.IndexOfStr(committee, committees) > -1 {
					cResult <- RpcSubResult{Result: true, Error: nil}
					return
				}
			}
		case <-closeChan:
			{
				cResult <- RpcSubResult{Result: jsonresult.UnsubcribeResult{Message: "Unsubscribe Beacon Committee " + committee}}
				return
			}
		}
	}
}
