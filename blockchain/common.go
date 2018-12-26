package blockchain

import (
	"sort"

	"github.com/ninjadotorg/constant/metadata"
)

type Txs []metadata.Transaction

func (p Txs) Len() int           { return len(p) }
func (p Txs) Less(i, j int) bool { return p[i].GetLockTime() < p[j].GetLockTime() }
func (p Txs) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p Txs) SortTxs(isDesc bool) Txs {
	if isDesc {
		sort.Sort(sort.Reverse(p))
	}
	sort.Sort(p)
	return p
}
