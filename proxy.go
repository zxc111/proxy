package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"golang.org/x/net/http2"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
)

type targetInfo struct {
	host string
	port string
}

var (
	proxy string
	local string
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var (
		localHost string
		localPort string
		proxyHost string
		proxyPort string
	)
	flag.StringVar(&localHost, "local_host", "", "-local_host=127.0.0.1")
	flag.StringVar(&localPort, "local_port", "", "-local_port=4000")
	flag.StringVar(&proxyHost, "proxy_host", "", "-porxy_host=xxx.xxx.xxx.xxx")
	flag.StringVar(&proxyPort, "proxy_port", "", "-proxy_port=3000")

	flag.Parse()
	if proxyHost == "" {
		flag.Usage()
		log.Println("proxy_host is required")
		os.Exit(1)
	}
	if proxyPort == "" {
		log.Println("proxy_port is required")
		flag.Usage()

		os.Exit(1)
	}
	if localHost == "" {
		log.Println("local_host is required")
		flag.Usage()

		os.Exit(1)
	}
	if localPort == "" {
		log.Println("local_port is requred")
		flag.Usage()

		os.Exit(1)
	}
	proxy = fmt.Sprintf("%s:%s", proxyHost, proxyPort)
	local = fmt.Sprintf("%s:%s", localHost, localPort)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	_, err := io.Copy(destination, source)
	if err != nil {
		log.Println(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodConnect:
		hijacker, _ := w.(http.Hijacker)
		client_conn, _, err := hijacker.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		fmt.Fprint(client_conn, "HTTP/1.1 200 Connection Established\r\n\r\n")
		w.WriteHeader(200)

		remote := "http://" + r.URL.Host
		toHttpProxy(client_conn, remote)
	default:
		remote := r.URL.Scheme + "://" + r.URL.Host
		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Println(err)
		}
		fmt.Println(string(dump))
		//newR := bytes.NewReader(dump)
		//net.Conn(newR)

		hijacker, _ := w.(http.Hijacker)
		client_conn, _, err := hijacker.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		GettoHttpProxy(r, remote, client_conn)
	}
}

func toHttpProxy(from net.Conn, remoteAddr string) {

	tr := &http2.Transport{
		DialTLS: func(network, addr string, config *tls.Config) (net.Conn, error) {
			return tls.Dial("tcp", proxy, &tls.Config{
				NextProtos:         []string{"h2"},
				InsecureSkipVerify: true,
			})
		},
		AllowHTTP: true,
	}

	r, w := io.Pipe()

	//remoteAddr := "http://216.58.200.14:443"
	log.Println(remoteAddr)
	req, err := http.NewRequest(
		http.MethodConnect,
		remoteAddr,
		r,
	)

	if err != nil {
		log.Println(err)
	}
	resp, err := tr.RoundTrip(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println(resp.StatusCode)
		io.Copy(os.Stdout, resp.Body)
		log.Println("Connect Proxy Server Error")
		return
	}
	//tmp := make([]byte, 1024)
	//from.Read(tmp)
	//fmt.Println(tmp)
	go io.Copy(w, from)
	io.Copy(from, resp.Body)
}

func GettoHttpProxy(from *http.Request, remote string, to net.Conn) {

	dump, err := httputil.DumpRequest(from, true)
	if err != nil {
		log.Println(err)
	}
	r, w := io.Pipe()

	tr := &http2.Transport{
		DialTLS: func(network, addr string, config *tls.Config) (net.Conn, error) {
			return tls.Dial("tcp",
				proxy,
				&tls.Config{
					NextProtos:         []string{"h2"},
					InsecureSkipVerify: true,
				})
		},
		AllowHTTP: true,
	}

	remoteAddr := remote
	log.Println(remoteAddr)

	req, err := http.NewRequest(
		http.MethodConnect,
		remoteAddr,
		r,
	)

	//req.Header = from.Header
	if err != nil {
		log.Println(err)
	}
	resp, err := tr.RoundTrip(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println(resp.StatusCode)
		io.Copy(os.Stdout, resp.Body)
		log.Println("Connect Proxy Server Error")
		return
	}
	// from.
	go io.Copy(w, bytes.NewReader(dump))
	//result := make([]byte, 1024)
	//resp.Body.Read(result)
	//fmt.Println(string(result))
	io.Copy(to, resp.Body)

}
func main() {

	log.Printf("local: %s", local)
	log.Printf("remote: %s", proxy)
	server := &http.Server{
		Addr: local,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//fmt.Println("connect 1")
			switch r.Method {
			case http.MethodConnect:

				fmt.Println("connect")
				handler(w, r)
			default:

				handler(w, r)
			}
		}),
	}
	server.ListenAndServe()
}
