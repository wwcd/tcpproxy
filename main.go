package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
)

var (
	target string
	port   int
)

func init() {
	flag.StringVar(&target, "target", "", "target (<host>:<port>)")
	flag.IntVar(&port, "port", 1337, "port")
}

func main() {
	flag.Parse()

	signals := make(chan os.Signal, 1)
	stop := make(chan bool)
	signal.Notify(signals, os.Interrupt)
	go func() {
		for _ = range signals {
			fmt.Println("\nReceived an interrupt, stopping...")
			stop <- true
		}
	}()

	incoming, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("could not start server on %d: %v", port, err)
	}
	fmt.Printf("server running on %d\n", port)

	go func() {
		for {
			conn, err := incoming.Accept()
			if err != nil {
				log.Fatal("could not accept client connection", err)
			}
			fmt.Printf("client '%v' connected!\n", conn.RemoteAddr())

			go func(conn net.Conn) {
				defer conn.Close()

				target, err := net.Dial("tcp", target)
				if err != nil {
					log.Fatal("could not connect to target", err)
				}
				defer target.Close()

				fmt.Printf("connection to server %v established!\n", target.RemoteAddr())

				var in, out int64
				go func() { in, _ = io.Copy(target, conn) }()
				out, _ = io.Copy(conn, target)
				fmt.Printf("disconnect client %v and target %v, in %d, out %d\n", conn.RemoteAddr(), target.RemoteAddr(), in, out)
			}(conn)
		}
	}()

	<-stop
}
