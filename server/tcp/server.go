package tcp

import (
	"GameSocketServer/models"
	"fmt"
	"net"
	log "github.com/sirupsen/logrus"
)

type ServerConfig struct {
	Port int
}

type Server struct {
	config   ServerConfig
	listener net.Listener
	connections map[string]*connection
	msgChan  chan<- *models.Message
}

func NewTcpServer(config ServerConfig) (*Server, error) {
	return &Server{
		config:   config,
		listener: nil,
		msgChan:  nil,
		connections: make(map[string]*connection),
	}, nil
}

func (s *Server) Start(msgChan chan<- *models.Message) error {

	s.msgChan = msgChan

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))

	if err != nil {
		return err
	}

	s.listener = listener

	log.Infof("started listening on port %d", s.config.Port)

	defer func() {
		_ = listener.Close()
	}()

	s.acceptConnections()

	return nil
}

func (s *Server) Send(msg string, connectionId string) {
	c := s.connections[connectionId]

	if c == nil {
		return
	}

	err := c.send(msg)

	if err != nil {
		log.Errorf("error while sending data, err: %s", err)
	}
}

func (s *Server) acceptConnections() {
	for {

		c, err := s.listener.Accept()

		if err != nil {
			log.Errorf("failed to accept connection, error: %s", err)
			continue
		}

		log.Infof("accepted connection from %s", c.RemoteAddr())

		if err != nil {
			log.Errorf("error while accepting conn, err: %s", err)
			continue
		}

		go s.handleConnection(c)
	}
}

func (s *Server) handleConnection(conn net.Conn) {

	c := newConnection(conn)

	s.connections[c.id] = c

	for {
		data, err := c.read()

		if err != nil {
			c.conn.Close()
			log.Errorf("conn: %s, failed to read the message size, err: %s", c.id, err)
			log.Errorf("conn: %s, closing the connection", c.id, err)
			break
		}

		msg := &models.Message{
			ConnId: c.id,
		}

		err = msg.Parse(data)

		if err != nil {
			log.Errorf("conn: %s, message was incorrect", c.id, err)
		}

		log.Infof("sending message")
		s.msgChan <- msg
		log.Infof("sending message done")
	}

}

