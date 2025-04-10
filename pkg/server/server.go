package server

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	tcpConns   sync.Map
	udpAddrs   sync.Map
	handlers   map[uint32]Handler
	runId     uint32
	tcpLn        *net.TCPListener
	udpConn *net.UDPConn
	OnConn    Event
	OnDisConn Event
	UdpPacketSize uint32
}

type Handler func(s *Server, connId uint32, data []byte) error
type Event func(s *Server, connId uint32)

func (s *Server) Register(handlerId uint32, handler Handler) {
	s.handlers[handlerId] = handler
}

func (s *Server) Start(port string) error {
	// setup udp
	udpAddr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		return err
	}

	s.udpConn, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	go s.serveUDP()

	// setup tcp
	tcpAddr, err := net.ResolveTCPAddr("tcp", port)
	if err != nil {
		return err
	}

	s.tcpLn, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	for {
		conn, err := s.tcpLn.AcceptTCP()
		if err != nil {
			fmt.Println("Error accepting client")
			continue
		}

		go s.handleTcpConn(s.runId, conn)
		s.runId++
	}
}


func New() *Server {
	s := new(Server)
	s.handlers = make(map[uint32]Handler)
	s.UdpPacketSize = 1024
	return s
}
