package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
)

type Server struct {
	connMap sync.Map
	callMap map[uint32]Handler
	runId   uint32
	ln      *net.TCPListener
}

type Handler func(senderId uint32, data []byte) error

func (s *Server) Register(id uint32, handler Handler) {
	s.callMap[id] = handler
}

func (s *Server) Start(port string) error {
	addr, err := net.ResolveTCPAddr("tcp", port)
	if err != nil {
		return err
	}
	s.ln, err = net.ListenTCP("tcp", addr)

	for {
		conn, err := s.ln.AcceptTCP()
		if err != nil {
			fmt.Println("Error accepting client")
			continue
		}

		s.connMap.Store(s.runId, conn)
		go s.handleConn(s.runId, conn)
		s.runId++
	}
}

func (s *Server) handleConn(id uint32, conn *net.TCPConn) {
	// send id
	idBuf := new(bytes.Buffer)
	if err := binary.Write(idBuf, binary.BigEndian, id); err != nil {
		fmt.Println(err)
		return
	}

	if _, err := conn.Write(idBuf.Bytes()); err != nil {
		fmt.Println(err)
		return
	}

	// listen for messages
	info := make([]byte, 8)
	for {
		// First read the info part of the packet
		if _, err := conn.Read(info); err != nil {
			fmt.Println(err)
			return
		}

		var call uint32
		if err := binary.Read(bytes.NewReader(info[:4]), binary.BigEndian, &call); err != nil {
			fmt.Println(err)
			return
		}

		var size uint32
		if err := binary.Read(bytes.NewReader(info[4:]), binary.BigEndian, &size); err != nil {
			fmt.Println(err)
			return
		}

		// Read data
		data := make([]byte, size)
		if _, err := conn.Read(data); err != nil {
			fmt.Println(err)
			return
		}

		handler, ok := s.callMap[call]
		if ok {
			handler(id, data)
		} else {
			fmt.Printf("No handler with id %d\n", call)
			return
		}
	}
}

func New() *Server {
	s := new(Server)
	s.callMap = make(map[uint32]Handler)
	return s
}
