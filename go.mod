module github.com/halseth/lndmobile

go 1.13

require (
	github.com/golang/protobuf v1.3.2
	github.com/jessevdk/go-flags v1.4.0
	github.com/lightninglabs/loop v0.3.0-alpha
	github.com/lightningnetwork/lnd v0.8.0-beta-rc3.0.20191115230031-4d7a151b4763
	google.golang.org/grpc v1.25.1
)

replace github.com/lightningnetwork/lnd => github.com/halseth/lnd v0.1.1-alpha.0.20191202123139-f0586e3ee074
