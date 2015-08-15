package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

var (
	listenAddr = flag.String("l", ":8080", "local listen port")
	dstAddr    = flag.String("d", "localhost:7070", "proxy to dest addr")
)

func handleConn(conn net.Conn) {
	log.Println("connected from:", conn.RemoteAddr())
	defer conn.Close()
	dst, err := net.Dial("tcp", *dstAddr)
	if err != nil {
		log.Println(err)
		return
	}
	defer dst.Close()
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func(wg *sync.WaitGroup) {
		_, err = io.Copy(conn, dst)
		if err != nil {
			log.Println(err)
		}
		wg.Done()
	}(wg)
	go func(wg *sync.WaitGroup) {
		_, err = io.Copy(dst, conn)
		if err != nil {
			log.Println(err)
		}
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
