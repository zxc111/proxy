package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
)

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	_, err := io.Copy(destination, source)
	if err != nil {
		log.Println(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	dest_conn, err := net.Dial("tcp", r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	fmt.Fprintln(w, "HTTP/1.1 200 Connection Established\r\n\r\n")
	hijacker, _ := w.(http.Hijacker)
	client_conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	go transfer(dest_conn, client_conn)
	go transfer(client_conn, dest_conn)
}
func main() {
	server := &http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("connect 1")
			if r.Method == http.MethodConnect {
				fmt.Println("connect")
				handler(w, r)
			} else {
				fmt.Println(r.Method)
			}
		}),
	}
	server.ListenAndServe()
}
