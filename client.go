package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
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
	conn net.Conn
}

// Ping send ping with certain interval.
func (c *ConnClient) Ping(done <-chan struct{}) {
	waitSig := time.Tick(360 * time.Second)
	for {
		select {
		case <-done:
			return
		case <-waitSig:
			c.conn.Write([]byte("PING -1\r\n"))
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
				if err != nil {
					if err.Error() == io.EOF.Error() {
						out <- "quit"
						return
					}
				}
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
	file, err := os.OpenFile("./log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("error opening file :", err.Error())
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(file)

	// Set termbox
	if err := termbox.Init(); err != nil {
		fmt.Println("ERROR: Cannot initialize termbox")
		os.Exit(1)
	}
	defer termbox.Close()
	// Set screens
	eb := &EditBox{}
	ws := &WholeScreen{}
	connMsg := NewTextBox(20)
	ts := &TextScreen{}
	ts.SetTextArea(connMsg)
	ws.append(eb)
	ws.append(ts)

	// Draw initial screen
	termbox.SetInputMode(termbox.InputEsc)
	ws.drawAll()
	//eb.Draw()
	//termbox.Flush()

	log.Println("Start TCP setting")
	// Set TCP connection
	tcpAddr, err := net.ResolveTCPAddr("tcp", host+":"+port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return
	}

	clientAddr := new(net.TCPAddr)
	clientAddr.IP = net.ParseIP(clientIP)
	clientAddr.Port, _ = strconv.Atoi(clientPort)

	log.Println("Start TCP dial")
	conn, err := net.DialTCP("tcp", clientAddr, tcpAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return
	}

	done := make(chan struct{})
	defer close(done)
	defer conn.Close()

	log.Println("Set TCP conn deadline")
	conn.SetReadDeadline(time.Now().Add(400 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(400 * time.Second))

	c := &ConnClient{conn: conn}

	log.Println("Start sending PING message")
	go c.Ping(done)
	//input := scan(done)

	log.Println("Start receiving message")
	response := c.Receive(done)

	log.Println("Start getting keyboard inputs")
	keyInput := Input(done)

	log.Println("Start main loop")
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
						log.Println("Exit by quit signal from keyboard input")
						done <- struct{}{}
						return
					}
					// Server require new line character
					message += "\r\n"
					conn.Write([]byte(message))
				case termbox.KeyEsc:
					log.Println("Exit by KeyEsc signal")
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
			connMsg.AppendText("Server response: " + responseMsg)
			switch responseMsg {
			case "quit":
				log.Println("Exit by quit signal from server message")
				done <- struct{}{}
				return
			case "OK PING\r\n":
				c.conn.SetReadDeadline(time.Now().Add(400 * time.Second))
				c.conn.SetWriteDeadline(time.Now().Add(400 * time.Second))
			case "SVR_PING\r\n":
				c.Send("OK SVR_PING")
			}
		default:
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			ws.drawAll()
		}
	}
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
