package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"
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

func (c *Client) connectUdp(port string) error {
	addr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		return err
	}

	attempts := 0
	for {
		c.udpServer, err = net.DialUDP("udp", nil, addr)
		if err != nil {
			fmt.Println("Failed to connect to udp, will try again in 5 secs")
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

func (c *Client) handleUdpConn() {
	c.udpConnected = true
	defer func() {
		c.udpConnected = false
		fmt.Println("udp lost")
	}()

	if err := c.SendFast(^uint32(0), []byte{}); err != nil{
		fmt.Println(err)
	}

	// listen for messages
	buf := make([]byte, c.UdpPacketSize+4)
	for {
		// fmt.Println("Waiting on udp")

		n, err := c.udpServer.Read(buf)
		if err != nil {
			fmt.Println(err)
			return
		}
		// fmt.Println("Got something from udp")

		var handlerId uint32
		if err := binary.Read(bytes.NewReader(buf[:4]), binary.BigEndian, &handlerId); err != nil {
			fmt.Println(err)
			continue
		}

		// fmt.Println(handlerId)

		handler, ok := c.handlers[handlerId]
		if ok {
			handler(c, buf[4:n])
		} else {
			fmt.Printf("No handler with id %d from udp\n", handlerId)
			continue
		}
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

func (c *Client) SendFast(handlerId uint32, data []byte) error {
	idBuf := new(bytes.Buffer)
	if err := binary.Write(idBuf, binary.BigEndian, c.Id); err != nil {
		return err
	}
	handlerBuf := new(bytes.Buffer)
	if err := binary.Write(handlerBuf, binary.BigEndian, handlerId); err != nil {
		return err
	}

	packet := append(append(idBuf.Bytes(), handlerBuf.Bytes()...), data...)
	if _, err := c.udpServer.Write(packet); err != nil {
		return err
	}

	// fmt.Printf("Sent UDP to server: connId=%d, handlerId=%d, size=%d\n", c.Id, handlerId, len(data))

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
