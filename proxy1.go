package main

import (
    "io"
	"fmt"
    "log"
    "net"
	// "net/http"

    "github.com/valyala/fasthttp"
)

func transfer(destination io.WriteCloser, source io.ReadCloser) {
    fmt.Println("transfer")
    defer destination.Close()
    defer source.Close()
    fmt.Println("-------------------")
    io.Copy(destination, source)
    fmt.Println("===================")
}

func main() {
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    requestHandler := func(ctx *fasthttp.RequestCtx) {
        fmt.Println("host: ", string(ctx.Host()))
        dest_conn, err := net.Dial("tcp", string(ctx.Host()), )
        if err != nil {
            log.Println(err)
            return
        }

        hijackHandler := func(client_conn net.Conn) {
            fmt.Println("333333")
            go transfer(dest_conn, client_conn)
            go transfer(client_conn, dest_conn)
        }

        ctx.Hijack(hijackHandler)
    }

    h := func(ctx *fasthttp.RequestCtx) {
        fmt.Println(string(ctx.Method()))
        fmt.Println("123")
        // ctx.SetBody([]byte("this is completely new body contents"))
        // ctx.SetStatusCode(200)
        // fmt.Fprintln(ctx, "Connection Established\r\n\r\n")
        fmt.Println(ctx.Request.String())
        fmt.Println(ctx.Response.String())
        ctx.Response.Header.SetContentLength(-1)
        requestHandler(ctx)
        ctx.Response.Header.SetContentLength(-1)
        fmt.Println("content length", ctx.Response.Header.ContentLength())
        fmt.Println("321")
    }

    s := &fasthttp.Server{
        // Handler: requestHandler,
        Handler: h,
    }
    if err := s.ListenAndServe("127.0.0.1:8080"); err != nil {
        log.Fatalf("error in ListenAndServe: %s", err)
    }
}
