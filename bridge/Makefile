all: swap burn

swap: incognito_proxy/incognito_proxy.abi vault/vault.abi
	go test -run=TestSimulatedSwapBeacon

burn: incognito_proxy/incognito_proxy.abi vault/vault.abi
	go test -run=TestSimulatedBurn

.PHONY: all swap burn

incognito_proxy/incognito_proxy.abi: incognito_proxy/incognito_proxy.vy
	./gengo.sh incognito_proxy/incognito_proxy.vy incognito_proxy

vault/vault.abi: vault/vault.vy
	./gengo.sh vault/vault.vy vault

