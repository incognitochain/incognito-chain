all: incognito_proxy.abi
	go test -run=TestSwap

incognito_proxy.abi: incognito_proxy.vy
	./gengo.sh incognito_proxy.vy

