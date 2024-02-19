package main

import (
	"errors"
	"net"
	"sync/atomic"
)

var ErrUsing = errors.New("Socks5 connection is using")
var ErrUnsupportNetwork = errors.New("Network unreachable")

type socks5Conn struct {
	state int32
	conn  net.Conn
}

func dialSocks5(addr *net.TCPAddr) (s *socks5Conn, err error) {
	s = &socks5Conn{}
	if s.conn, err = net.DialTCP("tcp", nil, addr); err != nil {
		return
	}
	return
}

func (s *socks5Conn) Dial(network string, addr string) (c net.Conn, err error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
	default:
		return nil, ErrUnsupportNetwork
	}
	if !atomic.CompareAndSwapInt32(&s.state, 0, 1) {
		return nil, ErrUsing
	}
	panic("TODO")
	return s.conn, nil
}

func (s *socks5Conn) ListenPacket(network string, addr string) (c net.PacketConn, err error) {
	switch network {
	case "udp", "udp4", "udp6":
	default:
		return nil, ErrUnsupportNetwork
	}
	if !atomic.CompareAndSwapInt32(&s.state, 0, 1) {
		return nil, ErrUsing
	}
	panic("TODO")
	return
}

type Socks5 struct {
	addr *net.TCPAddr
}

func NewSocks5(addr string) (s *Socks5, err error) {
	s = &Socks5{}
	if s.addr, err = net.ResolveTCPAddr("tcp", addr); err != nil {
		return
	}
	return
}

func (s *Socks5) Dial(network string, addr string) (c net.Conn, err error) {
	sc, err := dialSocks5(s.addr)
	if err != nil {
		return
	}
	return sc.Dial(network, addr)
}

func (s *Socks5) ListenPacket(network string, addr string) (c net.PacketConn, err error) {
	sc, err := dialSocks5(s.addr)
	if err != nil {
		return
	}
	return sc.ListenPacket(network, addr)
}
