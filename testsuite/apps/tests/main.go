package main

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/testsuite/apps/tests/delegation"
)

func RunTest(tests ...func() error) {
	for _, f := range tests {
		err := f()
		if err != nil {
			fmt.Printf("%+v\n", err)
			return
		}
	}
}
func main() {
	RunTest(
		//delegation.TestShardStakingWithDelegation,
		//delegation.TestDelegationAfterStake,
		delegation.TestShardStakingWithReDelegation,
	)
}
