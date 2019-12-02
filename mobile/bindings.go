package lndmobile

import (
	"fmt"
	"os"
	"strings"

	flags "github.com/jessevdk/go-flags"
	"github.com/lightninglabs/loop/loopd"
	"github.com/lightningnetwork/lnd"
)

// Start starts lnd and loop in new goroutines.
//
// extraArgs can be used to pass command line arguments to lnd that will
// override what is found in the config file. Example:
//	extraArgs = "--bitcoin.testnet --lnddir=\"/tmp/folder name/\" --profile=5050"
//
// The unlockerReady callback is called when the WalletUnlocker service is
// ready, and rpcReady is called after the wallet has been unlocked and lnd is
// ready to accept RPC calls.
func Start(lndArgs, loopArgs string, unlockerReady, rpcReady Callback) {
	// Add the extra arguments to os.Args, as that will be parsed during
	// startup.
	osArgs := copyArgs(os.Args)
	os.Args = append(os.Args, splitArgs(lndArgs)...)

	// Set up channels that will be notified when the RPC servers are ready
	// to accept calls.
	var (
		unlockerListening = make(chan struct{})
		rpcListening      = make(chan struct{})
	)

	// We call the main method with the custom in-memory listeners called
	// by the mobile APIs, such that the grpc server will use these.
	cfg := lnd.ListenerCfg{
		WalletUnlocker: &lnd.ListenerWithSignal{
			Listener: walletUnlockerLis,
			Ready:    unlockerListening,
		},
		RPCListener: &lnd.ListenerWithSignal{
			Listener: lightningLis,
			Ready:    rpcListening,
		},
	}

	// Call the "real" main in a nested manner so the defers will properly
	// be executed in the case of a graceful shutdown.
	go func() {
		if err := lnd.Main(cfg); err != nil {
			if e, ok := err.(*flags.Error); ok &&
				e.Type == flags.ErrHelp {
			} else {
				fmt.Fprintln(os.Stderr, err)
			}
			os.Exit(1)
		}
	}()

	// Spin up a go routine for the loop daemon.
	go func() {
		<-rpcListening

		// Get a connection to the lnd instance we just started. Since
		// loop is aware of the macaroons and TLS certificate required
		// by lnd, we can give it the raw listener without any added
		// authentication options.
		lndConn, err := lightningLis.Dial()
		if err != nil {
			rpcReady.OnError(err)
			return
		}

		// Start the swap client itself.
		lisCfg := loopd.RpcConfig{
			RPCListener: swapClientLis,
			LndConn:     lndConn,
		}

		// Set the command line arguments to the copy we created
		// earlier, with the added loop arguments.
		os.Args = append(osArgs, splitArgs(loopArgs)...)

		err = loopd.Start(lisCfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	// Finally we start two go routines that will call the provided
	// callbacks when the RPC servers are ready to accept calls.
	go func() {
		<-unlockerListening
		unlockerReady.OnResponse([]byte{})
	}()

	go func() {
		<-rpcListening

		// Now that the RPC server is ready, we can get the needed
		// authentication options, and add them to the global dial
		// options.
		auth, err := lnd.Authenticate()
		if err != nil {
			rpcReady.OnError(err)
			return
		}

		// Add the auth options to the listener's dial options.
		addLightningLisDialOption(auth...)

		rpcReady.OnResponse([]byte{})
	}()
}

func copyArgs(args []string) []string {
	c := make([]string, len(args))
	for i, a := range args {
		c[i] = a
	}
	return c
}

func splitArgs(args string) []string {
	// Split the argument string on "--" to get separated command line
	// arguments.
	var splitArgs []string
	for _, a := range strings.Split(args, "--") {
		if a == "" {
			continue
		}
		// Finally we prefix any non-empty string with --, and trim
		// whitespace to mimic the regular command line arguments.
		splitArgs = append(splitArgs, strings.TrimSpace("--"+a))
	}

	return splitArgs
}
