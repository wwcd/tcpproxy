package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/wwcd/tcpproxy/pkg/tcpproxy"
)

var (
	target          string
	port            int
	sourceCipherKey string
	targetCipherKey string
)

func init() {
	flag.StringVar(&target, "target", "", "target (<host>:<port>)")
	flag.IntVar(&port, "port", 1337, "port")
	flag.StringVar(&sourceCipherKey, "source_cipher_key", "", "source cipher key")
	flag.StringVar(&targetCipherKey, "target_cipher_key", "", "destination cipher key")
}

func main() {
	flag.Parse()

	if target == "" {
		fmt.Fprintf(os.Stderr, "usage:\n")
		fmt.Fprintf(os.Stderr, "  %s -target IP:PORT\n", os.Args[0])
		return
	}

	signals := make(chan os.Signal, 1)
	stop := make(chan bool)
	signal.Notify(signals, os.Interrupt)
	go func() {
		for range signals {
			fmt.Println("\nReceived an interrupt, stopping...")
			stop <- true
		}
	}()

	p := tcpproxy.New(target, tcpproxy.WithDestinationKey(targetCipherKey), tcpproxy.WithSourceKey(sourceCipherKey))

	p.ListenAndServe(fmt.Sprintf(":%d", port))

	<-stop
}
