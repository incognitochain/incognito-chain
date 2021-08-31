package rpcserver

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"strconv"
	"time"
)

// genCertPair generates a key/cert pair to the paths provided.
func GenCertPair(certFile, keyFile string) error {
	Logger.log.Info("Generating TLS certificates...")

	org := "autogenerated cert"
	validUntil := time.Now().Add(10 * 365 * 24 * time.Hour)
	cert, key, err := NewTLSCertPair(org, validUntil, nil)
	if err != nil {
		return err
	}

	// Write cert and key files.
	if err = ioutil.WriteFile(certFile, cert, 0666); err != nil {
		return err
	}
	if err = ioutil.WriteFile(keyFile, key, 0600); err != nil {
		os.Remove(certFile)
		return err
	}

	Logger.log.Infof("Done generating TLS certificates")
	return nil
}

// NewTLSCertPair returns a new PEM-encoded x.509 certificate pair
// based on a 521-bit ECDSA private key.  The machine's local interface
// addresses and all variants of IPv4 and IPv6 localhost are included as
// valid IP addresses.
func NewTLSCertPair(organization string, validUntil time.Time, extraHosts []string) (cert, key []byte, err error) {
	now := time.Now()
	if validUntil.Before(now) {
		return nil, nil, errors.New("validUntil would create an already-expired certificate")
	}

	priv, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	// end of ASN.1 time
	endOfTime := time.Date(2049, 12, 31, 23, 59, 59, 0, time.UTC)
	if validUntil.After(endOfTime) {
		validUntil = endOfTime
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %s", err)
	}

	host, err := os.Hostname()
	if err != nil {
		return nil, nil, err
	}

	ipAddresses := []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}
	dnsNames := []string{host}
	if host != "localhost" {
		dnsNames = append(dnsNames, "localhost")
	}

	addIP := func(ipAddr net.IP) {
		for _, ip := range ipAddresses {
			if ip.Equal(ipAddr) {
				return
			}
		}
		ipAddresses = append(ipAddresses, ipAddr)
	}
	addHost := func(host string) {
		for _, dnsName := range dnsNames {
			if host == dnsName {
				return
			}
		}
		dnsNames = append(dnsNames, host)
	}

	addrs, err := interfaceAddrs()
	if err != nil {
		return nil, nil, err
	}
	for _, a := range addrs {
		ipAddr, _, err := net.ParseCIDR(a.String())
		if err == nil {
			addIP(ipAddr)
		}
	}

	for _, hostStr := range extraHosts {
		host, _, err := net.SplitHostPort(hostStr)
		if err != nil {
			host = hostStr
		}
		if ip := net.ParseIP(host); ip != nil {
			addIP(ip)
		} else {
			addHost(host)
		}
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{organization},
			CommonName:   host,
		},
		NotBefore: now.Add(-time.Hour * 24),
		NotAfter:  validUntil,

		KeyUsage: x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature |
			x509.KeyUsageCertSign,
		IsCA:                  true, // so can sign self.
		BasicConstraintsValid: true,

		DNSNames:    dnsNames,
		IPAddresses: ipAddresses,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template,
		&template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %+v", err)
	}

	certBuf := &bytes.Buffer{}
	err = pem.Encode(certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode certificate: %+v", err)
	}

	keybytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal private key: %+v", err)
	}

	keyBuf := &bytes.Buffer{}
	err = pem.Encode(keyBuf, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keybytes})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode private key: %+v", err)
	}

	return certBuf.Bytes(), keyBuf.Bytes(), nil
}

// interfaceAddrs returns a list of the system's network interface addresses.
// It is wrapped here so that we can substitute it for other functions when
// building for systems that do not allow access to net.InterfaceAddrs().
func interfaceAddrs() ([]net.Addr, error) {
	return net.InterfaceAddrs()
}

type Pdexv3AddLiquidityRequest struct {
	TokenID     string `json:"TokenID"`
	TokenAmount string `json:"TokenAmount"`
	PoolPairID  string `json:"PoolPairID"`
	Amplifier   string `json:"Amplifier"`
	PairHash    string `json:"PairHash"`
	NftID       string `json:"NftID"`
}

type Pdexv3WithdrawLiquidityRequest struct {
	TokenID      string `json:"TokenID"`
	TokenAmount  string `json:"TokenAmount"`
	PoolPairID   string `json:"PoolPairID"`
	Index        string `json:"Index"`
	Token0Amount string `json:"Token0Amount"`
	Token1Amount string `json:"Token1Amount"`
}

type Pdexv3StakingRequest struct {
	TokenID     string `json:"TokenID"`
	TokenAmount string `json:"TokenAmount"`
	NftID       string `json:"NftID"`
}

// Uint64Reader wraps the unmarshaling of uint64 numbers from both integer & string formats.
type Uint64Reader uint64

func (u Uint64Reader) MarshalJSON() ([]byte, error) {
	return json.Marshal(u)
}
func (u *Uint64Reader) UnmarshalJSON(raw []byte) error {
	var theNum uint64
	err := json.Unmarshal(raw, &theNum)
	if err != nil {
		var theStr string
		json.Unmarshal(raw, &theStr)
		temp, err := strconv.ParseUint(theStr, 10, 64)
		*u = Uint64Reader(temp)
		return err
	}
	*u = Uint64Reader(theNum)
	return err
}
