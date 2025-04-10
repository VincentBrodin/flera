package server

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

func (s *Server) SendToClientSafe(connId, handlerId uint32, data []byte) error {
	return s.sendTcp(connId, handlerId, data)
}

func (s *Server) SendToClientsSafe(connIds []uint32, handlerId uint32, data []byte) error {
	var errs []error
	for _, connId := range connIds {
		if err := s.sendTcp(connId, handlerId, data); err != nil {
			errs = append(errs, fmt.Errorf("conn %d: %w", connId, err))
		}
	}
	return errors.Join(errs...)
}

func (s *Server) BroadcastSafe(handlerId uint32, data []byte) error {
	var errs []error
	s.tcpConns.Range(func(key, value any) bool {
		connId, ok := key.(uint32)
		if !ok {
			errs = append(errs, fmt.Errorf("invalid connection id: %v", key))
			return true
		}

		if err := s.sendTcp(connId, handlerId, data); err != nil {
			errs = append(errs, fmt.Errorf("conn %d: %w", connId, err))
		}
		return true
	})
	return errors.Join(errs...)
}

func (s *Server) sendTcp(connId, callId uint32, data []byte) error {
	// fmt.Printf("Trying to send message to %d with tcp\n", connId)

	conn, err := s.getTcpConn(connId)
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

	// fmt.Printf("Message sent to %d with tcp\n", connId)
	return nil
}

func (s *Server) getTcpConn(connId uint32) (*net.TCPConn, error) {
	val, ok := s.tcpConns.Load(connId)
	if !ok {
		return nil, fmt.Errorf("Could not find %d in the connMap", connId)
	}

	conn, ok := val.(*net.TCPConn)
	if !ok {
		return nil, fmt.Errorf("Could convert conn map output to TCPConn")
	}

	return conn, nil
}

func (s *Server) handleTcpConn(connId uint32, conn *net.TCPConn) {
	// store client
	s.tcpConns.Store(connId, conn)

	defer func() {
		s.tcpConns.Delete(connId)
		// s.udpAddrs.Delete(connId)
		fmt.Printf("Conn %d lost via tcp\n", connId)
		if s.OnDisConn != nil {
			s.OnDisConn(s, connId)
		}
	}()

	fmt.Printf("Conn %d connected\n", connId)
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

		var handlerId uint32
		if err := binary.Read(bytes.NewReader(info[:4]), binary.BigEndian, &handlerId); err != nil {
			fmt.Println(err)
			continue
		}

		var size uint32
		if err := binary.Read(bytes.NewReader(info[4:]), binary.BigEndian, &size); err != nil {
			fmt.Println(err)
			continue
		}

		// Read data
		data := make([]byte, size)
		if _, err := conn.Read(data); err != nil {
			fmt.Println(err)
			return
		}

		if handler, ok := s.handlers[handlerId]; ok {
			go func() {
				if err := handler(s, connId, data); err != nil {
					fmt.Println(err)
				}
			}()
		} else {
			fmt.Printf("No handler with id %d\n", handlerId)
			continue
		}
	}
}
