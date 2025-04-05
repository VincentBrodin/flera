package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

type Client struct {
	Id      uint32
	callMap map[uint32]Handler
	server  *net.TCPConn
}

type Handler func(senderId uint32, data []byte) error

func (c *Client) Register(id uint32, handler Handler) {
	c.callMap[id] = handler
}

func (c *Client) Connect(port string) error {
	addr, err := net.ResolveTCPAddr("tcp", port)
	if err != nil {
		return err
	}
	attempts := 0
	for {
		c.server, err = net.DialTCP("tcp", nil, addr)
		if err != nil {
			fmt.Println("Failed to connect, will try again in 5 secs")
			time.Sleep(5000)
			if 5 == attempts {
				return err
			}
		}
		break
	}

	// Grab id
	idBuf := make([]byte, 4)
	if _, err = c.server.Read(idBuf); err != nil {
		return err
	}
	if err := binary.Read(bytes.NewReader(idBuf), binary.BigEndian, &c.Id); err != nil {
		return err
	}

	fmt.Println(c.Id)

	return nil
}

func (c *Client) Send(id uint32, data []byte) error {
	callBuf := new(bytes.Buffer)
	if err := binary.Write(callBuf, binary.BigEndian, uint32(id)); err != nil {
		return err
	}

	sizeBuf := new(bytes.Buffer)
	if err := binary.Write(sizeBuf, binary.BigEndian, uint32(len(data))); err != nil {
		return err
	}

	packet := append(append(callBuf.Bytes(), sizeBuf.Bytes()...), data...)
	if _, err := c.server.Write(packet); err != nil {
		return err
	}

	return nil
}

func New() *Client {
	c := new(Client)
	c.callMap = make(map[uint32]Handler)
	return c
}
