package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

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

	if err := c.SendFast(^uint32(0), []byte{}); err != nil {
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
