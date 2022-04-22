package pdex

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strings"

	v2 "github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
)

type Order = rawdbv2.Pdexv3Order

type Orderbook struct {
	orders []*Order
}

func (ob Orderbook) MarshalJSON() ([]byte, error) {
	temp := struct {
		Orders []*Order `json:"orders,omitempty"`
	}{ob.orders}
	return json.Marshal(temp)
}

func (ob *Orderbook) UnmarshalJSON(data []byte) error {
	var temp struct {
		Orders []*Order `json:"orders"`
	}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	ob.orders = temp.Orders
	return nil
}

// InsertOrder() appends a new order while keeping the list sorted (ascending by Token1Rate / Token0Rate)
func (ob *Orderbook) InsertOrder(ord *Order) {
	insertAt := func(lst []*Order, i int, newItem *Order) []*Order {
		if i == len(lst) {
			return append(lst, newItem)
		}
		lst = append(lst[:i+1], lst[i:]...)
		lst[i] = newItem
		return lst
	}
	index := sort.Search(len(ob.orders), func(i int) bool {
		ordRate := big.NewInt(0).SetUint64(ob.orders[i].Token0Rate())
		ordRate.Mul(ordRate, big.NewInt(0).SetUint64(ord.Token1Rate()))
		myRate := big.NewInt(0).SetUint64(ob.orders[i].Token1Rate())
		myRate.Mul(myRate, big.NewInt(0).SetUint64(ord.Token0Rate()))

		// compare Token1Rate / Token0Rate of current order in the list to ord
		rateCmp := ordRate.Cmp(myRate)
		// break equality of rate by comparing ID
		if rateCmp == 0 {
			// sell0 orders must precede sell1 orders of the same rate
			if ord.TradeDirection() != ob.orders[i].TradeDirection() {
				return ord.TradeDirection() == v2.TradeDirectionSell0
			}
			// no equality for ID since duplicate ID was handled in addOrder flow
			idCmp := strings.Compare(ord.Id(), ob.orders[i].Id())
			// best rate for sell0 is at start of list, so we put smaller ID first to match. The opposite is true for sell1
			if ord.TradeDirection() == v2.TradeDirectionSell0 {
				return idCmp < 0
			} else {
				return idCmp > 0
			}
		}

		return ordRate.Cmp(myRate) < 0
	})
	ob.orders = insertAt(ob.orders, index, ord)
}

// NextOrder() returns the matchable order with the best rate that has any outstanding balance to sell
func (ob *Orderbook) NextOrder(tradeDirection byte) (*v2.MatchingOrder, string, error) {
	lstLen := len(ob.orders)
	switch tradeDirection {
	case v2.TradeDirectionSell0:
		for i := lstLen - 1; i >= 0; i-- {
			currentOrder := &v2.MatchingOrder{ob.orders[i]}
			if check, err := currentOrder.CanMatch(tradeDirection); check && err == nil {
				return currentOrder, ob.orders[i].Id(), nil
			}
		}
		// no active order
		return nil, "", nil
	case v2.TradeDirectionSell1:
		for i := 0; i < lstLen; i++ {
			currentOrder := &v2.MatchingOrder{ob.orders[i]}
			if check, err := currentOrder.CanMatch(tradeDirection); check && err == nil {
				return currentOrder, ob.orders[i].Id(), nil
			}
		}
		// no active order
		return nil, "", nil
	default:
		return nil, "", fmt.Errorf("Invalid trade direction %d", tradeDirection)
	}
}

// RemoveOrder() removes one order by its index
func (ob *Orderbook) RemoveOrder(index int) error {
	if index < 0 || index >= len(ob.orders) {
		return fmt.Errorf("Invalid order index %d for orderbook length %d", index, len(ob.orders))
	}
	ob.orders = append(ob.orders[:index], ob.orders[index+1:]...)
	return nil
}

func (ob *Orderbook) getDiff(otherBook *Orderbook,
	poolPairChange *v2.PoolPairChange) *v2.PoolPairChange {
	newPoolPairChange := poolPairChange
	theirOrdersByID := make(map[string]*Order)
	for _, ord := range otherBook.orders {
		theirOrdersByID[ord.Id()] = ord
	}
	myOrdersByID := make(map[string]*Order)
	for _, ord := range ob.orders {
		myOrdersByID[ord.Id()] = ord
	}

	// mark new & updated orders as changed
	for _, ord := range ob.orders {
		if existingOrder, exists := theirOrdersByID[ord.Id()]; !exists ||
			!reflect.DeepEqual(*ord, *existingOrder) {
			newPoolPairChange.OrderIDs[ord.Id()] = true
		}
	}

	// mark deleted orders as changed
	for _, ord := range otherBook.orders {
		if _, exists := myOrdersByID[ord.Id()]; !exists {
			newPoolPairChange.OrderIDs[ord.Id()] = true
		}
	}
	return newPoolPairChange
}

func (ob *Orderbook) Clone() Orderbook {
	result := &Orderbook{make([]*Order, len(ob.orders))}
	for index, item := range ob.orders {
		var temp Order = *item
		result.orders[index] = &temp
	}
	return *result
}

func (ob *Orderbook) NftIDs() map[string]string {
	result := make(map[string]string)
	for _, ord := range ob.orders {
		result[ord.Id()] = ord.NftID().String()
	}
	return result
}

func (ob *Orderbook) Orders() []*Order {
	return ob.orders
}
