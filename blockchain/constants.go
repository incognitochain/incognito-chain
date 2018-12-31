package blockchain

// constant for network
const (
	//Network fixed params
	ThresholdRatioOfDCBCrisis = 90
	ThresholdRatioOfGovCrisis = 90

	// Mainnet
	Mainnet                           = 0x01
	MainetName                        = "mainnet"
	MainnetDefaultPort                = "9333"
	MainnetInitFundSalary             = 0
	MainnetInitDCBToken               = 0
	MainnetInitGovToken               = 0
	MainnetInitCmBToken               = 0
	MainnetInitBondToken              = 0
	MainnetGenesisblockPaymentAddress = "1UuyYcHgVFLMd8Qy7T1ZWRmfFvaEgogF7cEsqY98ubQjoQUy4VozTqyfSNjkjhjR85C6GKBmw1JKekgMwCeHtHex25XSKwzb9QPQ2g6a3"

	// Testnet
	Testnet               = 0x02
	TestnetName           = "testnet"
	TestnetDefaultPort    = "9444"
	TestnetInitFundSalary = 1000000000000000
	TestnetInitDCBToken   = 10000
	TestnetInitGovToken   = 10000

	//board and proposal parameters
	TestnetInitCmBToken               = 10000
	TestnetInitBondToken              = 10000
	TestnetGenesisBlockPaymentAddress = "1Uv3jP4ixNx3BkEtmUUxKXA1TXUduix3KMCWXHvLqVyA9CFfoLRZ949zTBNqDUPSzaPCZPrQKSfiEHguFazK6VeDmEk1RMLfX1kQiSqJ6"
)

const (
	// BlockVersion is the current latest supported block version.
	BlockVersion = 1
)
