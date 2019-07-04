package common

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// appDataDir returns an operating system specific directory to be used for
// storing application data for an application.  See AppDataDir for more
// details.  This unexported version takes an operating system argument
// primarily to enable the testing package to properly test the function by
// forcing an operating system that is not the currently one.
func appDataDir(goos, appName string, roaming bool) string {
	if appName == "" || appName == "." {
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
	if err != nil || homeDir == "" {
		homeDir = os.Getenv("HOME")
	}

	switch goos {
	// Attempt to use the LOCALAPPDATA or APPDATA environment variable on
	// Windows.
	case "windows":
		// Windows XP and before didn't have a LOCALAPPDATA, so fallback
		// to regular APPDATA when LOCALAPPDATA is not set.
		appData := os.Getenv("LOCALAPPDATA")
		if roaming || appData == "" {
			appData = os.Getenv("APPDATA")
		}

		if appData != "" {
			return filepath.Join(appData, appNameUpper)
		}

	case "darwin":
		if homeDir != "" {
			return filepath.Join(homeDir, "Library",
				"Application Support", appNameUpper)
		}

	case "plan9":
		if homeDir != "" {
			return filepath.Join(homeDir, appNameLower)
		}

	default:
		if homeDir != "" {
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

// InterfaceSlice receives a slice which is a interface
// and converts it into slice of interface
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

// ParseListeners determines whether each listen address is IPv4 and IPv6 and
// returns a slice of appropriate net.Addrs to listen on with TCP. It also
// properly detects addresses which apply to "all interfaces" and adds the
// address as both IPv4 and IPv6.
func ParseListeners(addrs []string, netType string) ([]SimpleAddr, error) {
	simpleAddrs := make([]SimpleAddr, len(addrs))

	for i, addr := range addrs {
		simpleAddr, err := ParseListener(addr, netType)
		if err != nil {
			return []SimpleAddr{}, err
		}

		simpleAddrs[i] = *simpleAddr
	}

	return simpleAddrs, nil
}

// ParseListener determines whether the listen address is IPv4 and IPv6 and
// returns a slice of appropriate net.Addrs to listen on with TCP. It also
// properly detects address which apply to "all interfaces" and adds the
// address as both IPv4 and IPv6.
func ParseListener(addr string, netType string) (*SimpleAddr, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		// Shouldn't happen due to already being normalized.
		return nil, err
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 0 {
		return nil, errors.New("port is invalid")
	}

	var netAddr *SimpleAddr
	// Empty host or host of * on plan9 is both IPv4 and IPv6.
	if host == EmptyString || (host == "*" && runtime.GOOS == "plan9") {
		netAddr = &SimpleAddr{Net: netType + "4", Addr: addr}
		return netAddr, nil
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
		fmt.Printf("'%s' is not a valid IP address\n", host)
		return nil, errors.New("IP address is invalid")
	}

	// To4 returns nil when the IP is not an IPv4 address, so use
	// this determine the address type.
	if ip.To4() == nil {
		//netAddrs = append(netAddrs, simpleAddr{net: netType + "6", addr: addr})
	} else {
		netAddr = &SimpleAddr{Net: netType + "4", Addr: addr}
	}
	return netAddr, nil
}

// SliceExists receives a slice and a item in interface type
// checks whether the slice contain the item or not
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

// GetShardIDFromLastByte receives a last byte of public key and
// returns a corresponding shardID
func GetShardIDFromLastByte(b byte) byte {
	return byte(int(b) % MAX_SHARD_NUMBER)
}

// IndexOfStr receives a list of strings and a item string
// It checks whether a item is contained in list or not
// and returns the first index of the item in the list
// It returns -1 if the item is not in the list
func IndexOfStr(item string, list []string) int {
	for k, v := range list {
		if strings.Compare(item, v) == 0 {
			return k
		}
	}
	return -1
}

// IndexOfStrInHashMap receives a map[Hash]string and a value string
// It checks whether a value is contained in map or not
// It returns -1 if the item is not in the list and return 1 otherwise
func IndexOfStrInHashMap(v string, m map[Hash]string) int {
	for _, value := range m {
		if strings.Compare(value, v) == 0 {
			return 1
		}
	}
	return -1
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

// CheckDuplicate returns true if there are at least 2 elements have the same value in the array
func CheckDuplicateBigIntArray(arr []*big.Int) bool {
	sort.Slice(arr, func(i, j int) bool {
		return arr[i].Cmp(arr[j]) == -1
	})

	for i := 0; i < len(arr)-1; i++ {
		if arr[i].Cmp(arr[i+1]) == 0 {
			return true
		}
	}

	return false
}

// RandBigIntMaxRange generates a big int with maximum value
func RandBigIntMaxRange(max *big.Int) (*big.Int, error) {
	return rand.Int(rand.Reader, max)
}

// CompareStringArray receives 2 arrays of string
// and check whether 2 arrays is the same or not
func CompareStringArray(src []string, dst []string) bool {
	if len(src) != len(dst) {
		return false
	}
	for idx, val := range src {
		if dst[idx] != val {
			return false
		}
	}
	return true
}

func BytesToInt32(b []byte) int32 {
	return int32(binary.LittleEndian.Uint32(b))
}

func Int32ToBytes(value int32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(value))
	return b
}

func BytesToUint32(b []byte) uint32 {
	return binary.LittleEndian.Uint32(b)
}

func Uint32ToBytes(value uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, value)
	return b
}

func BytesToUint64(b []byte) uint64 {
	fmt.Printf("BytesToUint64 b: %v\n", b)
	return binary.LittleEndian.Uint64(b)
}

func Uint64ToBytes(value uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, value)
	return b
}

func Int64ToBytes(value int64) []byte {
	return Uint64ToBytes(uint64(value))
}

func BoolToByte(value bool) byte {
	var bitSetVar byte
	if value {
		bitSetVar = 1
	}
	return bitSetVar
}

func IndexOfByte(item byte, arrays []byte) int {
	for k, v := range arrays {
		if v == item {
			return k
		}
	}
	return -1
}

type ErrorSaver struct {
	err error
}

func (s *ErrorSaver) Save(errs ...error) error {
	if s.err != nil {
		return s.err
	}
	for _, err := range errs {
		if err != nil {
			s.err = err
			return s.err
		}
	}
	return nil
}

func (s *ErrorSaver) Get() error {
	return s.err
}

func CheckError(errs ...error) error {
	errSaver := &ErrorSaver{}
	return errSaver.Save(errs...)
}

func AppendSliceString(arrayStrings ...[][]string) [][]string {
	res := [][]string{}
	for _, arrayString := range arrayStrings {
		res = append(res, arrayString...)
	}
	return res
}
