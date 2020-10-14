package committeestate

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/incognitochain/incognito-chain/incognitokey"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/trie"
	"github.com/jrick/logrotate/rotator"
)

var (
	key0                                       = "121VhftSAygpEJZ6i9jGkCVwX5W7tZY6McnoXxZ3xArZQcKduS78P6F6B6T8sjNkoxN7pfjJruViCG3o4X5CiEtHCv9Ufnqp7W3qB9WkuSbGnEKtsNNGpHxJEpdEw4saeueY6kRhqFDcF2NQjgocGLyZsc5Ea6KPBj56kMtUtfcois8pBuFPn2udAsSza7HpkiW7e9kYmzu6Nqnca2jPc8ugCJYHsQDtjmzENC1tje2dfFzCnfkHqam8342bF2wEJgiEwTkkZBY2uLkbQT2X39tSsfzmbqjfrEExjorhFA5yx2ZpKrsA4H9sE34Khy8RradfGCK4L6J4gz1G7YQJ1v2hihEsw3D2fp5ktUh46sicTLmTQ2sfzjnNgMq5uAZ2cJx3HeNiETJ65RVR9J71ujLzdw8xDZvbAPRsdB11Hj2KgKFR"
	key                                        = "121VhftSAygpEJZ6i9jGkEKLMQTKTiiHzeUfeuhpQCcLZtys8FazpWwytpHebkAwgCxvqgUUF13fcSMtp5dgV1YkbRMj3z42TW2EebzAaiGg2DkGPodckN2UsbqhVDibpMgJUHVkLXardemfLdgUqWGtymdxaaRyPM38BAZcLpo2pAjxKv5vG5Uh9zHMkn7ZHtdNHmBmhG8B46UeiGBXYTwhyMe9KGS83jCMPAoUwHhTEXj5qQh6586dHjVxwEkRzp7SKn9iG1FFWdJ97xEkP2ezAapNQ46quVrMggcHFvoZofs1xdd4o5vAmPKnPTZtGTKunFiTWGnpSG9L6r5QpcmapqvRrK5SiuFhNM5DqgzUeHBb7fTfoiWd2N29jkbTGSq8CPUSjx3zdLR9sZguvPdnAA8g25cFPGSZt8aEnFJoPRzM"
	key2                                       = "121VhftSAygpEJZ6i9jGkEqPGAXcmKffwMbzpwxnEfzJxen4oZKPukWAUBbqvV5xPnowZ2eQmAj2mEebG2oexebQPh1MPFC6vEZAk6i7AiRPrZmfaRrRVrBp4WXnVJmL3xK4wzTfkR2rZkhUmSZm112TTyhDNkDQSaBGJkexrPbryqUygazCA2eyo6LnK5qs7jz2RhhsWqUTQ3sQJUuFcYdf2pSnYwhqZqphDCSRizDHeysaua5L7LwS8fY7KZHhPgTuFjvUWWnWSRTmV8u1dTY5kcmMdDZsPiyN9WfqjgVoTFNALjFG8U4GMvzV3kKwVVjuPMsM2XqyPDVpdNQUgLnv2bJS8Tr22A9NgF1FQfWyAny1DYyY3N5H3tfCggsybzZXzrbYPPgokvEynac91y8hPkRdgKW1e7FHzuBnEisPuKzy"
	key3                                       = "121VhftSAygpEJZ6i9jGkGLcYhJBeaJTGY5aFjqQA2WwyxU69Utrviuy9AJ3ATkeEyigVGScQUZw22cD1HeFKiyASYAs82WEamujt3nefYA9FPhURBpRTn6jDmGKUdb4QNbs7HVCJkRRaL9aktg1yaQaZE8TJFg2UeE9tBqUdmvD8fy36aDCYM5W86jaTVCXeEJQWPxUunP2EEL3e283PJ8zqPeBkpoFvkvhB28Hk3oRDeCCTC7QhbaV18ayKeToYqAxoUMBBihanfA33ixeX1daeKpajLCgDZ6jrfphwdYwQbf7dMcZ2NVvQ1a5JUCTJUZypwgKRt8tnTAKCowt2L1KNGP4NJJZm61cfHAGbKRyG9QxCJgK2SdMKsKPVefZSc9LbVaB7VeBby5LHxvMoCD7bN7g1HYRp4BX9n1fZJUeEkVa"
	key4                                       = "121VhftSAygpEJZ6i9jGkDjJj7e2cfgQvrLsPsmLhGMmGD9U9Knffa1MZAw79EijnpueVfTStN2VYt5jRqEr2DTjVqzUinwHVKWH4Tg4szHUntiBdWeqzNC4E8iiwC9Y2KtcRr3hBkpfqvyuBvchigatrigRvFVWu8H2RQqjvopLL51DQ4LFD87L9Zgj9HhasMeyr6f37yirs47JgtGs4BM7EhhpM5zD3TCsFabPphtwDKnfuLMaGzoAw5fM8zEXvdLMuohk96oayjdYothncdtZom17DxB1Mmw535eEjxBwz9ELoZRKk3LYiheSd4xGN9QsxrT2WnZCTd8B5QktARte5S91QYvRMixKC8UEuovQhXt8jMZNkq7CmMeXoybfYdmNaAHuqbY1QeUT2AgaqPho4ay3z5eeKRhnB28H18RGWQ1L"
	key5                                       = "121VhftSAygpEJZ6i9jGkS3D5FaqhzhSF79YFYKHLTHMY5erPhm5vT9VxMtFdWbUVfmhKvfKosXiUKsygyw8knbejNWineCFpx35KegXBbfnVv6AcE3KD4Rs46pDKrqDvWmpaPJoUDdiJeVPQQsFuTykMGs1txt14hhnWMWx9Bf8caDpxR3SKQY7PyHbEhRhdasL3eJC3X1y83PkzJistXPHFdoK4bszU5iE8EiMrXP5GiHTLLTyTpRxScg6AVnrFnmRzPsEMXJAz5zmwUkxwGNrj5iBi7ZJBHo5m3bTVYdQqTSBgVXSqAGZ6fPqAXPGkH6NfgGeZhXR33D3Q4JhEoBs4QWnr89gaVUDAwGXFoXEVfmGwGFFy7jPdLYKuc2u1yJ9YRa1MbSxcPLATui2wmN73UAFa6uSUdN71rCDHJEfCxpS"
	key6                                       = "121VhftSAygpEJZ6i9jGkQXi69eX7p8fmJosf8F4KEdBSqfh3cGxGMd6sGa4hfXTg9vxq16AN97mrqerdNM6ZUGBDgPAipbaGznaHSC8gE7gBpSrVKbNb93nwXSBHSBKFVC6MK5NAFN6bpK25YHrmC248FPr3VQMf9tfG57P5TTH7KWr4bn7v2YbTxNRkZFD9JwkTmwXAwEfWJ12rrc1kMDUkAkrSYYhmpykXTjK9wEBkKFA2z5rnw24cBVL9Tt6M2BEqUM3tuEoUfhiA6E6tdPAkYc7LusTjwikzpwRbVYi4cVMCmC7Dd2UccaA2iiotuyP85AYQSUaHzV2MaF2Cv7GtLqTMm6bRqvpetU1kpkunEnQmAuLVLG7QHPRVKdkX6wRYBE6uRcJ1FaejVbbrF3Tgyh6dsMhRVgEvvvocYPULcJ5"
	key7                                       = "121VhftSAygpEJZ6i9jGk68R6pmXasuHTdRxSeLvBa6dWdc7Mp7c9AxfP6Po9BAi7yRnmiffbEFvs6p6zLFRxwUV1gZLa4ijV7nhPHjxBmJW9vYwV6cJFv2VCN4P1ncnUPf75U8wFxt7AXBQ4o67rsqnHrWvifisURmZzqqaRSUsKAbgqvkmnb3GPcCdjGqFgiYkbwCf4QRWEPnCCdRKabbA2SHDo3bzxJS6CiQNXmKL9SRCrfm1aDcTMUrhPg4w2Gtx8YrQZpHDRYAhgigDgUHPLyLf4Gado342tNTBi9XwvyghJQ6i4PguGrqUvRns8kJ3mbouNWLBc8tQGi3NVN7vb9fmoNm4KSDc22RWESSDkUkj6pAqBiRtJvXjS24DqKTNwQU7FJWobc8a6Qudyxabb5TksrK6d4QirEW8CkX5ahnk"
	key8                                       = "121VhftSAygpEJZ6i9jGkAWwCGm383V8zyMqU2VbEsymfkv3tCPRcRFWtvuTeNVH4r8TDRAdHjaM2j5Nwvw6vqEr58seiM3SMgdDeZwkv942XhG1DmwdrvBPM5RyA3Na32DXRykeHqkAoGP7HbUfUQDZzwkVi3ufHnVEsEVM2CsBTFubBR5YREZVkC4L81a4Hb7BVQZ8yap1kGpZctkTdSCCyGMge2AfqyqvhQ7zn6UCw8aMNnajprw8hJCtuSLEQXA8MwYis6X9cRjKACxYQ9hzyKCvg19PSE7ntf9fXyLxTCmcvCHdNd7cAFrBiDKJHpzp9FVwARyNJF4jEKYmfFi599njpuSSyhQTqEanKg9JnWmp2TNENCEsZ8L9DjbUwbeEWs8uS4Skvx9HeG9itgHL2T3dWKFaisAfBS9YVqVpUnGL"
	key9                                       = "121VhftSAygpEJZ6i9jGk6fLBCjGt1qsb68RVCi2dXNCW2pvwmko9mgsCVsETtbCmjpAtK9PRhfLqVm3TWhpgUf1SuSHgqGYdJnZZBnNaNXhYxT3y5n7Rwx1tS6cXqhp8RqYvbhE2jPuvWvxLzWXpMT1P4kqHeShRGUVLxYZLsY95TZjS3aWuLH1SXMMS1LzZrpSp43PSHDS2qvYMPh4jEHd2r8DqXdUEqxFxfyNDkisFLKZZBNiHGRkt1wjQiDdDsW48zBARS3P32FYZoRhxYB3v4zxGJ3LYeKFuLtxY3uLCqU2nSbpxiGX2f793yEYeGFa394QQyPv5o7km3f7oPMJxdFahqy1xpo45nNgbsiuw287aDn8C3D1YgYnJLACCXreGqQsHZkTtHrNr7ZKh4iGWUTV9ZMj4vCToZXZo2wrhQs1"
	key10                                      = "121VhftSAygpEJZ6i9jGk4unGzNh6zLGgxD83cjWGv7rAtLiRkph2nrPc1CKzCiPyAs8rAJXEfp2wDnhdrU8UvmQfbp1bD95RB1oNvFJrQj3uE6Ei9wfXk3qJ42SfnjRGARVdKppErB5btVcdBb9UzjoR9StKuDVuxtuh9Ntg5Wqjrc6unkoYDAvfvmKkrPgWQM5dy12HtfsNnRkvCHi5UJFKssMqgBpDDLD67KgKuufN63eMRKBZwCN2boZ8N1jGzNujNg2dmZYEn2aQBNC6Kgt7qym6kDvULxLM7QSJ7BJFKcrN63XRYvJFZovNUmnHzxDQn3hA7suUZzFp54XJje86QUicLFThyoAza8PBW7NrJxYYhkkczc6qMSPds7tRgbgfn4LvzFqTim8wNHUVrJAjRecqfKxSXhbCj1qfXcjj2vB"
	key11                                      = "121VhftSAygpEJZ6i9jGkNXmwEzWhQTLpqZbfKfCJVsT6uZwGnwJSXpVkwroPmi1U54av2gwMGmGFDEVhAdt8whvEsmvRrrwQYY8XjoBTG1Kd2NaHNvfPHY3yZLT1ftkr4BbgGseEwJnYaqosTPq1mgLApEbmEcN3YKuuA4eia7s32x2UvKozXXma5EtShwmJ4Q7fGmznpwn91o4ZAT3HnKiCj17rVnBpeWLsr2kWUZPzENo3F9YxzAQ9sNrMpMBWisdAXLvXvakMvCeWwFJ6CRg7GjBzZP4hqsTv5ogt3BRKPoF8be4UDxBVrAreZwsM5pThuSctm55aqTSM4RgTEPBWNoTffaNMGQ4NVQrsuGkWZF8gd2MCRwp3apvU5NxftXsFbghkzmw1ie4JKa4bnjL1r3SxytBxJK5PRDaKhMqyVqc"
	key12                                      = "121VhftSAygpEJZ6i9jGkNVGCa9e2y5pCj4B6kMuryA6kTv8usPJQCWkELX4sNwy5fu1hNB61VLYRLcs4YeNJr6v1AjFRaiupQ4ydPUToQopX7y3kcqEyXWk8fWGxRBkVLyWQb4DZNRWQDk229HUTdfHYwHh6dau1mVS8bVGHg1mTAjEFsTccGowPLWkY22aZocZqA65JhrAPF7TZQt6AkrHaCipxYN2RxGwVsjEBxbk2qbTEw3Yh6i3mhcycxom8VwsyKc62scgwbxXnxdEJr1o4UnZtz8V4wGJXGX6ZRkcqhemZZ9dRhdTKKRGEeEX11Y2yLwA8iKNfYxqrh7qJZmpKbfRkd4d6e7d77qU8BcHnS3r8JPUnN5VsJhDWJuq1Uj8Lfz2St5mKkEVEP1CG1RhT62Q75nd5TRtZqqKRw3ww7Tm"
	incKey0, incKey, incKey2, incKey3, incKey4 *incognitokey.CommitteePublicKey
	incKey5, incKey6, incKey7, incKey8         *incognitokey.CommitteePublicKey
	incKey9, incKey10, incKey11, incKey12      *incognitokey.CommitteePublicKey

	wrarperDB           statedb.DatabaseAccessWarper
	emptyRoot           = common.HexToHash(common.HexEmptyRoot)
	diskDB              incdb.Database
	paymentAddreessKey0 string
)

//initPublicKey init incognito public key for testing by base 58 string
func initPublicKey() {
	paymentAddreessKey0 = "12Rs8bHvYZELqHrv28bYezBQQpteZUEbYjUf2oqV9pJm6Gx4sD4n9mr4UgQe5cDeP9A2x1DsB4mbJ9LT8x2ShaY41cZJWrL7RpFpp2v"
	incKey0 = new(incognitokey.CommitteePublicKey)
	incKey = new(incognitokey.CommitteePublicKey)
	incKey2 = new(incognitokey.CommitteePublicKey)
	incKey3 = new(incognitokey.CommitteePublicKey)
	incKey4 = new(incognitokey.CommitteePublicKey)
	incKey5 = new(incognitokey.CommitteePublicKey)
	incKey6 = new(incognitokey.CommitteePublicKey)
	incKey7 = new(incognitokey.CommitteePublicKey)
	incKey8 = new(incognitokey.CommitteePublicKey)
	incKey9 = new(incognitokey.CommitteePublicKey)
	incKey10 = new(incognitokey.CommitteePublicKey)
	incKey11 = new(incognitokey.CommitteePublicKey)
	incKey12 = new(incognitokey.CommitteePublicKey)

	err := incKey.FromBase58(key)
	if err != nil {
		panic(err)
	}

	err = incKey2.FromBase58(key2)
	if err != nil {
		panic(err)
	}

	err = incKey3.FromBase58(key3)
	if err != nil {
		panic(err)
	}

	err = incKey4.FromBase58(key4)
	if err != nil {
		panic(err)
	}

	err = incKey5.FromBase58(key5)
	if err != nil {
		panic(err)
	}

	err = incKey6.FromBase58(key6)
	if err != nil {
		panic(err)
	}

	err = incKey7.FromBase58(key7)
	if err != nil {
		panic(err)
	}

	err = incKey8.FromBase58(key8)
	if err != nil {
		panic(err)
	}

	err = incKey9.FromBase58(key9)
	if err != nil {
		panic(err)
	}

	err = incKey10.FromBase58(key10)
	if err != nil {
		panic(err)
	}

	err = incKey11.FromBase58(key11)
	if err != nil {
		panic(err)
	}

	err = incKey12.FromBase58(key12)
	if err != nil {
		panic(err)
	}

	err = incKey0.FromBase58(key0)
	if err != nil {
		panic(err)
	}
}

func initStateDB() {
	dbPath, err := ioutil.TempDir(os.TempDir(), "data")
	if err != nil {
		panic(err)
	}
	diskDB, _ = incdb.Open("leveldb", dbPath)
	wrarperDB = statedb.NewDatabaseAccessWarper(diskDB)
	trie.Logger.Init(common.NewBackend(nil).Logger("test", true))
	dataaccessobject.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}

// initLogRotator initializes the logging rotater to write logs to logFile and
// create roll files in the same directory.  It must be called before the
// package-global log rotater variables are used.
func initLogRotator(logFile string) {
	logDir, _ := filepath.Split(logFile)
	err := os.MkdirAll(logDir, 0700)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create log directory: %v\n", err)
		os.Exit(common.ExitByLogging)
	}
	r, err := rotator.New(logFile, 10*1024, false, 3)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create file rotator: %v\n", err)
		os.Exit(common.ExitByLogging)
	}

	logRotator = r
}

// logWriter implements an io.Writer that outputs to both standard output and
// the write-end pipe of an initialized log rotator.
type logWriter struct{}

var logRotator *rotator.Rotator

func (logWriter) Write(p []byte) (n int, err error) {
	os.Stdout.Write(p)
	logRotator.Write(p)
	return len(p), nil
}

func initLog() {
	initLogRotator("./committee-state.log")
	committeeStateLogger := common.NewBackend(logWriter{}).Logger("Committee State log ", false)
	Logger.Init(committeeStateLogger)
}
