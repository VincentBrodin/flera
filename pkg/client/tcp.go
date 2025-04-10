package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

func (c *Client) connectTcp(port string) error {
	addr, err := net.ResolveTCPAddr("tcp", port)
	if err != nil {
		return err
	}

	attempts := 0
	for {
		c.tcpServer, err = net.DialTCP("tcp", nil, addr)
		if err != nil {
			fmt.Println("Failed to connect to tcp, will try again in 5 secs")
			if 5 == attempts {
				return err
			}
			attempts++
			time.Sleep(5 * time.Second)
			continue
		}
		return nil
	}
}

func (c *Client) handleTcpConn() {
	c.tcpConnected = true
	defer func() {
		c.tcpConnected = false
		fmt.Println("tcp lost")
	}()

	// listen for messages
	info := make([]byte, 8)
	for {
		// fmt.Println("Waiting on tcp")

		// First read the info part of the packet
		if _, err := c.tcpServer.Read(info); err != nil {
			fmt.Println(err)
			return
		}
		// fmt.Println("Got something on tcp")

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
		if _, err := c.tcpServer.Read(data); err != nil {
			fmt.Println(err)
			continue
		}

		if handler, ok := c.handlers[handlerId]; ok {
			handler(c, data)
		} else {
			fmt.Printf("No handler with id %d from tcp\n", handlerId)
			continue
		}
	}
}

func (c *Client) SendSafe(handlerId uint32, data []byte) error {
	handlerBuf := new(bytes.Buffer)
	if err := binary.Write(handlerBuf, binary.BigEndian, uint32(handlerId)); err != nil {
		return err
	}

	sizeBuf := new(bytes.Buffer)
	if err := binary.Write(sizeBuf, binary.BigEndian, uint32(len(data))); err != nil {
		return err
	}

	packet := append(append(handlerBuf.Bytes(), sizeBuf.Bytes()...), data...)
	if _, err := c.tcpServer.Write(packet); err != nil {
		return err
	}

	// fmt.Printf("Sent TCP to server: connId=%d, handlerId=%d, size=%d\n", c.Id, handlerId, len(data))

	return nil
}
