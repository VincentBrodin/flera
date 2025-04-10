package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

type Client struct {
	Id            uint32
	handlers      map[uint32]Handler
	tcpServer     *net.TCPConn
	udpServer     *net.UDPConn
	tcpConnected  bool
	udpConnected  bool
	UdpPacketSize uint32
}

type Handler func(c *Client, data []byte) error

func (c *Client) Register(id uint32, handler Handler) {
	c.handlers[id] = handler
}

func (c *Client) Connect(port string) error {

	if err := c.connectTcp(port); err != nil {
		return err
	}
	// Grab id
	idBuf := make([]byte, 4)
	if _, err := c.tcpServer.Read(idBuf); err != nil {
		return err
	}
	if err := binary.Read(bytes.NewReader(idBuf), binary.BigEndian, &c.Id); err != nil {
		return err
	}
	fmt.Println(c.Id)
	go c.handleTcpConn()

	if err := c.connectUdp(port); err != nil {
		return err
	}
	go c.handleUdpConn()

	return nil
}

func (c *Client) Connected() bool {
	return c.tcpConnected && c.udpConnected
}

func New() *Client {
	c := new(Client)
	c.handlers = make(map[uint32]Handler)
	c.tcpConnected = false
	c.udpConnected = false
	c.UdpPacketSize = 1024
	return c
}
