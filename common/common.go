package common

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"unicode"

	"log"
	"math"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
)

// appDataDir returns an operating system specific directory to be used for
// storing application data for an application.  See AppDataDir for more
// details.  This unexported version takes an operating system argument
// primarily to enable the testing package to properly test the function by
// forcing an operating system that is not the currently one.
func appDataDir(goos, appName string, roaming bool) string {
	if appName == EmptyString || appName == "." {
		return "."
	}

	// The caller really shouldn't prepend the appName with a period, but
	// if they do, handle it gracefully by trimming it.
	appName = strings.TrimPrefix(appName, ".")
	appNameUpper := string(unicode.ToUpper(rune(appName[0]))) + appName[1:]
	appNameLower := string(unicode.ToLower(rune(appName[0]))) + appName[1:]

	// Get the OS specific home directory via the Go standard lib.
	var homeDir string
	usr, err := user.Current()
	if err == nil {
		homeDir = usr.HomeDir
	}

	// Fall back to standard HOME environment variable that works
	// for most POSIX OSes if the directory from the Go standard
	// lib failed.
	if err != nil || homeDir == EmptyString {
		homeDir = os.Getenv("HOME")
	}

	switch goos {
	// Attempt to use the LOCALAPPDATA or APPDATA environment variable on
	// Windows.
	case "windows":
		// Windows XP and before didn't have a LOCALAPPDATA, so fallback
		// to regular APPDATA when LOCALAPPDATA is not set.
		appData := os.Getenv("LOCALAPPDATA")
		if roaming || appData == EmptyString {
			appData = os.Getenv("APPDATA")
		}

		if appData != EmptyString {
			return filepath.Join(appData, appNameUpper)
		}

	case "darwin":
		if homeDir != EmptyString {
			return filepath.Join(homeDir, "Library",
				"Application Support", appNameUpper)
		}

	case "plan9":
		if homeDir != EmptyString {
			return filepath.Join(homeDir, appNameLower)
		}

	default:
		if homeDir != EmptyString {
			return filepath.Join(homeDir, "."+appNameLower)
		}
	}

	// Fall back to the current directory if all else fails.
	return "."
}

// AppDataDir returns an operating system specific directory to be used for
// storing application data for an application.
//
// The appName parameter is the name of the application the data directory is
// being requested for.  This function will prepend a period to the appName for
// POSIX style operating systems since that is standard practice.  An empty
// appName or one with a single dot is treated as requesting the current
// directory so only "." will be returned.  Further, the first character
// of appName will be made lowercase for POSIX style operating systems and
// uppercase for Mac and Windows since that is standard practice.
//
// The roaming parameter only applies to Windows where it specifies the roaming
// application data profile (%APPDATA%) should be used instead of the local one
// (%LOCALAPPDATA%) that is used by default.
//
// Example results:
//  dir := AppDataDir("myapp", false)
//   POSIX (Linux/BSD): ~/.myapp
//   Mac OS: $HOME/Library/Application Support/Myapp
//   Windows: %LOCALAPPDATA%\Myapp
//   Plan 9: $home/myapp
func AppDataDir(appName string, roaming bool) string {
	return appDataDir(runtime.GOOS, appName, roaming)
}

/*
Convert interface of slice to slice
*/
func InterfaceSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		log.Println("InterfaceSlice() given a non-slice type")
		return nil
	}

	ret := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
}

/*
// parseListeners determines whether each listen address is IPv4 and IPv6 and
// returns a slice of appropriate net.Addrs to listen on with TCP. It also
// properly detects addresses which apply to "all interfaces" and adds the
// address as both IPv4 and IPv6.
*/
func ParseListeners(addrs []string, netType string) ([]SimpleAddr, error) {
	netAddrs := make([]SimpleAddr, 0, len(addrs)*2)
	for _, addr := range addrs {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			// Shouldn't happen due to already being normalized.
			return nil, err
		}

		// Empty host or host of * on plan9 is both IPv4 and IPv6.
		if host == EmptyString || (host == "*" && runtime.GOOS == "plan9") {
			netAddrs = append(netAddrs, SimpleAddr{Net: netType + "4", Addr: addr})
			//netAddrs = append(netAddrs, simpleAddr{net: netType + "6", addr: addr})
			continue
		}

		// Strip IPv6 zone id if present since net.ParseIP does not
		// handle it.
		zoneIndex := strings.LastIndex(host, "%")
		if zoneIndex > 0 {
			host = host[:zoneIndex]
		}

		// Parse the IP.
		ip := net.ParseIP(host)
		if ip == nil {
			return nil, fmt.Errorf("'%s' is not a valid IP address", host)
		}

		// To4 returns nil when the IP is not an IPv4 address, so use
		// this determine the address type.
		if ip.To4() == nil {
			//netAddrs = append(netAddrs, simpleAddr{net: netType + "6", addr: addr})
		} else {
			netAddrs = append(netAddrs, SimpleAddr{Net: netType + "4", Addr: addr})
		}
	}
	return netAddrs, nil
}

/*
JsonUnmarshallByteArray - because golang default base64 encode for byte[] data
*/
func JsonUnmarshallByteArray(string string) []byte {
	bytes, _ := base64.StdEncoding.DecodeString(string)
	return bytes
}

/*
SliceExists - Check slice contain item
*/
func SliceExists(slice interface{}, item interface{}) (bool, error) {
	s := reflect.ValueOf(slice)

	if s.Kind() != reflect.Slice {
		return false, errors.New("SliceExists() given a non-slice type")
	}

	for i := 0; i < s.Len(); i++ {
		interfacea := s.Index(i).Interface()
		if interfacea == item {
			return true, nil
		}
	}

	return false, nil
}

/*
SliceBytesExists - Check slice []byte contain item
*/
func GetBytes(key interface{}) ([]byte) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(key)
	return buf.Bytes()
}
func SliceBytesExists(slice interface{}, item interface{}) (int64, error) {
	s := reflect.ValueOf(slice)

	if s.Kind() != reflect.Slice {
		return -1, errors.New("SliceBytesExists() given a non-slice type")
	}

	// TODO upgrade
	min, max := s.Index(0).Interface(), s.Index(s.Len() - 1).Interface()
	var low, high int64
	low, high = 0, int64(s.Len()-1)
	for {
		if bytes.Compare(GetBytes(item), GetBytes(min)) == -1 {
			return int64(low), nil
		}

		if bytes.Compare(GetBytes(item), GetBytes(max)) == 1 {
			return int64(high + 1), nil
		}
		// make a guess of the location
		var guess int64
		if high == low {
			guess = high
		} else {
			size := high - low
			item_Int := new(big.Int).SetBytes(GetBytes(item))
			min_Int := new(big.Int).SetBytes(GetBytes(min))
			max_Int := new(big.Int).SetBytes(GetBytes(max))
			SizeSub1 := new(big.Int).SetInt64(int64(size - 1))
			item_Int.Sub(item_Int, min_Int)
			max_Int.Sub(max_Int, min_Int)
			div := item_Int.Div(item_Int, max_Int)
			offset := SizeSub1.Mul(SizeSub1, div).Int64()
			guess = low + offset
		}
		// maybe we found it?
		x := s.Index(int(guess)).Interface()
		if bytes.Equal(GetBytes(x), GetBytes(item)) {
			//array[guess] == key {
			// scan backwards for start of value range
			temp := s.Index(int(guess) - 1).Interface()
			for guess > 0 && bytes.Equal(GetBytes(temp), GetBytes(item)) {
				guess--
			}
			return int64(guess), nil
		}
		// if we guessed to high, guess lower or vice versa
		if bytes.Compare(GetBytes(x), GetBytes(item)) == 1 {
			high = guess - 1
			max = s.Index(int(high)).Interface()
		} else {
			low = guess + 1
			min = s.Index(int(low)).Interface()
		}
	}
	return -1, nil
}

func GetTxSenderChain(senderLastByte byte) (byte, error) {
	modResult := senderLastByte % 100
	for index := byte(0); index < 5; index++ {
		if (modResult-index)%5 == 0 {
			return byte((modResult - index) / 5), nil
		}
	}
	return 0, errors.New("can't get sender's chainID")
}

func IntArrayEquals(a []int, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func IndexOfStr(item string, list []string) int {
	for k, v := range list {
		if item == v {
			return k
		}
	}
	return -1
}

func ValidateNodeAddress(nodeAddr string) bool {
	if len(nodeAddr) == 0 {
		return false
	}

	strs := strings.Split(nodeAddr, "/ipfs/")
	if len(strs) != 2 {
		return false
	}

	_, err := multiaddr.NewMultiaddr(strs[0])
	if err != nil {
		return false
	}

	_, err = peer.IDB58Decode(strs[1])
	if err != nil {
		return false
	}

	return true
}

// cleanAndExpandPath expands environment variables and leading ~ in the
// passed path, cleans the result, and returns it.
func CleanAndExpandPath(path string, defaultHomeDir string) string {
	// Expand initial ~ to OS specific home directory.
	if strings.HasPrefix(path, "~") {
		homeDir := filepath.Dir(defaultHomeDir)
		path = strings.Replace(path, "~", homeDir, 1)
	}

	// NOTE: The os.ExpandEnv doesn't work with Windows-style %VARIABLE%,
	// but they variables can still be expanded via POSIX-style $VARIABLE.
	return filepath.Clean(os.ExpandEnv(path))
}

func ConstantToMiliConstant(constant uint64) uint64 {
	return constant * uint64(math.Pow(10, NanoConstant))
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func ToBytes(obj interface{}) []byte {
	buff := new(bytes.Buffer)
	json.NewEncoder(buff).Encode(obj)
	return buff.Bytes()
}

// CheckDuplicate returns true if there are at least 2 elements in array have same values
func CheckDuplicateBigInt(arr []*big.Int) bool {
	return false
}

// CheckSND return true if snd exists in snDerivators list
/*
func CheckSNDExistence(snd *big.Int, db database.DatabaseInterface) (bool, error) {
	ok, err := db.HasSNDerivator(*snd, 14)
	if err != nil {
		return false, err
	}
	return ok, nil
}
*/
