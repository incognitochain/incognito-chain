install-privacy:
	@cd ./privacy/server && sh ./build_macos.sh;

start-privacy:
ifeq (,$(wildcard ./privacy/server/build/main))
	@echo Please run make install-privacy first
else
	@cd ./privacy/server/build && ./main
endif
