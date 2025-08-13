package main

import (
	"fmt"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

type websocketConnWrapper struct {
	websocketConn *websocket.Conn
}

func (wrapper *websocketConnWrapper) Close() error {
	return wrapper.websocketConn.Close()
}

func (wrapper *websocketConnWrapper) Read(b []byte) (n int, err error) {
	_, reader, err := wrapper.websocketConn.NextReader()
	if err != nil {
		return 0, err
	}
	if reader == nil {
		return 0, fmt.Errorf("failed to get reader")
	}
	return reader.Read(b)
}

func (wrapper *websocketConnWrapper) Write(b []byte) (n int, err error) {
	writer, err := wrapper.websocketConn.NextWriter(websocket.BinaryMessage)
	return writer.Write(b)
}

func (wrapper *websocketConnWrapper) LocalAddr() net.Addr {
	return wrapper.websocketConn.LocalAddr()
}

func (wrapper *websocketConnWrapper) RemoteAddr() net.Addr {
	return wrapper.websocketConn.RemoteAddr()
}

func (wrapper *websocketConnWrapper) SetDeadline(t time.Time) error {
	return wrapper.websocketConn.NetConn().SetDeadline(t)
}

func (wrapper *websocketConnWrapper) SetReadDeadline(t time.Time) error {
	return wrapper.websocketConn.SetReadDeadline(t)
}

func (wrapper *websocketConnWrapper) SetWriteDeadline(t time.Time) error {
	return wrapper.websocketConn.SetWriteDeadline(t)
}
