package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/nsf/termbox-go"
)

const (
	host       = "localhost" //"192.168.3.3"
	port       = "1586"
	clientIP   = "127.0.0.1" //"192.168.3.8"
	clientPort = "1515"
)

// ConnClient has a basic conversation functions for TCP connection.
type ConnClient struct {
	conn *net.TCPConn
}

// Ping send ping with certain interval.
func (c *ConnClient) Ping(done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		default:
			c.conn.Write([]byte("PING -1\r\n"))
			time.Sleep(360000 * time.Millisecond)
		}
	}
}

// Receive get message from server
func (c *ConnClient) Receive(done <-chan struct{}) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				return
			default:
				msg := make([]byte, 1024)
				readlen, err := c.conn.Read(msg)
				errCheck(err)
				out <- string(msg[:readlen])
			}
		}
	}()
	return out
}

// Send send message to server.
func (c *ConnClient) Send(msg string) error {
	// TODO: implement
	_, err := c.conn.Write([]byte("msg"))
	return err
}

func main() {
	// Set termbox
	if err := termbox.Init(); err != nil {
		fmt.Println("ERROR: Cannot initialize termbox")
		os.Exit(1)
	}
	defer termbox.Close()
	var eb EditBox
	termbox.SetInputMode(termbox.InputEsc)
	eb.Draw()
	termbox.SetCell(1, 1, 't', termbox.ColorDefault, termbox.ColorDefault)
	termbox.Flush()

	// Set TCP connection
	tcpAddr, err := net.ResolveTCPAddr("tcp", host+":"+port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return
	}

	clientAddr := new(net.TCPAddr)
	clientAddr.IP = net.ParseIP(clientIP)
	clientAddr.Port, _ = strconv.Atoi(clientPort)

	conn, err := net.DialTCP("tcp", clientAddr, tcpAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return
	}

	done := make(chan struct{})
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(400 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(400 * time.Second))
	c := &ConnClient{conn: conn}
	go c.Ping(done)
	//input := scan(done)
	response := c.Receive(done)
	keyInput := Input(done)

	for {
		select {
		case k := <-keyInput:
			switch k.Type {
			case termbox.EventKey:
				switch k.Key {
				case termbox.KeyArrowRight, termbox.KeyCtrlF:
					eb.MoveCursorOneRuneForward()
				case termbox.KeyArrowLeft, termbox.KeyCtrlB:
					eb.MoveCursorOneRuneBackward()
				case termbox.KeyBackspace, termbox.KeyBackspace2:
					eb.DeleteRuneBackward()
				case termbox.KeyEnter:
					bmsg := eb.GetAndDeleteText()
					message := string(bmsg[:])
					if message == "quit" {
						fmt.Println("OK quit")
						done <- struct{}{}
						return
					}
					// Server require new line character
					message += "\r\n"
					conn.Write([]byte(message))
				case termbox.KeyEsc:
					done <- struct{}{}
					return
				case termbox.KeySpace:
					r, _ := utf8.DecodeLastRune([]byte(" "))
					eb.InsertRune(r)
				default:
					eb.InsertRune(k.Ch)
				}
			case termbox.EventError:
				done <- struct{}{}
				return
			}
		case responseMsg := <-response:
			testConversation[0] = "Server response: " + responseMsg
			switch responseMsg {
			case "OK PING\r\n":
				conn.SetReadDeadline(time.Now().Add(400 * time.Second))
				conn.SetWriteDeadline(time.Now().Add(400 * time.Second))
			}
		default:
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			eb.Draw()
		}

	}

	// for {
	// 	select {
	// 	case message := <-input:
	// 		if strings.TrimSuffix(message, "\n\r\n") == "quit" {
	// 			fmt.Println("OK quit")
	// 			done <- struct{}{}
	// 			return
	// 		}
	// 		conn.Write([]byte(message))
	// 	case responseMsg := <-response:
	// 		fmt.Println("Server response: " + responseMsg)
	// 		switch responseMsg {
	// 		case "OK PING\r\n":
	// 			conn.SetReadDeadline(time.Now().Add(400 * time.Second))
	// 			conn.SetWriteDeadline(time.Now().Add(400 * time.Second))
	//              }
	// 	}
	// }

}

func scan(done chan struct{}) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		select {
		case <-done:
			return
		default:
			for {
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				input += "\r\n"
				out <- input
			}
		}
	}()
	return out
}

func errCheck(err error) {
	if err != nil {
		if err.Error() == io.EOF.Error() {
			fmt.Fprintf(os.Stdout, "End connection\n")
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}
