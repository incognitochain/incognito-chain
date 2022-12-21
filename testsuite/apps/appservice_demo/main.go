package main

import (
	"flag"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	devframework "github.com/incognitochain/incognito-chain/testsuite"
	"github.com/incognitochain/incognito-chain/testsuite/account"
	"github.com/incognitochain/incognito-chain/wallet"
	"time"
)

func StakeShard(client *devframework.RemoteRPCClient, privateKey string, miningPrivateKey string, stakeAmount uint64) (*jsonresult.CreateTransactionResult, error) {
	burnAddr := "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA"
	wl, _ := wallet.Base58CheckDeserialize(privateKey)
	RewardAddr := wl.Base58CheckSerialize(wallet.PaymentAddressType)
	funderPayment := wl.Base58CheckSerialize(wallet.PaymentAddressType)

	miningWl, _ := wallet.Base58CheckDeserialize(miningPrivateKey)
	privateSeedBytes := common.HashB(common.HashB(miningWl.KeySet.PrivateKey))
	privateSeed := base58.Base58Check{}.Encode(privateSeedBytes, common.Base58Version)

	txResp, err := client.CreateAndSendStakingTransaction(privateKey, map[string]interface{}{burnAddr: float64(stakeAmount)}, 1, 0, map[string]interface{}{
		"StakingType":                  float64(63),
		"CandidatePaymentAddress":      funderPayment,
		"PrivateSeed":                  privateSeed,
		"RewardReceiverPaymentAddress": RewardAddr,
		"AutoReStaking":                true,
	})
	return &txResp, err
}

func StakeBeacon(client *devframework.RemoteRPCClient, privateKey string, miningPrivateKey string, stakeAmount uint64) (*jsonresult.CreateTransactionResult, error) {
	burnAddr := "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA"
	wl, _ := wallet.Base58CheckDeserialize(privateKey)
	RewardAddr := wl.Base58CheckSerialize(wallet.PaymentAddressType)
	funderPayment := wl.Base58CheckSerialize(wallet.PaymentAddressType)

	miningWl, _ := wallet.Base58CheckDeserialize(miningPrivateKey)
	privateSeedBytes := common.HashB(common.HashB(miningWl.KeySet.PrivateKey))
	privateSeed := base58.Base58Check{}.Encode(privateSeedBytes, common.Base58Version)

	txResp, err := client.CreateAndSendStakingTransaction(privateKey, map[string]interface{}{burnAddr: float64(stakeAmount)}, 1, 0, map[string]interface{}{
		"StakingType":                  float64(64),
		"CandidatePaymentAddress":      funderPayment,
		"PrivateSeed":                  privateSeed,
		"RewardReceiverPaymentAddress": RewardAddr,
		"AutoReStaking":                true,
	})
	return &txResp, err
}

func AddStake(client *devframework.RemoteRPCClient, privateKey string, miningPrivateKey string, stakeAmount uint64) (*jsonresult.CreateTransactionResult, error) {
	burnAddr := "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA"
	wl, _ := wallet.Base58CheckDeserialize(privateKey)
	funderPayment := wl.Base58CheckSerialize(wallet.PaymentAddressType)

	miningWl, _ := wallet.Base58CheckDeserialize(miningPrivateKey)
	privateSeedBytes := common.HashB(common.HashB(miningWl.KeySet.PrivateKey))
	privateSeed := base58.Base58Check{}.Encode(privateSeedBytes, common.Base58Version)

	txResp, err := client.AddStake(privateKey, map[string]interface{}{burnAddr: float64(stakeAmount)}, -1, -1, map[string]interface{}{
		"CandidatePaymentAddress": funderPayment,
		"PrivateSeed":             privateSeed,
		"AddStakingAmount":        stakeAmount,
	})
	return &txResp, err
}

func UnStake(client *devframework.RemoteRPCClient, privateKey string, miningPrivateKey string) (*jsonresult.CreateTransactionResult, error) {
	burnAddr := "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA"
	wl, _ := wallet.Base58CheckDeserialize(privateKey)
	funderPayment := wl.Base58CheckSerialize(wallet.PaymentAddressType)

	miningWl, _ := wallet.Base58CheckDeserialize(miningPrivateKey)
	privateSeedBytes := common.HashB(common.HashB(miningWl.KeySet.PrivateKey))
	privateSeed := base58.Base58Check{}.Encode(privateSeedBytes, common.Base58Version)

	txResp, err := client.Unstake(privateKey, map[string]interface{}{burnAddr: float64(0)}, 1, -1, map[string]interface{}{
		"CandidatePaymentAddress": funderPayment,
		"PrivateSeed":             privateSeed,
		"UnStakingType":           210,
	})
	return &txResp, err
}

func SendPRV(fullnodeRPC devframework.RemoteRPCClient, args ...interface{}) (string, error) {
	var sender string
	var receivers = make(map[string]uint64)
	for i, arg := range args {
		if i == 0 {
			sender = arg.(string)
		} else {
			switch arg.(type) {
			default:
				if i%2 == 1 {
					amount, ok := args[i+1].(float64)
					if !ok {
						amountF64 := args[i+1].(float64)
						amount = amountF64
					}
					receivers[arg.(string)] = uint64(amount)
				}
			}
		}
	}
	res, err := fullnodeRPC.CreateAndSendTransaction(sender, mapUintToInterface(receivers), -1, 1)
	if err != nil {
		return "", err
	}
	return res.TxID, nil
}
func mapUintToInterface(m map[string]uint64) map[string]interface{} {
	mfl := make(map[string]interface{})
	for a, v := range m {
		mfl[a] = float64(v)
	}
	return mfl
}

func CreateAccounts(rpc *devframework.RemoteRPCClient, seed string, size int) []account.Account {
	shard0 := make([]account.Account, size)
	semaphore := make(chan int, 50)
	for i := 0; i < size; i++ {
		semaphore <- 1
		go func(i int) {
			acc, _ := account.GenerateAccountByShard(0, 0, fmt.Sprintf("%v%v", seed, i))
			shard0[i] = acc
			if rpc != nil {
				rpc.SubmitKey(acc.PrivateKey)
				time.Sleep(time.Millisecond * 10)
			}
			<-semaphore
		}(i)
	}
	for {
		if len(semaphore) == 0 {
			break
		}
		time.Sleep(time.Second)
	}
	return shard0[:]
}

func main() {
	flag.Parse()
	fullnodeRPC := devframework.RemoteRPCClient{"http://127.0.0.1:30001"}
	shard0RPC := devframework.RemoteRPCClient{"http://127.0.0.1:20004"}
	var icoAccount, _ = account.NewAccountFromPrivatekey("112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or")
	stakers := CreateAccounts(&shard0RPC, "seed", 20)
	beaconStaker := CreateAccounts(&shard0RPC, "xxxx", 4)

	for i, staker := range stakers {
		fmt.Println("Shard", i, staker.MiningKey)
	}

	for i, staker := range beaconStaker {
		fmt.Println("Beacon", i, staker.SelfCommitteePubkey, staker.MiningKey)
	}

	submitKey := func() {
		fullnodeRPC.SubmitKey(icoAccount.PrivateKey)
		shard0RPC.SubmitKey(icoAccount.PrivateKey)
		shard0RPC.CreateConvertCoinVer1ToVer2Transaction(icoAccount.PrivateKey)

	}

	send_prv := func() {
		receiver := []interface{}{icoAccount.PrivateKey}
		for _, bstaker := range beaconStaker {
			receiver = append(receiver, bstaker.PaymentAddress)
			receiver = append(receiver, float64(100000*1e10))
		}
		for _, staker := range stakers {
			receiver = append(receiver, staker.PaymentAddress)
			receiver = append(receiver, float64(2000*1e9))
		}

		tx, err := SendPRV(shard0RPC, receiver...)
		fmt.Println(tx, err)
	}

	stake_shard := func() {
		for i, staker := range stakers {
			fmt.Println("Stake shard", i)
			stakeRes, err := StakeShard(&shard0RPC, staker.PrivateKey, staker.PrivateKey, 1750*1e9)
			fmt.Println(stakeRes, err)
		}
		fmt.Println("Stake shard", beaconStaker[0].SelfCommitteePubkey)
		beaconStakerBalance, _ := shard0RPC.GetBalanceByPrivateKey(beaconStaker[0].PrivateKey)

		fmt.Println("BeaconStaker", int(float64(beaconStakerBalance)*1e-9))
		stakeRes, err := StakeShard(&shard0RPC, beaconStaker[0].PrivateKey, beaconStaker[0].PrivateKey, 1750*1e9)
		fmt.Println(stakeRes, err)

		fmt.Println("Stake shard", beaconStaker[1].SelfCommitteePubkey)
		beaconStakerBalance, _ = shard0RPC.GetBalanceByPrivateKey(beaconStaker[1].PrivateKey)
		fmt.Println("BeaconStaker", float64(beaconStakerBalance)*1e-9)
		stakeRes, err = StakeShard(&shard0RPC, beaconStaker[1].PrivateKey, beaconStaker[1].PrivateKey, 1750*1e9)
		fmt.Println(stakeRes, err)

		fmt.Println("Stake shard", beaconStaker[2].SelfCommitteePubkey)
		stakeRes, err = StakeShard(&shard0RPC, icoAccount.PrivateKey, beaconStaker[2].PrivateKey, 1750*1e9)
		fmt.Println(stakeRes, err)
		time.Sleep(time.Second * 10)

		fmt.Println("Stake shard", beaconStaker[3].SelfCommitteePubkey)
		stakeRes, err = StakeShard(&shard0RPC, icoAccount.PrivateKey, beaconStaker[3].PrivateKey, 1750*1e9)
		fmt.Println(stakeRes, err)
	}

	stake_beacon := func() {
		stakeRes, err := StakeBeacon(&shard0RPC, beaconStaker[0].PrivateKey, beaconStaker[0].PrivateKey, 92750*1e9)
		fmt.Println(stakeRes, err)
		stakeRes, err = StakeBeacon(&shard0RPC, beaconStaker[1].PrivateKey, beaconStaker[1].PrivateKey, 91000*1e9)
		fmt.Println(stakeRes, err)
		stakeRes, err = StakeBeacon(&shard0RPC, icoAccount.PrivateKey, beaconStaker[2].PrivateKey, 89250*1e9)
		fmt.Println(stakeRes, err)
		time.Sleep(time.Second * 10)
		stakeRes, err = StakeBeacon(&shard0RPC, icoAccount.PrivateKey, beaconStaker[3].PrivateKey, 87500*1e9)
		fmt.Println(stakeRes, err)
	}

	add_stake := func() {
		stakeRes, err := AddStake(&shard0RPC, beaconStaker[0].PrivateKey, beaconStaker[0].PrivateKey, 1750*10*1e9)
		fmt.Println(stakeRes, err)
		stakeRes, err = AddStake(&shard0RPC, beaconStaker[1].PrivateKey, beaconStaker[1].PrivateKey, 1750*15*1e9)
		fmt.Println(stakeRes, err)
	}

	unstake := func() {
		stakeRes, err := UnStake(&shard0RPC, beaconStaker[0].PrivateKey, beaconStaker[0].PrivateKey)
		fmt.Println(stakeRes, err)
		stakeRes, err = UnStake(&shard0RPC, icoAccount.PrivateKey, beaconStaker[3].PrivateKey)
		fmt.Println(stakeRes, err)
	}

	submitKey()
	time.Sleep(time.Second * 15)
	icoBalance, _ := shard0RPC.GetBalanceByPrivateKey(icoAccount.PrivateKey)
	fmt.Println("Balance", float64(icoBalance)*1e-9)
	send_prv()
	time.Sleep(time.Second * 10)
	stake_shard()
	//time.Sleep(time.Second * 10)
	//stake_beacon()
	//time.Sleep(time.Second * 10)
	//add_stake()
	//time.Sleep(time.Second * 10)
	//unstake()
	go func() {
		for {
			bs, _ := fullnodeRPC.GetBeaconBestState()
			//fmt.Println("BeaconCommittee", bs.BeaconCommittee)
			//fmt.Println("BeaconPendingValidator", bs.BeaconPendingValidator)
			fmt.Println("ShardCommittee", bs.ShardCommittee)
			fmt.Println("ShardPendingValidator", bs.ShardPendingValidator)
			fmt.Println("SyncingValidators", bs.SyncingValidators)
			fmt.Println("BeaconWaiting", bs.BeaconWaiting)
			fmt.Println("BeaconLocking", bs.BeaconLocking)
			time.Sleep(time.Second * 5)
		}
	}()
	select {}

	return
	stake_shard()
	submitKey()
	send_prv()
	stake_beacon()
	add_stake()
	unstake()

}
