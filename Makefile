#!make
include .env
export

install-privacy:
	@cd ./privacy/server && sh ./build_macos.sh;

start-privacy:
ifeq (,$(wildcard ./privacy/server/build/main))
	@echo Please run make install-privacy first
else
	@cd ./privacy/server/build && ./main
endif

start-nodes:
	@sh bin/start-local-nodes.sh $$TOTAL_OF_NODE $$DATA_DIR

stop-nodes:
	@sh bin/clear-local-nodes.sh
