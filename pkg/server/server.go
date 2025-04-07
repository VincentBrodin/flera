package server

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"
)

type Server struct {
	connMap   sync.Map
	callMap   map[uint32]Handler
	runId     uint32
	ln        *net.TCPListener
	OnConn    Event
	OnDisConn Event
}

type Handler func(s *Server, connId uint32, data []byte) error
type Event func(s *Server, connId uint32)

func (s *Server) Register(callId uint32, handler Handler) {
	s.callMap[callId] = handler
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

		go s.handleConn(s.runId, conn)
		s.runId++
	}
}

func (s *Server) SendToClient(connId, callId uint32, data []byte) error {
	return s.send(connId, callId, data)
}
func (s *Server) SendToClients(connIds []uint32, callId uint32, data []byte) error {
	var errs []error
	for _, connId := range connIds {
		if err := s.send(connId, callId, data); err != nil {
			errs = append(errs, fmt.Errorf("conn %d: %w", connId, err))
		}
	}
	return errors.Join(errs...)
}

func (s *Server) BroadCast(callId uint32, data []byte) error {
	var errs []error
	s.connMap.Range(func(key, value any) bool {
		connId, ok := key.(uint32)
		if !ok {
			errs = append(errs, fmt.Errorf("invalid connection id: %v", key))
			return true
		}

		if err := s.send(connId, callId, data); err != nil {
			errs = append(errs, fmt.Errorf("conn %d: %w", connId, err))
		}
		return true
	})
	return errors.Join(errs...)
}

func (s *Server) send(connId, callId uint32, data []byte) error {
	fmt.Printf("Trying to send message to %d\n", connId)

	conn, err := s.GetConn(connId)
	if err != nil {
		return err
	}

	callBuf := new(bytes.Buffer)
	if err := binary.Write(callBuf, binary.BigEndian, uint32(callId)); err != nil {
		return err
	}

	sizeBuf := new(bytes.Buffer)
	if err := binary.Write(sizeBuf, binary.BigEndian, uint32(len(data))); err != nil {
		return err
	}

	packet := append(append(callBuf.Bytes(), sizeBuf.Bytes()...), data...)
	if _, err := conn.Write(packet); err != nil {
		return err
	}

	fmt.Printf("Message sent to %d\n", connId)
	return nil
}

func (s *Server) GetConn(connId uint32) (*net.TCPConn, error) {
	val, ok := s.connMap.Load(connId)
	if !ok {
		return nil, fmt.Errorf("Could not find %d in the connMap", connId)
	}

	conn, ok := val.(*net.TCPConn)
	if !ok {
		return nil, fmt.Errorf("Could convert conn map output to TCPConn")
	}

	return conn, nil
}

func (s *Server) handleConn(connId uint32, conn *net.TCPConn) {
	// Store player
	s.connMap.Store(connId, conn)
	defer func() {
		s.connMap.Delete(connId)
		fmt.Printf("Conn %d lost\n", connId)
		if s.OnDisConn != nil {
			s.OnDisConn(s, connId)
		}
	}()
	// send id
	idBuf := new(bytes.Buffer)
	if err := binary.Write(idBuf, binary.BigEndian, connId); err != nil {
		fmt.Println(err)
		return
	}

	if _, err := conn.Write(idBuf.Bytes()); err != nil {
		fmt.Println(err)
		return
	}

	// send on conn event
	if s.OnConn != nil {
		s.OnConn(s, connId)
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
			handler(s, connId, data)
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
