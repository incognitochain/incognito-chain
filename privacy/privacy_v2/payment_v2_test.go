package privacy_v2

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/stretchr/testify/assert"
)

func TestByteConversionTxWithVer1(t *testing.T) {
	inpData := "{\"CoinDetails\":\"1BcyBDZNNTtF6QAoePQcamAz4sd9tUobDsRxcYeWBu5uEawJrTkhDM2UsrhHxiECNLKUBtzxpAfAWJhFNEYkCszi1TKpBm4TbrKXmWoVggnHsZCfuYDEXvhGoXYWnGaDtMCYJMWYcdRNbE3L7m145WxbmjnUcaKeYXRFfJeb8UcwiSYN2nF4awNA18SYjphscUx5PHvYaB8Yv4HoN2nbC46nN6UfHtdogn8pwf6jYQG2gunN7Kr8\"}"

	inputCoins := make([]*coin.InputCoin, 1)
	inputCoins[0] = new(coin.InputCoin)
	err := json.Unmarshal([]byte(inpData), inputCoins[0])
	if err != nil {
		fmt.Println("Cannot parse inputCoins")
		assert.Equal(t, false, true)
	}

	outData := []string{
		"{\"CoinDetails\":\"12p461nwsF5ijmKy8uyEWhokDNsjQWeW9pdEqGpXi6wwGfgHtfsSq3aiRi2Fc8zJVNosNdoTenMugCK2jHMudw8HBwvdA9W816rt71sjzR26Qwmo3imgeJzXi6KAhe5bHQaF1DYFBQC4GRQQLB3HuTbm7nuubBHNK2RPumJme2bbdLtzSZdAXbN5okAx83Q4bMJZHk\",\"CoinDetailsEncrypted\":null}",
		"{\"CoinDetails\":\"12qBZ2eS4ocD7eP5TWhVgPownegiRCFWD6YC4CiaFhsUv2E7n2oMWcdqifMZxk7jjHw31K9hXBTacrTx6upDKXJtvY2AycCgPaZsnNcpDqAg7kSN5xYJpC42oGMytgdmtr36hNDwmgGeb25aydrttWX4NCDkJkaTHXVmrq3xXC9gJs8ySKH1WrYEzseA4FjrWG2Fpq\",\"CoinDetailsEncrypted\":null}",
		"{\"CoinDetails\":\"12qAxm15xwdW9zCsaoa8hGg1ZrSTeVD79VDH7KZqLYszMNLvUeqaofUHWiAbZa1QdSz3QqKLpXPkwrcMp626EAuC7jRTU1o3L7qMUnDqNWSWTKbJ1ypi9NbH9h2GP8iTVuisqwRFRQ5gKQesySujM98rvEiqCQBbsz63FMYB2nA17Sg4vo8qgQfeaDD8HCmkFscT9j\",\"CoinDetailsEncrypted\":null}",
		"{\"CoinDetails\":\"12o4NbSqg3t4WaupwBYzsTtCoVEdGNvrTNfuUdHhsr5P8hqCDus16vYaWipB9Z6vM73ZnjBYPJ2sk9DMf5xXCEzFTWCuHWkh8iXjVUuExJfoxbZcYiPxHXzHeTjZLWjcTbrRQngcYXPsRnN8Mr5Ecm6uGrMoVjPJi4Uy2WprvoC9DFoaTmFNx6sDVY5Fp3CSn1ZR1o\",\"CoinDetailsEncrypted\":null}",
		"{\"CoinDetails\":\"12oZn7Ae6h7LXERKR8eUjfwu247rLED4aszs4tRhmfkyLmmE8msKJjswfHMzWkPq9gGHxqSwVrRyKukk8Bbftk98XCt5dRY3irEQK8dGn9B1w5JGwCoUPXfxmPPRmiEe1v81zyUVdjC3U533bBLzVJQ3dhoMR2rz6RoSdeRjMHPYujBRAuRv1CsqJYLk72or5yrqD7\",\"CoinDetailsEncrypted\":null}",
		"{\"CoinDetails\":\"12qiD69y8Z9ktUHmJ2Cm6fSYLF2rETAvxVeJugxzn4pCBRD1MejTBZ4FGJeHenp93qTJzygckJRYoxpWbYRVCLazFtSpWB5gAQgqaqLzmnfx67F45QAxJZNiMUycD77Ncy9fv35QnbGJBqu3DpikkrKj8uYoZ1zWUAXunmGe5xAPTbhg4jxGVP3CP3XFoK68ZKGxA9\",\"CoinDetailsEncrypted\":null}",
		"{\"CoinDetails\":\"12qiFrYv3dA7evFnHqXWGhMtoDd1L5ctVBxvhNCQ99iG91jWTvPZ96hpy1nRSPPBU3MAH1qmCKvtZFuJ81smWm4PisrFp14dwkcSZ9WdfgeQmTNf653nzX1AXj4JJQSdoX8U1WN1HGRVduHFpzUy8QVuhERpYh2jQCwYrZKV8EX5RL15LPGy6XvSMoRpi1jcSCmXfM\",\"CoinDetailsEncrypted\":null}",
		"{\"CoinDetails\":\"12pzRynmiGot4hdzpQ7CzoZWWkCJFhZ9D9MHgnDj2HyvL4TaBrqkfvJGnJGvmdEphqn2WPAMcUZ2YTpB3DfAXteSukaWT3SQn2HmAozUZejZGkmPjnHMTxzVNTMj3waJp5JXkd7iRoZEiRxpimtJ7FzvPm6hY3ebzsun1uCA51Tkknsts1UmnfKRZp8BrHEvjMYA5V\",\"CoinDetailsEncrypted\":null}",
		"{\"CoinDetails\":\"12pNoLxUwAaJTwTgZFBAVL2RBguP8vdGqyK7hJb6XpwySKjSGFAMj25YcLSATJ2jMrWJw3nc1muL74vCpjtXC94W84dVkk9dNBLSmz5snYsMC5RzdkbRkZ95oKgAX43Q1EMtocRgnJo5jUvL2ro67sxUgaxr2ipiAsk2njBwzUGRg9UPaY22DkinYX8SEarAQbGL93\",\"CoinDetailsEncrypted\":null}",
		"{\"CoinDetails\":\"1ckY2i52cMGvFxQAyvqsY51TNUYbjGJGnwTMd5Hz7fNTiKbgjwVVr7nvFoggsWUsjycxpofW3rjJ1GLhsnPAbsMt7hSHdCtMuH2fFNKdjZFoZhTn7bJU9V5dUKwhRP4T9VnakSLU6Ap7WheZBZED4C4zF9u8nroQqWYHKGE4bKivPKFQSDK9TzQj6ikhVnknW9LifAWM\",\"CoinDetailsEncrypted\":null}",
	}
	outputCoins := make([]*coin.OutputCoin, len(outData))
	for i, data := range outData {
		outputCoins[i] = new(coin.OutputCoin)
		err := json.Unmarshal([]byte(data), outputCoins[i])
		assert.Equal(t, nil, err)
	}

	paymentData := []string{
		"{\"PaymentAddress\":{\"Pk\":\"eayrfPJK8y1n7NzzIJlbAgx2qZb+jLYIPuVQ8zWuo7k=\",\"Tk\":\"jOV/yfXWC1teu8AdByIJ5N5P1vu10EMmJK1IUvOmgTA=\"},\"Amount\":200000000000,\"Message\":null}",
		"{\"PaymentAddress\":{\"Pk\":\"0v+1asP3MBb8s6gun+Ew6uv+EbZVwUWa5juRn4c8mhs=\",\"Tk\":\"xrki1cpdNb8gRMKYZa5zOq5mke9EYo86uEIaBHwaQR4=\"},\"Amount\":200000000000,\"Message\":null}",
		"{\"PaymentAddress\":{\"Pk\":\"0jFS2BKzLcXOon4LZWJ8P3n2j6w2OWEDKprP4DCm6iQ=\",\"Tk\":\"l3QG3RtsHvMJ0E7pnz6BE8cYNgVZPGS0zZgPN2/4g54=\"},\"Amount\":200000000000,\"Message\":null}",
		"{\"PaymentAddress\":{\"Pk\":\"Ku1CbzjDmky2MYs4Oc3xPXrlpWYbHmLYqnoPwhtHF1Y=\",\"Tk\":\"rYdD15EPXC/gY9iaPBarfIBgT2KVxqyqno2AhOwUBRI=\"},\"Amount\":200000000000,\"Message\":null}",
		"{\"PaymentAddress\":{\"Pk\":\"Uwx9z/Q2U/K0diEituD3QZFcQm2e6Cnd5eG55lWMsPg=\",\"Tk\":\"zDKcdTvdEaAbuETHOV1oaAXq/wfL133GKHdV7wW0hDI=\"},\"Amount\":200000000000,\"Message\":null}",
		"{\"PaymentAddress\":{\"Pk\":\"/NPaKTA3pYTjYxdWpSvOlsp+hAzbwc6Q8x0qynd2+ao=\",\"Tk\":\"avALxL0vIbNCiJQxg41df7ml71CgzA/BL1M7kjmXiAQ=\"},\"Amount\":200000000000,\"Message\":null}",
		"{\"PaymentAddress\":{\"Pk\":\"/OSBuAJ4jsfR4Pjzzp/DLcl5yqfJRP/EMqQBh4PANi0=\",\"Tk\":\"nWi3+vcRAAXFSmgmYG+xxMkYY6WsflTVfZEynE1uVMg=\"},\"Amount\":200000000000,\"Message\":null}",
		"{\"PaymentAddress\":{\"Pk\":\"w9L7RKxapN3Z8UBsO8rtLn9/Jfa439w3JYTn6D+JpR8=\",\"Tk\":\"jTuZ4n6AEn/0SaI1pxHDlekm2ReHjifdpitdjlSlol0=\"},\"Amount\":200000000000,\"Message\":null}",
		"{\"PaymentAddress\":{\"Pk\":\"kzTvOZddX24XoN1KwFGgKKbNVho4ixNV43nngAYBYlA=\",\"Tk\":\"LfixpIcaHGk7fQdGeDEEZR48CJm/hIkaF9NM6DnKWuw=\"},\"Amount\":200000000000,\"Message\":null}",
		"{\"PaymentAddress\":{\"Pk\":\"5xVSzcZpA3uHmBO5ejENk13iayexILopySACdieLugA=\",\"Tk\":\"cz3hqWx+ow6++Eh2o0iup36c5HF5uJ4dgVY++Ggrf64=\"},\"Amount\":24998199999999993,\"Message\":null}",
	}
	paymentInfo := make([]*key.PaymentInfo, len(paymentData))
	for i, data := range paymentData {
		paymentInfo[i] = new(key.PaymentInfo)
		err := json.Unmarshal([]byte(data), outputCoins[i])
		assert.Equal(t, nil, err)
		b, err := base64.StdEncoding.DecodeString(string(paymentInfo[i].PaymentAddress.Tk))
		assert.Equal(t, nil, err)
		paymentInfo[i].PaymentAddress.Tk = b
	}

	proof, err := Prove(&inputCoins, &outputCoins, false, &paymentInfo)
	assert.Equal(t, nil, err)
	b := proof.Bytes()

	temp := new(PaymentProofV2)
	err = temp.SetBytes(b)
	b2 := temp.Bytes()
	assert.Equal(t, true, bytes.Equal(b2, b))
}
