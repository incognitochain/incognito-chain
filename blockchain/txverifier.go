package blockchain

import (
	"fmt"
	"time"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/pkg/errors"
)

type TxsVerifier struct {
	txDB   statedb.StateDB
	txPool TxsCrawler
}

func (v *TxsVerifier) ValidateWithoutChainState(tx metadata.Transaction) (bool, error) {
	if err := tx.LoadCommitment(&v.txDB); err != nil {
		return false, err
	}
	ok, err := tx.ValidateSanityDataByItSelf()
	if !ok || err != nil {
		return ok, err
	}
	return tx.ValidateTxCorrectness()
}

func (v *TxsVerifier) ValidateWithChainState(
	tx metadata.Transaction,
	chainRetriever metadata.ChainRetriever,
	shardViewRetriever metadata.ShardViewRetriever,
	beaconViewRetriever metadata.BeaconViewRetriever,
	beaconHeight uint64,
) (bool, error) {
	//Get state db from beaconview
	if err := tx.LoadCommitment(&v.txDB); err != nil {
		return false, err
	}

	ok, err := tx.ValidateSanityDataWithBlockchain(
		chainRetriever,
		shardViewRetriever,
		beaconViewRetriever,
		beaconHeight,
	)
	if !ok || err != nil {
		return ok, err
	}
	return tx.ValidateDoubleSpendWithBlockChain(&v.txDB)
}

func (v *TxsVerifier) ValidateBlockTransactions(
	chainRetriever metadata.ChainRetriever,
	shardViewRetriever metadata.ShardViewRetriever,
	beaconViewRetriever metadata.BeaconViewRetriever,
	txs []metadata.Transaction,
) bool {
	_, newTxs := v.txPool.CheckValidatedTxs(txs)
	ok := v.validateTxsWithoutChainstate(newTxs)
	if !ok {
		return ok
	}
	ok = v.validateTxsWithChainstate(
		txs,
		chainRetriever,
		shardViewRetriever,
		beaconViewRetriever,
	)
	return ok
}

func (v *TxsVerifier) validateTxsWithoutChainstate(txs []metadata.Transaction) bool {
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()
	// txErrs, ctxNew := errgroup.WithContext(ctx)
	// for _, tx := range txs {
	// 	txErrs.Go(
	// 		func() error {
	// 			ok, err := v.ValidateWithoutChainState(tx)
	// 			if !ok || err != nil {
	// 				return errors.Errorf("This list txs contains a invalid tx %v, validate result %v, error %v", tx.Hash().String(), ok, err)
	// 			}
	// 			return nil
	// 		})
	// }
	// // Wait for the first error from any goroutine.
	// if err := txErrs.Wait(); err != nil {
	// 	fmt.Println(err)
	// }
	errCh := make(chan error)
	// MAX := runtime.NumCPU() - 1
	// nWorkers := make(chan int, MAX)
	for _, tx := range txs {
		select {
		case err := <-errCh:
			fmt.Println(err)
			return false
		default:
			// nWorkers <- 1
			go func() {
				ok, err := v.ValidateWithoutChainState(tx)
				if !ok || err != nil {
					errCh <- errors.Errorf("This list txs contains a invalid tx %v, validate result %v, error %v", tx.Hash().String(), ok, err)
				}
				// <-nWorkers
				if err != nil {
					fmt.Printf("Validate tx %v return error %v:\n", tx.Hash().String(), err)
				}
			}()
		}
	}
	return true

}

func (v *TxsVerifier) validateTxsWithChainstate(
	txs []metadata.Transaction,
	cView metadata.ChainRetriever,
	sView metadata.ShardViewRetriever,
	bcView metadata.BeaconViewRetriever,
) bool {
	errCh := make(chan error)
	doneCh := make(chan interface{}, len(txs))
	numOfValidTxs := 0
	timeout := time.After(5 * time.Second)
	// MAX := runtime.NumCPU() - 1
	// nWorkers := make(chan int, MAX)
	for _, tx := range txs {
		// nWorkers <- 1
		go func() {
			ok, err := v.ValidateWithChainState(
				tx,
				cView,
				sView,
				bcView,
				sView.GetBeaconHeight(),
			)
			if !ok || err != nil {
				errCh <- errors.Errorf("This list txs contains a invalid tx %v, validate result %v, error %v", tx.Hash().String(), ok, err)
				if err != nil {
					fmt.Printf("Validate tx %v return error %v:\n", tx.Hash().String(), err)
				}
			} else {
				doneCh <- nil
			}
			// <-nWorkers

		}()
	}

	for {
		select {
		case err := <-errCh:
			fmt.Println(err)
			return false
		case <-doneCh:
			numOfValidTxs++
			if numOfValidTxs == len(txs) {
				return true
			}
		case <-timeout:
			fmt.Println("Timeout!!!")
			return false
		}
	}
}
