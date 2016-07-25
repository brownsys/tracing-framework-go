package client

import (
	"encoding/binary"
	"io"
	"net"
	"sync/atomic"
)

type Client struct {
	messages chan message
	quit     chan struct{}
	closed   uint32
}

func NewClient(server string) (c *Client, err error) {
	conn, err := net.Dial("tcp", server)
	if err != nil {
		return nil, err
	}

	c = &Client{
		messages: make(chan message, 1024),
		quit:     make(chan struct{}, 1),
	}
	go c.daemon(server, conn)
	return c, nil
}

func (c *Client) daemon(server string, conn net.Conn) {
	for {
		var m message
		select {
		case <-c.quit:
			conn.Close()
			return
		case m = <-c.messages:
		}

		for {
			err := writeMessage(conn, m)
			if err == nil {
				break
			}
			// TODO: Log error

			for {
				conn, err = net.Dial("tcp", server)
				if err == nil {
					break
				}
				// TODO: Log error
			}
		}
	}
}

func (c *Client) Close() {
	select {
	case c.quit <- struct{}{}:
	default:
	}
	atomic.StoreUint32(&c.closed, 1)
}

func (c *Client) Publish(topic, msg string) {
	if atomic.LoadUint32(&c.closed) == 1 {
		panic("publish on closed client")
	}

	c.messages <- message{[]byte(topic), []byte(msg)}
}

type message struct {
	topic   []byte
	message []byte
}

func writeMessage(w io.Writer, m message) error {
	l := 8 + len(m.topic) + len(m.message)
	buf := make([]byte, l)
	binary.BigEndian.PutUint32(buf, uint32(len(m.topic)))
	copy(buf[4:], m.topic)
	binary.BigEndian.PutUint32(buf[4+len(m.topic):], uint32(len(m.message)))
	copy(buf[8+len(m.topic):], m.message)

	n, err := w.Write(buf)
	if err != nil && n < len(buf) {
		return err
	}
	return nil
}
