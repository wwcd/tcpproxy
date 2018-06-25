package tcpproxy

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net"
)

type TCPProxy struct {
	target string
	sKey   string
	dKey   string
}

type Option func(*TCPProxy)

func WithSourceKey(key string) Option {
	return func(p *TCPProxy) { p.sKey = key }
}

func WithDestinationKey(key string) Option {
	return func(p *TCPProxy) { p.dKey = key }
}

func New(target string, opts ...Option) *TCPProxy {
	p := &TCPProxy{target: target}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *TCPProxy) ListenAndServe(addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("could not start server on %s: %v", addr, err)
	}
	fmt.Printf("server running on %s\n", addr)

	go p.Serve(ln)
}

func (p *TCPProxy) Serve(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal("could not accept client connection", err)
		}
		fmt.Printf("client '%v' connected!\n", conn.RemoteAddr())

		go p.handler(conn)
	}
}

func (p *TCPProxy) handler(conn net.Conn) {
	defer conn.Close()

	target, err := net.Dial("tcp", p.target)
	if err != nil {
		log.Fatal("could not connect to target", err)
	}
	defer target.Close()

	fmt.Printf("connection to server %v established!\n", target.RemoteAddr())

	src := NewCipherStream(conn, p.sKey)
	dst := NewCipherStream(target, p.dKey)

	go io.Copy(dst.W, src.R)
	io.Copy(src.W, dst.R)

	fmt.Printf("disconnect client %v and target %v\n", conn.RemoteAddr(), target.RemoteAddr())
}

type CipherStream struct {
	R io.Reader
	W io.Writer
}

func NewCipherStream(rw io.ReadWriter, password string) *CipherStream {
	if password == "" {
		return &CipherStream{R: rw, W: rw}
	}

	hasher := md5.New()
	hasher.Write([]byte(password))
	key := hasher.Sum(nil)

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal("could not create cipher", err)
	}

	iv := make([]byte, aes.BlockSize)

	return &CipherStream{
		R: &cipher.StreamReader{S: cipher.NewCFBDecrypter(block, iv), R: rw},
		W: &cipher.StreamWriter{S: cipher.NewCFBEncrypter(block, iv), W: rw},
	}
}
