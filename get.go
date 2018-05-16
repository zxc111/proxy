package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
)

func buildReq() {
	req := &http.Request{
		Method: "CONNECT",
		URL: &url.URL{
			Scheme: "http", // of proxy.com
			Host:   "proxy.com",
			Opaque: "backend:443",
		},
	}

}

func auth(conn net.Conn) {
	res := make([]byte, 1024)
	conn.Read(res)
	fmt.Println(res)
	resp := []byte{5, 0}
	conn.Write(resp)
}

func buildDestConn(conn net.Conn) net.Conn {

	res := make([]byte, 1024)
	conn.Read(res)
	fmt.Println(res)
	if res[3] != 1 {
		log.Fatal("addr error")
	}
	addr := fmt.Sprintf("%d.%d.%d.%d", res[4], res[5], res[6], res[7])
	port := int(res[8])*256 + int(res[9])
	target := fmt.Sprintf("%s:%d", addr, port)

	resp := []byte{5, 0, 0, res[3], res[4], res[5], res[6], res[7], res[8], res[9]}
	conn.Write(resp)
	destConn, err := net.Dial("tcp", target)
	if err != nil {
		log.Fatal(err)
	}
	return destConn
}
func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	_, err := io.Copy(destination, source)
	if err != nil {
		log.Println(err)
	}
}

func handleConnection(conn net.Conn) {
	auth(conn)
	dest := buildDestConn(conn)
	go transfer(conn, dest)
	go transfer(dest, conn)
}

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn)
	}
}
