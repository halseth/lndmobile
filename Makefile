PKG := github.com/halseth/lndmobile

MOBILE_PKG := $(PKG)/mobile
GO_BIN := ${GOPATH}/bin
GOMOBILE_BIN := GO111MODULE=off $(GO_BIN)/gomobile
MOBILE_BUILD_DIR :=${GOPATH}/src/$(MOBILE_PKG)/build
IOS_BUILD_DIR := $(MOBILE_BUILD_DIR)/ios
IOS_BUILD := $(IOS_BUILD_DIR)/Lndmobile.framework
ANDROID_BUILD_DIR := $(MOBILE_BUILD_DIR)/android
ANDROID_BUILD := $(ANDROID_BUILD_DIR)/Lndmobile.aar

mobile-rpc:
	@$(call print, "Creating mobile RPC from protos.")
	cd ./mobile; ./gen_bindings.sh

vendor:
	@$(call print, "Re-creating vendor directory.")
	rm -r vendor/; GO111MODULE=on go mod vendor

ios: vendor mobile-rpc
	@$(call print, "Building iOS framework ($(IOS_BUILD)).")
	mkdir -p $(IOS_BUILD_DIR)
	$(GOMOBILE_BIN) bind -target=ios -tags="ios $(DEV_TAGS) autopilotrpc experimental signrpc walletrpc chainrpc invoicesrpc routerrpc" $(LDFLAGS) -v -o $(IOS_BUILD) $(MOBILE_PKG)

android: vendor mobile-rpc
	@$(call print, "Building Android library ($(ANDROID_BUILD)).")
	mkdir -p $(ANDROID_BUILD_DIR)
	$(GOMOBILE_BIN) bind -target=android -tags="android $(DEV_TAGS) autopilotrpc experimental signrpc walletrpc chainrpc invoicesrpc routerrpc" $(LDFLAGS) -v -o $(ANDROID_BUILD) $(MOBILE_PKG)

mobile: ios android

.PHONY: mobile-rpc \
	vendor \
	ios \
	android \
	mobile \
