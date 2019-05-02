package common

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
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

	peer "github.com/libp2p/go-libp2p-peer"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
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
		if host == "" || (host == "*" && runtime.GOOS == "plan9") {
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

func ParseListener(addr string, netType string) (*SimpleAddr, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		// Shouldn't happen due to already being normalized.
		return nil, err
	}
	var netAddr *SimpleAddr
	// Empty host or host of * on plan9 is both IPv4 and IPv6.
	if host == EmptyString || (host == "*" && runtime.GOOS == "plan9") {
		netAddr = &SimpleAddr{Net: netType + "4", Addr: addr}
		//netAddrs = append(netAddrs, simpleAddr{net: netType + "6", addr: addr})
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
		return nil, fmt.Errorf("'%s' is not a valid IP address", host)
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
func GetBytes(key interface{}) []byte {
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

	for i := 0; i < s.Len(); i++ {
		interfacea := s.Index(i).Interface()
		if bytes.Equal(interfacea.([]byte), item.([]byte)) {
			return int64(i), nil
		}
	}

	return -1, nil
}

func GetShardIDFromLastByte(b byte) byte {
	return byte(int(b) % MAX_SHARD_NUMBER)
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
		if strings.Compare(item, v) == 0 {
			return k
		}
	}
	return -1
}
func IndexOfStrInHashMap(v string, m map[Hash]string) int {
	for _, value := range m {
		if strings.Compare(value, v) == 0 {
			return 1
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
	return err == nil
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

/*func ConstantToMiliConstant(constant uint64) uint64 {
	return constant * uint64(math.Pow(10, NanoConstant))
}*/

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func Maxint32(x, y int32) int32 {
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

// CheckDuplicate returns true if there are at least 2 elements in an array have same values
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

func RandBigIntN(max *big.Int) (*big.Int, error) {
	return rand.Int(rand.Reader, max)
}

func IntArrayToString(A []int, delim string) string {
	var buffer bytes.Buffer
	for i := 0; i < len(A); i++ {
		buffer.WriteString(strconv.Itoa(A[i]))
		if i != len(A)-1 {
			buffer.WriteString(delim)
		}
	}

	return buffer.String()
}

// ECDSASigToByteArray converts signature to byte array
func ECDSASigToByteArray(r, s *big.Int) (sig []byte) {
	sig = append(sig, r.Bytes()...)
	sig = append(sig, s.Bytes()...)
	return
}

// FromByteArrayToECDSASig converts a byte array to signature
func FromByteArrayToECDSASig(sig []byte) (r, s *big.Int) {
	r = new(big.Int).SetBytes(sig[0:32])
	s = new(big.Int).SetBytes(sig[32:64])
	return
}

func Uint32ToString(I uint32) string {
	return strconv.FormatUint(uint64(I), 10)
}

//return slice of 4 element as little endian
func Uint32ToBytes(value uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, value)
	return b
}

func ByteToBytes(value byte) []byte {
	return []byte{value}
}

func BytesToUint32(b []byte) uint32 {
	return binary.LittleEndian.Uint32(b)
}

func Uint8ToBytes(value uint8) []byte {
	b := make([]byte, 1)
	b[0] = byte(value)
	return b
}

func BytesToUint8(b []byte) uint8 {
	return uint8(b[0])
}

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

func BytesToUint64(b []byte) uint64 {
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

func SliceInterfaceToSliceByte(Arr []interface{}) []byte {
	res := make([]byte, 0)
	for _, element := range Arr {
		res = append(res, element.(byte))
	}
	return res
}

func SliceInterfaceToSliceSliceByte(Arr []interface{}) [][]byte {
	res := make([][]byte, 0)
	for _, element := range Arr {
		res = append(res, element.([]byte))
	}
	return res
}

func SliceInterfaceToSliceString(Arr []interface{}) []string {
	res := make([]string, 0)
	for _, element := range Arr {
		res = append(res, element.(string))
	}
	return res
}

func BytesPlusOne(b []byte) []byte {
	res := make([]byte, len(b))
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] < 0xff {
			copy(res[0:i], b[0:i])
			res[i] = b[i] + 1
			break
		} else {
			res[i] = 0
		}
	}
	return res
}

func IsOffChainAsset(assetID *Hash) bool {
	return bytes.Equal(assetID[:8], BTCAssetID[:8])
}

func IsBondAsset(assetID *Hash) bool {
	return bytes.Equal(assetID[:8], BondTokenID[:8])
}

func IsConstantAsset(assetID *Hash) bool {
	return assetID.IsEqual(&ConstantID)
}

func IsDCBTokenAsset(assetID *Hash) bool {
	return assetID.IsEqual(&DCBTokenID)
}

func IsUSDAsset(assetID *Hash) bool {
	return assetID.IsEqual(&USDAssetID)
}

func IndexOfByte(item byte, arrays []byte) int {
	for k, v := range arrays {
		if v == item {
			return k
		}
	}
	return -1
}

// MilliEtherValue converts amount of milliether to cent using current price of ether
func MilliEtherValue(a uint64, p uint64) uint64 {
	milliEtherToEtherRatio := big.NewInt(1000)
	v := big.NewInt(int64(a))
	v.Mul(v, big.NewInt(int64(p)))
	v.Quo(v, milliEtherToEtherRatio)
	return v.Uint64()
}

// CentInMilliEther converts amount of cent to milliether using current price of ether
func CentInMilliEther(a uint64, p uint64) uint64 {
	milliEtherToEtherRatio := big.NewInt(1000)
	v := big.NewInt(int64(a))
	v.Mul(v, milliEtherToEtherRatio)
	v.Quo(v, big.NewInt(int64(p)))
	return v.Uint64()
}

func FileLog(fromShard bool, info string) {
	shardFile := "~/shard.txt"
	beaconFile := "~/beacon.txt"
	var foName string
	if fromShard {
		foName = shardFile
	} else {
		foName = beaconFile
	}
	fo, err := os.Open(foName)
	defer fo.Sync()
	if err != nil {
		fo, _ = os.Create(foName)
	}
	fo.WriteString(info + "\n")
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
