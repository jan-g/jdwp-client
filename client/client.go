package client

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Client interface {
	io.Closer
	Events() <-chan *Event
	Send(CommandSet, Command, []byte) (Id, <-chan *Reply, error)
	Dispose(Id)
	Call(CommandSet, Command, []byte) (*Reply, error)
}

type CommandSet uint8
type Command uint8
type Id uint32

type client struct {
	conn      net.Conn
	close     chan struct{}
	wg        sync.WaitGroup
	e         chan *Event
	id        Id
	responses sync.Map
}

var _ Client = &client{}

func Dial(network string, address string) (Client, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return New(conn)
}

func DialTimeout(network string, address string, timeout time.Duration) (Client, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	return New(conn)
}

func New(conn net.Conn) (Client, error) {
	c := &client{
		conn:  conn,
		close: make(chan struct{}),
		e:     make(chan *Event),
	}
	err := c.Handshake()
	if err != nil {
		c.Close()
		return nil, err
	}
	c.wg.Add(1)
	go c.read()
	return c, nil
}

func (c *client) Close() error {
	close(c.close)
	err := c.conn.Close()
	c.wg.Wait()
	return err
}

const Handshake = "JDWP-Handshake"

func (c *client) Handshake() error {
	handshake := []byte(Handshake)
	n, err := c.conn.Write(handshake)
	if err != nil {
		return err
	} else if n != len(handshake) {
		return fmt.Errorf("insufficient bytes of handshake written: %d/%d", n, len(handshake))
	}
	// Read the response
	resp := make([]byte, len(handshake))
	n, err = io.ReadFull(c.conn, resp)
	if err != nil {
		return err
	} else if n != len(handshake) {
		return fmt.Errorf("insufficient bytes of handshake written: %d/%d", n, len(handshake))
	}
	if string(resp[:n]) != Handshake {
		return fmt.Errorf("insufficient bytes of handshake written: %d/%d", n, len(handshake))
	}
	return nil
}

type command struct {
	Header
	commandSet CommandSet
	command    Command
	data       []byte
}

type Event struct {
	Header
	Set     CommandSet
	Command Command
	Data    []byte
}

type Header struct {
	Length uint32
	Id     Id
	Flags  uint8
}

type Reply struct {
	Header
	ErrCode uint16
	Data    []byte
}

const HeaderLength = 11

const (
	EventCommandSet   = CommandSet(64)
	CompositeCommands = Command(100)
)

func (c *client) read() {
	defer c.wg.Done()
	defer close(c.e)
	for {
		select {
		case <-c.close:
			logrus.Debug("jdwp client reader exits")
			return
		default:
			logrus.Debug("jdwp client waiting for read")
			var header Header
			err := readBytes(nil, c.conn, &header.Length)
			err = readBytes(err, c.conn, &header.Id)
			err = readBytes(err, c.conn, &header.Flags)
			var pair uint16
			err = readBytes(err, c.conn, &pair)
			if err != nil {
				logrus.WithError(err).Error("jdwp client encountered error during read")
				return
			}
			data := make([]byte, header.Length-HeaderLength)
			_, err = io.ReadFull(c.conn, data)
			if ch, ok := c.responses.Load(header.Id); ok {
				reply := Reply{Header: header, ErrCode: pair, Data: data}
				logrus.WithField("reply", reply).Debug("jdwp read")
				ch.(chan *Reply) <- &reply
			} else {
				event := Event{Header: header, Set: CommandSet(pair >> 8), Command: Command(pair & 0xff), Data: data}
				logrus.WithField("event", event).Debug("jdwp read")
				c.e <- &event
			}
		}
	}
}

func (c *client) Events() <-chan *Event {
	return c.e
}

func readBytes(err error, conn net.Conn, data interface{}) error {
	if err != nil {
		return err
	}
	return binary.Read(conn, binary.BigEndian, data)
}

func writeBytes(err error, out io.Writer, data interface{}) error {
	if err != nil {
		return err
	}
	return binary.Write(out, binary.BigEndian, data)
}

func writeString(err error, out io.Writer, s string) error {
	if err != nil {
		return err
	}
	data := []byte(s)
	l32 := int32(len(data))
	err = writeBytes(nil, out, l32)
	if err != nil {
		return err
	}
	_, err = out.Write(data)
	return err
}

func (c *client) Send(set CommandSet, cmd Command, data []byte) (Id, <-chan *Reply, error) {
	c.id++
	replyOn := make(chan *Reply, 1)
	c.responses.Store(c.id, replyOn)
	logrus.Debug("sending", set, cmd, c.id, data)
	err := writeBytes(nil, c.conn, uint32(len(data)+HeaderLength))
	err = writeBytes(err, c.conn, c.id)
	err = writeBytes(err, c.conn, uint8(0))
	err = writeBytes(err, c.conn, set)
	err = writeBytes(err, c.conn, cmd)
	for err == nil && len(data) > 0 {
		var n int
		n, err = c.conn.Write(data)
		data = data[n:]
	}
	return c.id, replyOn, err
}

func (c *client) Dispose(id Id) {
	c.responses.Delete(id)
}

func (c *client) Call(set CommandSet, cmd Command, data []byte) (*Reply, error) {
	id, ch, err := c.Send(set, cmd, data)
	defer c.Dispose(id)
	if err != nil {
		return nil, err
	}
	return <-ch, nil
}
