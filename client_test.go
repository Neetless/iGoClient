package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"
)

type ConnMock struct{}

func (c ConnMock) Read(b []byte) (n int, err error) {
	time.Sleep(1000 * time.Millisecond)
	for i, c := range []byte("I received") {
		b[i] = c
	}
	return len(b), nil
}
func (c ConnMock) Write(b []byte) (n int, err error) {
	fmt.Fprintf(os.Stdout, string(b[:len(b)]))
	return
}
func (c ConnMock) Close() error                       { return nil }
func (c ConnMock) LocalAddr() net.Addr                { return nil }
func (c ConnMock) RemoteAddr() net.Addr               { return nil }
func (c ConnMock) SetDeadline(t time.Time) error      { return nil }
func (c ConnMock) SetReadDeadline(t time.Time) error  { return nil }
func (c ConnMock) SetWriteDeadline(t time.Time) error { return nil }

func TestPing(t *testing.T) {
	var cm ConnMock
	c := &ConnClient{conn: cm}
	done := make(chan struct{})
	go c.Ping(done)
	time.Sleep(10 * time.Second)
	done <- struct{}{}
	return
}

func BenchmarkPing(b *testing.B) {
	var cm ConnMock
	c := &ConnClient{conn: cm}
	done := make(chan struct{})
	for i := 0; i < b.N; i++ {
		go c.Ping(done)
	}
	time.Sleep(20 * time.Second)
	done <- struct{}{}
	return
}

func testInput() <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for i := 0; i < 4; i++ {
			out <- fmt.Sprintf("%d\n", i)
			time.Sleep(1000 * time.Millisecond)
		}
	}()
	return out
}

func TestDone(t *testing.T) {
	var cm ConnMock
	c := &ConnClient{conn: cm}
	done := make(chan struct{})
	defer close(done)
	receive := c.Receive(done)
	keyInput := testInput()
	t.Logf("Start process\n")
	for {
		select {
		case in := <-keyInput:
			//t.Logf("keyInput %s", in)
			log.Printf("keyInput %s\n", in)
			if in == "3\n" {
				<-receive
				done <- struct{}{}
				log.Println("send done")
				return
			}
		case rcv := <-receive:
			log.Printf("Receive \n")
			t.Logf("receive %s\n", rcv)
		}
	}
}
