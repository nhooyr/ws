package ws

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"golang.org/x/net/http/httpguts"
)

type Conn struct {
	Client     bool
	ReadWriter io.ReadWriter
}

func (c Conn) WriteFrame(op Opcode, p []byte) error {
	panic("TODO")
}

func (c Conn) StreamFrame(op Opcode) io.WriteCloser {
	panic("TODO")
}

func (c Conn) ReadFrame() (typ Opcode, payload io.Reader, err error) {
	panic("TODO")
}

func Upgrade(w http.ResponseWriter, r *http.Request) (Conn, net.Conn, error) {
	if !httpguts.HeaderValuesContainsToken(r.Header["Connection"], "Upgrade") {
		return Conn{}, nil, errors.New(`Connection header must contain "Upgrade"`)
	}
	if !httpguts.HeaderValuesContainsToken(r.Header["Upgrade"], "websocket") {
		return Conn{}, nil, errors.New(`Upgrade header must contain "websocket"`)
	}
	if r.Method != http.MethodGet {
		return Conn{}, nil, fmt.Errorf("method must be %v", http.MethodGet)
	}
	if !r.ProtoAtLeast(1, 1) {
		return Conn{}, nil, errors.New("protocol must be at least HTTP/1.1")
	}
	if !httpguts.HeaderValuesContainsToken(r.Header["Sec-WebSocket-Version"], "13") {
		return Conn{}, nil, errors.New(`Sec-WebSocket-Version must contain "13"`)
	}

	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		return Conn{}, nil, errors.New("missing Sec-WebSocket-Key header")
	}

	w.Header().Set("Sec-WebSocket-Accept", secWebSocketAccept(key))
}


var acceptGUID = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")

func secWebSocketAccept(key string) string {
	h := sha1.New()
	io.WriteString(h, key)
	h.Write(acceptGUID)
	hash := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(hash)
}

// We use this key for all client requests.
// See https://stackoverflow.com/a/37074398/4283659
// Also see https://trac.ietf.org/trac/hybi/wiki/FAQ#WhatSec-WebSocket-KeyandSec-WebSocket-Acceptarefor
// The Sec-WebSocket-Key header is useless.
var secWebSocketKey = base64.StdEncoding.EncodeToString(make([]byte, 16))

func NewHandshake(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", secWebSocketKey)
	return req, nil
}

// Will be removed/modified in Go 1.12 as we can use http.Client to do the handshake.
func Handshake(rw *bufio.ReadWriter, req *http.Request) (Conn, *http.Response, error) {
	// Need to pass in writer explicitly as req.Write handles it specially.
	err := req.Write(rw.Writer)
	if err != nil {
		return Conn{}, nil, err
	}

	err = rw.Flush()
	if err != nil {
		return Conn{}, nil, err
	}

	resp, err := http.ReadResponse(rw.Reader, req)
	if err != nil {
		return Conn{}, nil, err
	}

	err = ServerResponse(resp)
	if err != nil {
		return Conn{}, resp, err
	}

	conn := Conn{
		ReadWriter: rw,
		Client:     true,
	}

	return conn, resp, nil
}

// TODO link rfc everywhere
func ServerResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusSwitchingProtocols {
		return Conn{}, resp, errors.New("websocket handshake failed; see response")
	}
}
