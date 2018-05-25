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
			client, err := incoming.Accept()
			if err != nil {
				log.Fatal("could not accept client connection", err)
			}
			fmt.Printf("client '%v' connected!\n", client.RemoteAddr())

			target, err := net.Dial("tcp", target)
			if err != nil {
				log.Fatal("could not connect to target", err)
			}
			fmt.Printf("connection to server %v established!\n", target.RemoteAddr())

			go func() {
				c := make(chan struct{})
				go func() { io.Copy(client, target); c <- struct{}{} }()
				go func() { io.Copy(target, client); c <- struct{}{} }()
				<-c
				client.Close()
				target.Close()
				fmt.Printf("disconnect client %v and target %v\n", client.RemoteAddr(), target.RemoteAddr())
			}()
		}
	}()

	<-stop
}
