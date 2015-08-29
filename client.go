package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nsf/termbox-go"
)

const (
	host       = "localhost" //"192.168.3.3"
	port       = "1586"
	clientIP   = "127.0.0.1" //"192.168.3.8"
	clientPort = "1517"
)

// Mode switches input and output style.
type Mode string

const (
	// RoomMode is used for identifing mode to controll rooms.
	RoomMode = "Room"
	// ChatMode is used for identifing mode to controll chat.
	ChatMode = "Chat"
	// DirectMode is used for sending message directly.
	DirectMode = "Direct"
)

// ConnClient has a basic conversation functions for TCP connection.
type ConnClient struct {
	conn net.Conn
	mode Mode
}

// Ping send ping with certain interval.
func (c *ConnClient) Ping(done <-chan struct{}) {
	waitSig := time.Tick(360 * time.Second)
	for {
		select {
		case <-done:
			return
		case <-waitSig:
			c.Send("PING -1")
			c.conn.SetReadDeadline(time.Now().Add(400 * time.Second))
			c.conn.SetWriteDeadline(time.Now().Add(400 * time.Second))
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
				log.Println("Server responce: " +
					string(msg[:readlen]))
				out <- string(msg[:readlen])
			}
		}
	}()
	return out
}

// Send send message to server.
func (c *ConnClient) Send(msg string) error {
	// TODO: implement
	var arrangedMsg string
	switch c.mode {
	case DirectMode:
		arrangedMsg = msg
	case RoomMode:
		// Get parameter like "OPEN 1".
		msgTokens := strings.Split(msg, " ")
		switch msgTokens[0] {
		case "OPEN":
			arrangedMsg = "OPEN_ROOM " + msgTokens[1]
		case "CLOSE":
			arrangedMsg = "CLOSE_ROOM " + msgTokens[1]
		}
	case ChatMode:
		// Get parameter like "1 some message"
		msgTokens := strings.SplitAfterN(msg, " ", 2)
		arrangedMsg = "SHOUT " + msgTokens[0] + " " + msgTokens[1]

	}
	_, err := c.conn.Write([]byte((arrangedMsg + "\r\n")[:]))

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
	roomList := NewRoomBox(20)
	chatLogs := NewChatBox(20)

	ts := &TextScreen{}
	ts.SetTextArea(connMsg)
	ws.append(eb)
	ws.append(ts)

	// Draw initial screen
	termbox.SetInputMode(termbox.InputEsc)
	ws.drawAll()

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

	c := &ConnClient{conn: conn, mode: DirectMode}

	log.Println("Start sending PING message")
	go c.Ping(done)
	//input := scan(done)

	log.Println("Start receiving message")
	response := c.Receive(done)

	log.Println("Start getting keyboard inputs")
	keyInput := Input(done)

	log.Println("Start login conversation")
	loginConversation(c)

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
					msgTokens := strings.Split(message, " ")
					switch msgTokens[0] {
					case "quit":
						log.Println("Exit by quit signal from keyboard input")
						done <- struct{}{}
						return
					case "ROOM":
						if c.mode == ChatMode {
							roomID, _ := strconv.Atoi(msgTokens[1])
							chatLogs.CurrentRoomID = roomID
						}
						continue
					}
					// Server require new line character
					c.Send(message)
				case termbox.KeyEsc:
					log.Println("Exit by KeyEsc signal")
					done <- struct{}{}
					return
				case termbox.KeySpace:
					r, _ := utf8.DecodeLastRune([]byte(" "))
					eb.InsertRune(r)
				case termbox.KeyF9:
					switch c.mode {
					case DirectMode:
						c.mode = RoomMode
						ts.SetTextArea(roomList)
					case RoomMode:
						c.mode = ChatMode
						ts.SetTextArea(chatLogs)
					case ChatMode:
						c.mode = DirectMode
						ts.SetTextArea(connMsg)
					}

				case termbox.KeyF3:
					ts.SetTextArea(connMsg)
				default:
					eb.InsertRune(k.Ch)
				}
			case termbox.EventError:
				done <- struct{}{}
				return
			}
		case responseMsg := <-response:
			texts := strings.Split(responseMsg, "\r\n")
			for _, s := range texts {
				connMsg.AppendText("Server response: " + s)
				tokens := strings.Split(s, " ")
				switch tokens[0] {
				case "quit":
					log.Println("Exit by quit signal from server message")
					done <- struct{}{}
					return
				case "MESSAGE":
					var chat string
					for _, str := range tokens[2:] {
						chat += str + " "
					}
					id, _ := strconv.Atoi(tokens[1])
					chatLogs.AppendText(id, chat)
				case "OK":
					switch tokens[1] {
					case "PING":
						c.conn.SetReadDeadline(time.Now().Add(400 * time.Second))
						c.conn.SetWriteDeadline(time.Now().Add(400 * time.Second))
					case "OPEN_ROOM":
						id, _ := strconv.Atoi(tokens[2])
						chatLogs.CurrentRoomID = id
						roomList.EnterRoom(id)
					case "CLOSE_ROOM":
						id, _ := strconv.Atoi(tokens[2])
						roomList.QuitRoom(id)
					}
				case "SVR_PING":
					c.Send("OK SVR_PING")
				case "ROOM_ADDED":
					id, _ := strconv.Atoi(tokens[1])
					ri := RoomInfo{id, tokens[3], tokens[4], false}
					roomList.AppendRoom(ri)
				case "ROOM_REMOVED":
					id, _ := strconv.Atoi(tokens[1])
					roomList.RemoveRoom(id)
				}
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

type userInfo struct {
	user         string
	id           int
	introduction string
	level        string
	clientInfo   string
}

func loginConversation(c *ConnClient) {
	user := &userInfo{
		"test",
		1,
		"I'm man",
		"strong",
		"java.vender, java.version, os.name, os.version"}
	c.Send("LOGIN " + user.user)
	c.Send("SET_INTRO " + user.introduction)
	c.Send("SET_LEVEL " + user.level)
	c.Send("CLIENT_INFO " + user.clientInfo)
	c.Send("SET_ID " + fmt.Sprintf("%d", user.id))

}
