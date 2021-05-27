package peerv2

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	cache "github.com/patrickmn/go-cache"

	"github.com/incognitochain/incognito-chain/peerv2/rpcclient"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"stathat.com/c/consistent"
)

type addresses []rpcclient.HighwayAddr // short alias
const MAX_RTT_STORE = 5

type RTTInfo struct {
	lastNcall [MAX_RTT_STORE]time.Duration
	avgRTT    time.Duration
	lastIdx   int
}

// AddrKeeper stores all highway addresses for ConnManager to choose from.
// The address can be used to:
// 1. Make an RPC call to get a new list of highway
// 2. Choose a highway (consistent hashed) and connect to it
// For the 1st type, if it fails, AddrKeeper will ignore the requested address
// for some time so that the next few calls will be more likely to succeed.
// For the 2nd type, caller can manually ignore the chosen address.
type AddrKeeper struct {
	currentHW      rpcclient.HighwayAddr
	addrs          addresses
	addrsByRPCUrl  map[string]*rpcclient.HighwayAddr
	locker         *sync.RWMutex
	ignoreRPCUntil map[rpcclient.HighwayAddr]time.Time // ignored for RPC call until this time
	ignoreHWUntil  map[rpcclient.HighwayAddr]time.Time // ignored for making connection until this time
	ignoreHW       *cache.Cache
	lastRTT        map[rpcclient.HighwayAddr]*RTTInfo
}

func NewAddrKeeper() *AddrKeeper {
	return &AddrKeeper{
		addrs:          addresses{},
		addrsByRPCUrl:  map[string]*rpcclient.HighwayAddr{},
		locker:         &sync.RWMutex{},
		ignoreRPCUntil: map[rpcclient.HighwayAddr]time.Time{},
		ignoreHWUntil:  map[rpcclient.HighwayAddr]time.Time{},
		ignoreHW:       cache.New(MaxTimeIgnoreHW, MaxTimeIgnoreHW),
		lastRTT:        map[rpcclient.HighwayAddr]*RTTInfo{},
	}
}

func (keeper *AddrKeeper) Start(
	host *Host,
	ps *ping.PingService,
	discoverer HighwayDiscoverer,
	// ourPID peer.ID,
) {
	for {
		err := keeper.updateListHighwayAddrs(discoverer)
		if err != nil {
			Logger.Error(err)
			time.Sleep(2 * time.Second)
		} else {
			go keeper.UpdateRTTData(host, ps, make(chan interface{}))
			break
		}
	}
	go func() {
		Logger.Infof("[newpeerv2] updateListHighwayAddrs")
		refreshTimestep := time.NewTicker(UpdateHighwayListTimestep)
		for range refreshTimestep.C {
			Logger.Infof("[newpeerv2] updateListHighwayAddrs")
			err := keeper.updateListHighwayAddrs(discoverer)
			if err != nil {
				Logger.Error(err)
			}
		}
	}()
}

// ChooseHighway refreshes the list of highways by asking a random one and choose a (consistently) random highway to connect
func (keeper *AddrKeeper) ChooseHighway(discoverer HighwayDiscoverer, ourPID peer.ID) (rpcclient.HighwayAddr, error) {
	// Get a list of new highways
	newAddrs, err := keeper.getHighwayAddrs(discoverer)
	if err != nil {
		return rpcclient.HighwayAddr{}, err
	}

	// Update the local list of known highways
	keeper.updateAddrs(newAddrs)
	Logger.Infof("Updated highway addresses: %+v", keeper.addrs)

	// Choose one and return
	chosenAddr, err := keeper.chooseHighwayFromList(ourPID)
	if err != nil {
		return rpcclient.HighwayAddr{}, err
	}
	Logger.Infof("Chosen address: %+v", chosenAddr)
	return chosenAddr, nil
}

func (keeper *AddrKeeper) GetHighway(selfPeerID *peer.ID) (*rpcclient.HighwayAddr, error) {
	cst := consistent.New()
	cst.NumberOfReplicas = 2000
	for rpcUrl, addr := range keeper.addrsByRPCUrl {
		Logger.Infof("%v %v", rpcUrl, addr)
		if _, ok := keeper.ignoreHWUntil[*addr]; !ok {
			cst.Add(rpcUrl)
		}
	}

	closest, err := cst.Get(selfPeerID.Pretty())
	if err != nil {
		return &rpcclient.HighwayAddr{}, errors.Errorf("could not get consistent-hashing peer %v %v", cst.Members(), selfPeerID)
	}
	if hwAddr, ok := keeper.addrsByRPCUrl[closest]; ok {
		return hwAddr, nil
	}
	return nil, errors.Errorf("Can not get new HW")
}

func (keeper *AddrKeeper) updateRTT(
	lastCallRTT time.Duration,
	hwAddr rpcclient.HighwayAddr,
) {
	if info, ok := keeper.lastRTT[hwAddr]; ok {
		firstId := (info.lastIdx + 1) % MAX_RTT_STORE
		if info.lastNcall[firstId] == 0 {
			avgNano := ((info.lastIdx+1)*int(info.avgRTT.Nanoseconds()) + int(lastCallRTT.Nanoseconds())) / (info.lastIdx + 2)
			avg, errParse := time.ParseDuration(fmt.Sprintf("%vns", avgNano))
			if errParse == nil {
				info.avgRTT = avg
				info.lastIdx = firstId
				info.lastNcall[info.lastIdx] = lastCallRTT
			} else {
				info.lastIdx = firstId
				info.lastNcall[info.lastIdx] = info.avgRTT
			}
		} else {
			info.avgRTT = info.avgRTT - info.lastNcall[firstId]/MAX_RTT_STORE + lastCallRTT/MAX_RTT_STORE
		}
		info.lastNcall[firstId] = lastCallRTT
		info.lastIdx = firstId
		keeper.lastRTT[hwAddr] = info
	} else {
		info := &RTTInfo{
			lastNcall: [MAX_RTT_STORE]time.Duration{},
			lastIdx:   0,
			avgRTT:    lastCallRTT,
		}
		for j := 0; j < MAX_RTT_STORE; j++ {
			info.lastNcall[j] = 0
		}
		info.lastNcall[info.lastIdx] = lastCallRTT
		keeper.lastRTT[hwAddr] = info
	}
}

func (keeper *AddrKeeper) UpdateRTTData(
	host *Host,
	ps *ping.PingService,
	stopCh chan interface{},
) {
	reEstimatedTimestep := time.NewTicker(ReEstimatedRTTTimestep)
	defer reEstimatedTimestep.Stop()
	for {
		ignoreList := []rpcclient.HighwayAddr{}
		for _, hwAddr := range keeper.addrs {
			if len(hwAddr.Libp2pAddr) == 0 {
				Logger.Infof("[RTT] hwAddr.Libp2pAddr %v is empty", hwAddr.Libp2pAddr)
				fmt.Printf("[RTT] hwAddr.Libp2pAddr %v is empty\n", hwAddr.Libp2pAddr)
				continue
			}
			addrInfo, err := getAddressInfo(hwAddr.Libp2pAddr)
			if err != nil {
				fmt.Printf("[RTT] cannot get getAddressInfo of hwAddr.Libp2pAddr %v\n", hwAddr.Libp2pAddr)
				continue
			}
			fmt.Println(hwAddr.Libp2pAddr, addrInfo.ID)
			fmt.Println(addrInfo.Addrs)
			ctx, cancel := context.WithTimeout(context.Background(), DialTimeout)
			if err := host.Host.Connect(ctx, *addrInfo); err != nil {
				fmt.Printf("Could not connect to highway: %v %v\n", err, addrInfo)
				ignoreList = append(ignoreList, hwAddr)
				cancel()
			} else {
				ctxPing, cancelPing := context.WithTimeout(context.Background(), PingTimeout)
				ts := ps.Ping(ctxPing, addrInfo.ID)
				s, err := getAvgRTT(ctxPing, ts)
				if err == nil {
					fmt.Printf("[RTT] AVG RTT to HW %v: %v\n", hwAddr.Libp2pAddr, s)
					keeper.updateRTT(s, hwAddr)
				} else {
					ignoreList = append(ignoreList, hwAddr)
					fmt.Printf("Can not get RTT infor to HW %v, error: %v\n", hwAddr, err)
				}
				cancelPing()
				cancel()
			}
		}
		for _, addr := range ignoreList {
			keeper.IgnoreAddress(addr)
		}
		select {
		case <-reEstimatedTimestep.C:
			continue
		case <-stopCh:
			return
		}
	}
	// return []rpcclient.HighwayAddr{}, nil
}

// Add saves a highway address; should only be used at the start for bootnode
// since there's no usage of mutex
func (keeper *AddrKeeper) Add(addr rpcclient.HighwayAddr) {
	keeper.locker.Lock()
	if _, existed := keeper.addrsByRPCUrl[addr.RPCUrl]; !existed {
		keeper.addrsByRPCUrl[addr.RPCUrl] = &addr
		keeper.addrs = append(keeper.addrs, addr)
	}
	keeper.locker.Unlock()
}

func (keeper *AddrKeeper) IgnoreAddress(addr rpcclient.HighwayAddr) {
	keeper.ignoreHW.Add(addr.RPCUrl, addr, MaxTimeIgnoreHW)
	// Logger.Infof("Ignoring address %v in %v", addr, MaxTimeIgnoreHW)
	// keeper.ignoreHWUntil[addr] = time.Now().Add(IgnoreHWDuration)
	// Logger.Infof("Ignoring address %v until %s", addr, keeper.ignoreHWUntil[addr].Format(time.RFC3339))
}

// updateAddrs saves the new list of highway addresses
// Address that aren't in the new list will have their ignore timing reset
// => the next time it appears we will reconnect to it
func (keeper *AddrKeeper) updateAddrs(newAddrs addresses) {
	for _, newAddr := range newAddrs {
		if addr, existed := keeper.addrsByRPCUrl[newAddr.RPCUrl]; existed {
			if (len(addr.Libp2pAddr) == 0) && (len(newAddr.Libp2pAddr) != 0) {
				addr.Libp2pAddr = newAddr.Libp2pAddr
			}
		} else {
			keeper.addrsByRPCUrl[newAddr.RPCUrl] = &newAddr
		}
	}

	// for _, oldAddr := range keeper.addrs {
	// 	found := false
	// 	for _, newAddr := range newAddrs {
	// 		if newAddr == oldAddr {
	// 			found = true
	// 			break
	// 		}
	// 	}

	// 	if len(oldAddr.Libp2pAddr) == 0 {
	// 		newAddrs = append(newAddrs, oldAddr) // Save the bootnode address
	// 	} else if !found {
	// 		delete(keeper.ignoreRPCUntil, oldAddr)
	// 		delete(keeper.ignoreHWUntil, oldAddr)
	// 		Logger.Infof("Resetting ignore time of %v", oldAddr)
	// 	}
	// }

	// // Save the new list
	// keeper.addrs = newAddrs
}

// chooseHighwayFromList returns a random highway address from the known list using consistent hashing; ourPID is the anchor of the hashing
func (keeper *AddrKeeper) chooseHighwayFromList(ourPID peer.ID) (rpcclient.HighwayAddr, error) {
	if len(keeper.addrs) == 0 {
		return rpcclient.HighwayAddr{}, errors.New("cannot choose highway from empty list")
	}

	// Filter out bootnode address (address with only rpcUrl)
	filterAddrs := addresses{}

	Logger.Infof("[testHW] %v", len(keeper.addrs))
	for _, addr := range keeper.addrs {
		Logger.Infof("[testHW] %v", addr.Libp2pAddr)
		if len(addr.Libp2pAddr) != 0 {
			filterAddrs = append(filterAddrs, addr)
		}
	}
	// panic("a")
	// Filter out ignored address
	Logger.Infof("Full known addrs: %v", filterAddrs)
	if addrs := getNonIgnoredAddrs(filterAddrs, keeper.ignoreHWUntil); len(addrs) > 0 {
		filterAddrs = addrs
	} else {
		// Clear timing if all addresses are ignored
		keeper.ignoreHWUntil = map[rpcclient.HighwayAddr]time.Time{}
	}
	Logger.Infof("Choosing highway to connect from non-ignored list %v", filterAddrs)

	// Sort first to make sure always choosing the same highway
	// if the list doesn't change
	// NOTE: this is redundant since hash key doesn't contain indexes
	// But we still keep it anyway to support other consistent hashing library
	sort.SliceStable(filterAddrs, func(i, j int) bool {
		return filterAddrs[i].Libp2pAddr < filterAddrs[j].Libp2pAddr
	})

	addr, err := choosePeer(filterAddrs, ourPID)
	if err != nil {
		return rpcclient.HighwayAddr{}, err
	}
	return addr, nil
}

// choosePeer picks a peer from a list using consistent hashing
func choosePeer(peers addresses, id peer.ID) (rpcclient.HighwayAddr, error) {
	cst := consistent.New()
	cst.NumberOfReplicas = 1000
	for _, p := range peers {
		cst.Add(p.Libp2pAddr)
	}

	closest, err := cst.Get(string(id))
	if err != nil {
		return rpcclient.HighwayAddr{}, errors.Errorf("could not get consistent-hashing peer %v %v", peers, id)
	}

	for _, p := range peers {
		if p.Libp2pAddr == closest {
			return p, nil
		}
	}
	return rpcclient.HighwayAddr{}, errors.Errorf("could not find closest peer %v %v %v", peers, id, closest)
}

// getHighwayAddrs picks a random highway, makes an RPC call to get an updated list of highways
// If fails, the picked address will be ignore for some time.
func (keeper *AddrKeeper) updateListHighwayAddrs(discoverer HighwayDiscoverer) error {
	if len(keeper.addrs) == 0 {
		return errors.New("No peer to get list of highways")
	}

	Logger.Infof("Full RPC address list: %v", keeper.addrs)
	for _, addr := range keeper.addrs {
		Logger.Infof("RPCing addr %v from list", addr)
		newAddrs, err := getAllHighways(discoverer, addr.RPCUrl)
		if err == nil {
			keeper.updateAddrs(newAddrs)
		}
	}
	return nil
}

// getHighwayAddrs picks a random highway, makes an RPC call to get an updated list of highways
// If fails, the picked address will be ignore for some time.
func (keeper *AddrKeeper) getHighwayAddrs(discoverer HighwayDiscoverer) (addresses, error) {
	if len(keeper.addrs) == 0 {
		return nil, errors.New("No peer to get list of highways")
	}

	// Pick random highway to make an RPC call
	Logger.Infof("Full RPC address list: %v", keeper.addrs)
	addrs := getNonIgnoredAddrs(keeper.addrs, keeper.ignoreRPCUntil)
	if len(addrs) == 0 {
		// All ignored, pick random one and clear timing for next call
		addrs = keeper.addrs
		keeper.ignoreRPCUntil = map[rpcclient.HighwayAddr]time.Time{}
	}
	addr := addrs[rand.Intn(len(addrs))]
	Logger.Infof("RPCing addr %v from list %v", addr, addrs)

	newAddrs, err := getAllHighways(discoverer, addr.RPCUrl)
	if err == nil {
		return newAddrs, nil
	}

	// Ignore for a while
	keeper.ignoreRPCUntil[addr] = time.Now().Add(IgnoreRPCDuration)
	Logger.Infof("Ignoring RPC of address %v until %s", addr, keeper.ignoreRPCUntil[addr].Format(time.RFC3339))
	return nil, err
}

func getNonIgnoredAddrs(addrs addresses, ignoreUntil map[rpcclient.HighwayAddr]time.Time) addresses {
	now := time.Now()
	valids := addresses{}
	for _, addr := range addrs {
		if deadline, ok := ignoreUntil[addr]; !ok || now.After(deadline) {
			valids = append(valids, addr)
		}
	}
	return valids
}

func getAllHighways(discoverer HighwayDiscoverer, rpcUrl string) (addresses, error) {
	mapHWPerShard, err := discoverer.DiscoverHighway(rpcUrl, []string{"all"})
	if err != nil {
		return nil, err
	}
	Logger.Infof("Got %v from bootnode", mapHWPerShard)
	return mapHWPerShard["all"], nil
}

func getAddressInfo(libp2pAddr string) (*peer.AddrInfo, error) {
	addr, err := multiaddr.NewMultiaddr(libp2pAddr)
	if err != nil {
		return nil, errors.WithMessagef(err, "invalid libp2p address: %s", libp2pAddr)
	}
	hwPeerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return nil, errors.WithMessagef(err, "invalid multi address: %s", addr)
	}
	return hwPeerInfo, nil
}
