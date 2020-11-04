package models

import "bytes"

type Message struct {
	ConnId string
	Cmd    string
	Args   []string
}

const (
	Broadcast  = "BROADCAST"
	Notify     = "NOTIFY"
	CreateRoom = "CREATE_ROOM"
	DeleteRoom = "DELETE_ROOM"
	JoinRoom   = "JOIN_ROOM"
	LeaveRoom  = "LEAVE_ROOM"
	Ping       = "PING"
)

var separator = []byte(" ")

func (m *Message) Parse(data []byte) error {
	parts := bytes.Split(data, separator)

	m.Cmd = string(parts[0])
	m.Args = make([]string, 0)

	if len(parts) > 1 {
		for _, p := range parts[1:] {
			m.Args = append(m.Args, string(p))
		}
	}

	return nil
}

func (m *Message) EnsureEnoughArguments(count int) bool {
	return len(m.Args) >= count
}
