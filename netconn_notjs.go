// +build !js

package websocket

import "net"

type LocalRemoteAddr interface {
	RemoteAddr() net.Addr
	LocalAddr() net.Addr
}

func (c *netConn) RemoteAddr() net.Addr {
	if ra, ok := c.c.rwc.(LocalRemoteAddr); ok {
		return ra.RemoteAddr()
	} else {
		return websocketAddr{}
	}
}

func (c *netConn) LocalAddr() net.Addr {
	if la, ok := c.c.rwc.(LocalRemoteAddr); ok {
		return la.RemoteAddr()
	} else {
		return websocketAddr{}
	}
}
