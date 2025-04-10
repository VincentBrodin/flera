package server

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

func (s *Server) SendToClientFast(connId, handlerId uint32, data []byte) error {
	return s.sendUdp(connId, handlerId, data)
}

func (s *Server) SendToClientsFast(connIds []uint32, handlerId uint32, data []byte) error {
	var errs []error
	for _, connId := range connIds {
		if err := s.sendUdp(connId, handlerId, data); err != nil {
			errs = append(errs, fmt.Errorf("conn %d: %w", connId, err))
		}
	}
	return errors.Join(errs...)
}

func (s *Server) BroadcastFast(handlerId uint32, data []byte) error {
	var errs []error
	s.udpAddrs.Range(func(key, value any) bool {
		connId, ok := key.(uint32)
		if !ok {
			errs = append(errs, fmt.Errorf("invalid connection id: %v", key))
			return true
		}

		if err := s.sendUdp(connId, handlerId, data); err != nil {
			errs = append(errs, fmt.Errorf("conn %d: %w", connId, err))
		}
		return true
	})
	return errors.Join(errs...)
}

func (s *Server) sendUdp(connId, callId uint32, data []byte) error {
	fmt.Printf("Trying to send message to %d with udp\n", connId)

	addr, err := s.getUdpAddr(connId)
	if err != nil {
		return err
	}

	callBuf := new(bytes.Buffer)
	if err := binary.Write(callBuf, binary.BigEndian, uint32(callId)); err != nil {
		return err
	}

	packet := append(callBuf.Bytes(), data...)
	if _, err := s.udpConn.WriteToUDP(packet, addr); err != nil {
		return err
	}

	fmt.Printf("Message sent to %d with udp\n", connId)
	return nil
}

func (s *Server) getUdpAddr(connId uint32) (*net.UDPAddr, error) {
	val, ok := s.udpAddrs.Load(connId)
	if !ok {
		return nil, fmt.Errorf("Could not find %d in the connMap", connId)
	}

	conn, ok := val.(*net.UDPAddr)
	if !ok {
		return nil, fmt.Errorf("Could convert conn map output to TCPConn")
	}

	return conn, nil
}

func (s *Server) serveUDP() {
	defer s.udpConn.Close()
	buf := make([]byte, s.UdpPacketSize+8)
	for {
		fmt.Println("Udp serving")
		n, _, err := s.udpConn.ReadFromUDP(buf)
		fmt.Println("Got udp message")
		if err != nil {
			fmt.Println(err)
			continue
		}

		var connId uint32
		if err := binary.Read(bytes.NewReader(buf[0:4]), binary.BigEndian, &connId); err != nil {
			fmt.Println(err)
			continue
		}

		var handlerId uint32
		if err := binary.Read(bytes.NewReader(buf[4:8]), binary.BigEndian, &handlerId); err != nil {
			fmt.Println(err)
			continue
		}

		if handler, ok := s.handlers[handlerId]; ok {
			go handler(s, connId, buf[8:n])
		} else {
			fmt.Printf("No handler found for function %d\n", handlerId)
		}

	}
}
