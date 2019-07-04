package common

import (
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"fmt"
	"errors"
)

/*
	Unit test for InterfaceSlice function
 */

func TestCommonInterfaceSlice(t *testing.T){
	data := []struct{
		slice interface{}
		len int
	}{
		{[]byte{1}, 1},
		{[]byte{1,2,3}, 3},
		{[]byte{1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5}, 25},
		{[]byte{1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5}, 30},
	}

	for _, item := range data{
		slice := InterfaceSlice(item.slice)
		assert.Equal(t, item.len, len(slice))
	}
}

func TestCommonInterfaceSliceWithInvalidSliceInterface(t *testing.T){
	data := []struct{
		slice interface{}
	}{
		{"abc"},
		{123},
		{struct{a int}{12}},
		{nil},
	}

	for _, item := range data{
		slice := InterfaceSlice(item.slice)
		assert.Equal(t, []interface{}(nil), slice)
	}
}


/*
	Unit test for ParseListener function
 */

func TestCommonParseListener(t *testing.T){
	data := []struct{
		addr string
		netType string
	}{
		{"1.2.3.4:9934", "test"},
		{"100.255.3.4:9934", "main"},
		{"192.168.3.4:9934", "main1"},
		{"0.0.0.0:9934", "main"},
		{":9934", "main"},		// empty host
		{"1.2.3.4:9934", ""}, // empty netType
	}

	for _, item := range data{
		SimpleAddr, err := ParseListener(item.addr, item.netType)
		fmt.Printf("SimpleAddr.Addr: %v\n", SimpleAddr.Addr)
		fmt.Printf("SimpleAddr.Net: %v\n", SimpleAddr.Net)

		assert.Equal(t, nil, err)
		assert.Equal(t, item.addr, SimpleAddr.Addr)
		assert.Equal(t, item.netType + "4", SimpleAddr.Net)
	}
}

func TestCommonParseListenerWithInvalidIPAddr(t *testing.T){
	data := []struct{
		addr string
		netType string
	}{
		{"256.2.3.4:9934", "test"},
		{"1.2.3:9934", "main1"},
		{"1.2:9934", "main1"},
		{"1:9934", "main1"},
		{"*:9934", "main1"},
		{"a.2.3.4:9934", "test"},
		{"-.2.3.4:9934", "test"},
	}

	for _, item := range data{
		_, err := ParseListener(item.addr, item.netType)
		assert.Equal(t, errors.New("IP address is invalid"), err)
	}
}

func TestCommonParseListenerWithInvalidPort(t *testing.T){
	data := []struct{
		addr string
		netType string
	}{
		{"100.255.3.4:-2", "main"},
		{"192.168.3.4:a", "main1"},
		{"0.0.0.0:?", "main"},
		{":", "main"},		// empty port
		{"1.2.3.4:...", ""}, // empty netType
	}

	for _, item := range data{
		_, err := ParseListener(item.addr, item.netType)
		assert.Equal(t, errors.New("port is invalid"), err)
	}
}

/*
	Unit test for ParseListeners function
 */

func TestCommonParseListeners(t *testing.T){
	addrs := []string{
		"1.2.3.4:9934",
		"100.255.3.4:9934",
		"100.255.3.4:9934",
		"0.0.0.0:9934",
		":9934",
		"1.2.3.4:9934",
	}

	netType := "test"

	simpleAddrs, err := ParseListeners(addrs, netType)

	assert.Equal(t, nil, err)
	assert.Equal(t, 6, len(simpleAddrs))
}

func TestCommonParseListenersWithInvalidIPAddr(t *testing.T){
	addrs := []string{
		"256.2.3.4:9934",
		"100.255.3.4:9934",
		"100.255.3.4:9934",
		"0.0.0.0:9934",
		":9934",
		"1.2.3.4:9934",
	}

	netType := "test"

	simpleAddrs, err := ParseListeners(addrs, netType)

	assert.Equal(t, errors.New("IP address is invalid"), err)
	assert.Equal(t, 0, len(simpleAddrs))
}

func TestCommonParseListenersWithInvalidPort(t *testing.T){
	addrs := []string{
		"100.2.3.4:a",
		"100.255.3.4:9934",
		"100.255.3.4:9934",
		"0.0.0.0:9934",
		":9934",
		"1.2.3.4:9934",
	}

	netType := "test"

	simpleAddrs, err := ParseListeners(addrs, netType)

	assert.Equal(t, errors.New("port is invalid"), err)
	assert.Equal(t, 0, len(simpleAddrs))
}

/*
	Unit test for SliceExists function
 */

func TestCommonSliceExists(t *testing.T){
	data := []struct{
		slice interface{}
		item interface{}
		isContain bool
	}{
		{[]byte{1,2,3,4,5,6}, byte(6), true},
		{[]int{1,2,3,4,5,6}, int(10), false},
		{[]byte{1,2,3,4,5,6}, 6, false},
		{[]string{"a", "b", "c", "d", "e"}, "E", false},
		{[]*big.Int{big.NewInt(int64(100)), big.NewInt(int64(1000)), big.NewInt(int64(10000)), big.NewInt(int64(100000)), big.NewInt(int64(10000000))}, big.NewInt(int64(100001)), false},
	}

	for _, dataItem := range data {
		isContain, err := SliceExists(dataItem.slice, dataItem.item)
		assert.Equal(t, nil, err)
		assert.Equal(t, dataItem.isContain, isContain)
	}
}