package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

var (
	listenAddr = flag.String("l", ":8080", "local listen port")
	dstAddr    = flag.String("d", "localhost:7070", "proxy to dest addr")
	bufferPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 256*1024)
		},
	}
	dialer = net.Dialer{
		KeepAlive: time.Minute * 5,
		Timeout:   time.Second * 5,
	}
)

func handleConn(conn net.Conn) {
	defer conn.Close()
	log.Println("connected from:", conn.RemoteAddr())
	dst, err := dialer.Dial("tcp", *dstAddr)
	if err != nil {
		log.Println(err)
		return
	}
	defer dst.Close()
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func(wg *sync.WaitGroup) {
		buf := bufferPool.Get()
		_, err = io.CopyBuffer(conn, dst, buf.([]byte))
		if err != nil {
			log.Println(err)
		}
		bufferPool.Put(buf)
		wg.Done()
	}(wg)
	go func(wg *sync.WaitGroup) {
		buf := bufferPool.Get()
		_, err = io.CopyBuffer(dst, conn, buf.([]byte))
		if err != nil {
			log.Println(err)
		}
		bufferPool.Put(buf)
		wg.Done()
	}(wg)
	wg.Wait()
}

func main() {
	flag.Parse()
	log.Printf("%s listen at :%s and proxy to %s\n", os.Args[0], *listenAddr, *dstAddr)
	ln, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConn(conn)
	}
}
