package portaltokens

import (
	"errors"
	"fmt"
	"sort"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type utxoItem struct {
	key   string
	value *statedb.UTXO
}

type unshieldItem struct {
	key   string
	value *statedb.WaitingUnshieldRequest
}

func (p PortalBTCTokenProcessor) sortUTXOsByAmountDescending(utxos map[string]*statedb.UTXO) []utxoItem {
	utxosArr := []utxoItem{}
	for k, req := range utxos {
		utxosArr = append(
			utxosArr,
			utxoItem{
				key:   k,
				value: req,
			})
	}
	sort.SliceStable(utxosArr, func(i, j int) bool {
		if utxosArr[i].value.GetOutputAmount() > utxosArr[j].value.GetOutputAmount() {
			return true
		} else if utxosArr[i].value.GetOutputAmount() == utxosArr[j].value.GetOutputAmount() {
			return utxosArr[i].key < utxosArr[j].key
		}
		return false
	})

	return utxosArr
}

func (p PortalBTCTokenProcessor) sortUnshieldReqsByBeaconHeightAscending(
	unshieldReqs map[string]*statedb.WaitingUnshieldRequest,
) []unshieldItem {
	// convert unshield amount to external token amount
	unshieldReqArr := []unshieldItem{}
	for k, req := range unshieldReqs {
		unshieldReqArr = append(
			unshieldReqArr,
			unshieldItem{
				key: k,
				value: statedb.NewWaitingUnshieldRequestStateWithValue(
					req.GetRemoteAddress(), p.ConvertIncToExternalAmount(req.GetAmount()), req.GetUnshieldID(), req.GetBeaconHeight()),
			})
	}

	sort.SliceStable(unshieldReqArr, func(i, j int) bool {
		if unshieldReqArr[i].value.GetBeaconHeight() < unshieldReqArr[j].value.GetBeaconHeight() {
			return true
		} else if unshieldReqArr[i].value.GetBeaconHeight() == unshieldReqArr[j].value.GetBeaconHeight() {
			return unshieldReqArr[i].key < unshieldReqArr[j].key
		}
		return false
	})

	return unshieldReqArr
}

func findUpNearestAmount(arr []utxoItem, amount uint64) (utxoItem, int, error) {
	if len(arr) == 0 {
		return utxoItem{}, -1, errors.New("The array is empty")
	}
	if arr[0].value.GetOutputAmount() < amount {
		return utxoItem{}, -1, errors.New("Not found the up nearest amount")
	}
	if arr[len(arr)-1].value.GetOutputAmount() >= amount {
		return arr[len(arr)-1], len(arr) - 1, nil
	}
	l := 0
	r := len(arr) - 1
	m := 0
	for l <= r {
		m = (l + r) / 2
		if arr[m].value.GetOutputAmount() == amount {
			return arr[m], m, nil
		}
		// find the right
		if arr[m].value.GetOutputAmount() > amount {
			if m+1 <= r {
				if arr[m+1].value.GetOutputAmount() < amount {
					return arr[m], m, nil
				} else if arr[m+1].value.GetOutputAmount() == amount {
					return arr[m+1], m + 1, nil
				}
			}
			l = m + 1
		}

		// find the left
		if arr[m].value.GetOutputAmount() < amount {
			if m-1 >= l && arr[m-1].value.GetOutputAmount() >= amount {
				return arr[m-1], m - 1, nil
			}
			r = m - 1
		}
	}
	return utxoItem{}, -1, errors.New("Not found the up nearest amount")
}

func (p PortalBTCTokenProcessor) ChooseUTXOsForUnshieldReq(utxos []utxoItem, unshieldAmount uint64,
) ([]utxoItem, []int, error) {
	if len(utxos) == 0 {
		return []utxoItem{}, []int{}, errors.New("The list utxos is empty")
	}
	// choose only one utxo for this unshield amount
	if utxos[0].value.GetOutputAmount() > unshieldAmount {
		utxo, index, err := findUpNearestAmount(utxos, unshieldAmount)
		if err != nil {
			return nil, nil, err
		}
		return []utxoItem{utxo}, []int{index}, nil
	}

	// choose multiple utxos for this unshield amount
	chosenUTXOs := []utxoItem{}
	chosenIndices := []int{}
	chosenAmt := uint64(0)
	var err error
	for i, u := range utxos {
		remainAmt := unshieldAmount - chosenAmt
		chosenUTXO := u
		chosenIndex := i
		if u.value.GetOutputAmount() > remainAmt {
			chosenUTXO, chosenIndex, err = findUpNearestAmount(utxos[i:], remainAmt)
			if err != nil {
				return nil, nil, err
			}
			chosenIndex += i
		}

		isValid := p.IsAcceptableTxSize(len(chosenUTXOs)+1, 2)
		if !isValid {
			return []utxoItem{}, []int{}, fmt.Errorf("Number of utxos for unshielding amount %v is exceeded\n",
				unshieldAmount)
		}
		chosenUTXOs = append(chosenUTXOs, chosenUTXO)
		chosenIndices = append(chosenIndices, chosenIndex)
		chosenAmt += chosenUTXO.value.GetOutputAmount()
		if chosenAmt >= unshieldAmount {
			break
		}
	}

	return chosenUTXOs, chosenIndices, nil
}

func (p PortalBTCTokenProcessor) MergeBatches(batchTxs []*BroadcastTx) ([]*BroadcastTx, error) {
	mergedBatches := []*BroadcastTx{}
	if len(batchTxs) == 0 {
		return mergedBatches, nil
	}

	tmpBatchUTXOs := []*statedb.UTXO{}
	tmpBatchUnshieldIDs := []string{}
	for i := 0; i < len(batchTxs); {
		tmpBatchUTXOs = batchTxs[i].UTXOs
		tmpBatchUnshieldIDs = batchTxs[i].UnshieldIDs
		isValid := true
		j := i + 1
		for isValid && j < len(batchTxs) {
			nextBatch := batchTxs[j]
			isValid = p.IsAcceptableTxSize(len(tmpBatchUTXOs)+len(nextBatch.UTXOs),
				len(tmpBatchUnshieldIDs)+len(nextBatch.UnshieldIDs)+1) // 1 for output change
			if isValid {
				tmpBatchUTXOs = append(tmpBatchUTXOs, nextBatch.UTXOs...)
				tmpBatchUnshieldIDs = append(tmpBatchUnshieldIDs, nextBatch.UnshieldIDs...)
				j++
			}
		}
		mergedBatches = append(mergedBatches, &BroadcastTx{
			UTXOs:       tmpBatchUTXOs,
			UnshieldIDs: tmpBatchUnshieldIDs,
		})
		i = j
	}
	return mergedBatches, nil
}

func (p PortalBTCTokenProcessor) EstimateTxSize(numInputs int, numOutputs int) uint {
	return p.ExternalInputSize*uint(numInputs) + p.ExternalOutputSize*uint(numOutputs)
}

func (p PortalBTCTokenProcessor) CalculateTinyUTXONumber(batchTx *BroadcastTx) int {
	estimatedTxSize := p.EstimateTxSize(len(batchTx.UTXOs), len(batchTx.UnshieldIDs)+1)
	remainTxSize := p.ExternalTxMaxSize - estimatedTxSize
	maxTinyUTXOs := int(remainTxSize / p.ExternalInputSize)

	tinyUTXOs := len(batchTx.UnshieldIDs)/3 + 1
	if tinyUTXOs > maxTinyUTXOs {
		tinyUTXOs = maxTinyUTXOs
	}
	return tinyUTXOs
}

func (p PortalBTCTokenProcessor) AppendTinyUTXOs(
	batchTxs []*BroadcastTx, sortedUTXOs []utxoItem, thresholdTinyValue uint64, minUTXOs uint64,
) []*BroadcastTx {
	indexUTXO := len(sortedUTXOs) - 1
	tmpIndexUTXO := indexUTXO
	for i, batch := range batchTxs {
		// only append tiny utxo when number of utxos in vault greater than minUTXOs param
		if uint64(indexUTXO+1) <= minUTXOs {
			return batchTxs
		}
		numTinyUTXOs := p.CalculateTinyUTXONumber(batch)
		fmt.Printf("Batch %v - numTinyUTXOs %v\n", i, numTinyUTXOs)
		for j := indexUTXO; j >= 0 && numTinyUTXOs > 0; j-- {
			if sortedUTXOs[j].value.GetOutputAmount() <= thresholdTinyValue {
				batch.UTXOs = append(batch.UTXOs, sortedUTXOs[j].value)
				numTinyUTXOs--
				tmpIndexUTXO--
			} else {
				return batchTxs
			}
		}
		indexUTXO = tmpIndexUTXO
	}
	return batchTxs
}
