package main

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"
)

type ConnMock struct{}

func (c ConnMock) Read(b []byte) (n int, err error) { return 0, nil }
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
