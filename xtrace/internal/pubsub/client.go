package pubsub

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"sync/atomic"
)

// A Client represents a connection to a pubsub server.
// The zero value is not a valid Client.
type Client struct {
	messages chan message
	quit     chan struct{}
	closed   uint32
}

// NewClient creates a new connection to server.
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
				if m.wg != nil {
					m.wg.Done()
				}
				break
			}
			fmt.Fprintf(os.Stderr, "pubsub client error: %v\n", err)

			for {
				conn, err = net.Dial("tcp", server)
				if err == nil {
					break
				}
				fmt.Fprintf(os.Stderr, "pubsub client error: %v\n", err)
			}
		}
	}
}

// Close closes the connection with the server.
func (c *Client) Close() {
	select {
	case c.quit <- struct{}{}:
	default:
	}
	atomic.StoreUint32(&c.closed, 1)
}

// Publish publishes msg on the given topic. Publish may block,
// but it is not guaranteed that when Publish returns, the message
// has been received by the server.
func (c *Client) Publish(topic, msg []byte) {
	if atomic.LoadUint32(&c.closed) == 1 {
		panic("publish on closed client")
	}

	c.messages <- message{topic: topic, message: msg}
}

// PublishBlock is like Publish, except that it blocks until
// the message has been written to the server. Note that this
// does not guarantee receipt by the server.
func (c *Client) PublishBlock(topic, msg []byte) {
	if atomic.LoadUint32(&c.closed) == 1 {
		panic("publish on closed client")
	}

	var wg sync.WaitGroup
	wg.Add(1)
	c.messages <- message{topic: topic, message: msg, wg: &wg}
	wg.Wait()
}

// PublishString is equivalent to Publish([]byte(topic), []byte(msg)).
func (c *Client) PublishString(topic, msg string) {
	c.Publish([]byte(topic), []byte(msg))
}

// PublishStringBlock is equivalent to
// PublishBlock([]byte(topic), []byte(msg)).
func (c *Client) PublishStringBlock(topic, msg string) {
	c.PublishBlock([]byte(topic), []byte(msg))
}

type message struct {
	topic   []byte
	message []byte
	wg      *sync.WaitGroup // used by PublishBlock to block until sent
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
