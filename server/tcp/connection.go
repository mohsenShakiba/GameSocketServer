package tcp

import (
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"io"
	"math/rand"
	"net"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type connection struct {
	conn net.Conn
	id string
}

func newConnection(conn net.Conn) *connection {

	id := RandStringBytes(3)

	return &connection{
		conn: conn,
		id:   id,
	}
}

func (c *connection) read() ([]byte, error) {
	msgSizeB := make([]byte, 4)

	_, err := io.ReadFull(c.conn, msgSizeB)

	if err != nil {
		return nil, err
	}

	// read the size of message
	msgSize := binary.BigEndian.Uint32(msgSizeB)

	// read exactly the size of message from connection
	msgB := make([]byte, msgSize)
	_, err = io.ReadFull(c.conn, msgB)

	if err != nil {
		return nil, err
	}

	log.Infof("read message %s", string(msgB))

	return msgB, nil
}

func (c *connection) send(msg string) error {

	bSize := make([]byte, 4)

	binary.BigEndian.PutUint32(bSize, uint32(len(msg)))

	_, err := c.conn.Write(bSize)

	if err != nil {
		return err
	}

	_, err = c.conn.Write([]byte(msg))

	if err != nil {
		return err
	}

	return nil
}

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}