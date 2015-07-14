package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

const (
	host       = "192.168.3.3"
	port       = "1586"
	clientIP   = "192.168.3.8"
	clientPort = "1586"
)

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", host+":"+port)
	errCheck(err)

	clientAddr := new(net.TCPAddr)
	clientAddr.IP = net.ParseIP(clientIP)
	clientAddr.Port, _ = strconv.Atoi(clientPort)

	conn, err := net.DialTCP("tcp", clientAddr, tcpAddr)
	errCheck(err)

	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(100 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(100 * time.Second))
	for {
		readBuf := make([]byte, 1024)
		readlen, err := conn.Read(readBuf)
		errCheck(err)
		fmt.Println("Server response: " + string(readBuf[:readlen]))

		message := scan()
		conn.Write([]byte(message))
	}

}

func scan() string {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input += "\r\n"
	return input
}

func errCheck(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}
