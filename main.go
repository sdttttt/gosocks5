package main

import (
	"io"
	"log"
	"net"
	"strconv"
)

func socks5Proxy(conn net.Conn) {
	var buf [8192]byte

	n, err := conn.Read(buf[:])
	if err != nil {
		log.Println(err)
		return
	}
	if buf[0] == 0x05 {
		conn.Write([]byte{0x05, 0x00})
		n, err = conn.Read(buf[:])
	}

	var host, port string

	switch buf[3] {
	case 0x01: //IP V4
		host = string(buf[4:7])
	case 0x03: //域名
		host = string(buf[5 : n-2]) //b[4]表示域名的长度
	case 0x04: //IP V6
		host = string((buf[4:19]))
	}
	port = strconv.Itoa(int(buf[n-2])<<8 | int(buf[n-1]))

	server, err := net.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		log.Println(err)
		return
	}

	conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) //响应客户端连接成功
	println(conn.RemoteAddr().String() + "<=>" + server.RemoteAddr().String())

	go func() {
		io.Copy(server, conn)
		conn.Close()
	}()

	io.Copy(conn, server)
	defer func() {
		server.Close()
	}()
}

func main() {
	server, err := net.Listen("tcp", ":10086")
	if err != nil {
		log.Panic(err)
	}
	defer server.Close()
	for {
		client, err := server.Accept()
		if err != nil {
			log.Println(err)
			return
		}
		go socks5Proxy(client)
	}
}
