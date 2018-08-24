package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func auth(conn net.Conn) {
	// VER	NMETHODS	METHODS
	// 1	1			1-255

	// VER是SOCKS版本，这里应该是0x05；
	// NMETHODS是METHODS部分的长度；
	// METHODS是客户端支持的认证方式列表，每个方法占1字节。当前的定义是：
	// 0x00 不需要认证
	// 0x01 GSSAPI
	// 0x02 用户名、密码认证
	// 0x03 - 0x7F由IANA分配（保留）
	// 0x80 - 0xFE为私人方法保留
	// 0xFF 无可接受的方法

	res := make([]byte, 2)
	_, err := conn.Read(res)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res)
	methodLength := res[1]
	method := make([]byte, methodLength)
	conn.Read(method)
	fmt.Println(method)

	// 服务器从客户端提供的方法中选择一个并通过以下消息通知客户端（以字节为单位）：
	//
	// VER	METHOD
	// 1	1
	// VER是SOCKS版本，这里应该是0x05；
	// METHOD是服务端选中的方法。如果返回0xFF表示没有一个认证方法被选中，客户端需要关闭连接。

	// REP应答字段
	// 0x00表示成功
	// 0x01普通SOCKS服务器连接失败
	// 0x02现有规则不允许连接
	// 0x03网络不可达
	// 0x04主机不可达
	// 0x05连接被拒
	// 0x06 TTL超时
	// 0x07不支持的命令
	// 0x08不支持的地址类型
	// 0x09 - 0xFF未定义
	resp := []byte{5, 0}
	conn.Write(resp)
}

func buildDestConn(conn net.Conn) net.Conn {
	// VER	CMD	RSV		ATYP	DST.ADDR	DST.PORT
	// 1	1	0x00	1		动态		2
	// VER是SOCKS版本，这里应该是0x05；
	// CMD是SOCK的命令码
	// 0x01表示CONNECT请求
	// 0x02表示BIND请求
	// 0x03表示UDP转发
	// RSV 0x00，保留
	// ATYP DST.ADDR类型
	// 0x01 IPv4地址，DST.ADDR部分4字节长度
	// 0x03 域名，DST.ADDR部分第一个字节为域名长度，DST.ADDR剩余的内容为域名，没有\0结尾。
	// 0x04 IPv6地址，16个字节长度。
	// DST.ADDR 目的地址
	// DST.PORT 网络字节序表示的目的端口

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
	ln, err := net.Listen("tcp", ":8081")
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
