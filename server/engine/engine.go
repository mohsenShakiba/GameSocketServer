package engine

import (
	"GameSocketServer/models"
	"GameSocketServer/server/tcp"
	"fmt"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Port int
}

type Engine struct {
	server        *tcp.Server
	connIdRoomMap map[string]string
	rooms         []*Room
	msgChan       <-chan *models.Message
}

func (e *Engine) Start(cnf Config) error {
	server, err := tcp.NewTcpServer(tcp.ServerConfig{Port: cnf.Port})

	if err != nil {
		return err
	}

	e.server = server
	e.rooms = make([]*Room, 0, 100)
	e.connIdRoomMap = make(map[string]string)
	msgChan := make(chan *models.Message, 100)
	e.msgChan = msgChan

	go func() {
		err = server.Start(msgChan)

		if err != nil {
			log.Errorf("failed to start the server, error: %s", err)
		}
	}()


	e.process()

	log.Infof("started the engine")

	return nil
}

func (e *Engine) process() {
	for {
		msg := <- e.msgChan

		log.Infof("processing message %s", msg.Cmd)

		e.processMsg(msg)
	}
}

func (e *Engine) processMsg(msg *models.Message) {
	switch msg.Cmd {

	case models.CreateRoom:
		enoughArguments := msg.EnsureEnoughArguments(1)

		if !enoughArguments {
			log.Errorf("not enough arguments provider")
			e.server.Send("Not enough arguments provided!", msg.ConnId)
			return
		}
		e.create(msg.Args[0], msg.ConnId)
		break
	case models.DeleteRoom:
		enoughArguments := msg.EnsureEnoughArguments(1)

		if !enoughArguments {
			log.Errorf("not enough arguments provider")
			e.server.Send("Not enough arguments provided!", msg.ConnId)
			return
		}
		e.delete(msg.Args[0], msg.ConnId)
		break
	case models.JoinRoom:
		enoughArguments := msg.EnsureEnoughArguments(1)

		if !enoughArguments {
			log.Errorf("not enough arguments provider")
			e.server.Send("Not enough arguments provided!", msg.ConnId)
			return
		}
		e.join(msg.Args[0], msg.ConnId)
		break
	case models.LeaveRoom:
		room, ok := e.getRoomForConnId(msg.ConnId)

		if !ok {
			log.Errorf("this connection hasn't joined any room yet")
			e.server.Send("You haven't joined any room yet!", msg.ConnId)
			return
		}

		e.leave(room, msg.ConnId)
		break
	case models.Broadcast:
		enoughArguments := msg.EnsureEnoughArguments(1)

		if !enoughArguments {
			log.Errorf("not enough arguments provider")
			e.server.Send("Not enough arguments provided!", msg.ConnId)
			return
		}

		room, ok := e.getRoomForConnId(msg.ConnId)

		if !ok {
			log.Errorf("this connection hasn't joined any room yet")
			e.server.Send("You haven't joined any room yet!", msg.ConnId)
			return
		}

		e.broadcast(room, msg.Args[0])
		break
	case models.Notify:
		enoughArguments := msg.EnsureEnoughArguments(2)

		if !enoughArguments {
			log.Errorf("not enough arguments provider")
			e.server.Send("Not enough arguments provided!", msg.ConnId)
			return
		}

		room, ok := e.getRoomForConnId(msg.ConnId)

		if !ok {
			log.Errorf("this connection hasn't joined any room yet")
			e.server.Send("You haven't joined any room yet!", msg.ConnId)
			return
		}

		e.notify(room, msg.Args[0], msg.Args[1])
		break
	case models.Ping:
		e.server.Send("Pong!", msg.ConnId)
		break
	}
}

func (e *Engine) create(roomId string, connId string) {
	// make sure no room with the same id exists
	for _, r := range e.rooms {
		if r.name == roomId {
			e.server.Send("Room name already exists", connId)
			return
		}
	}

	room := &Room{
		connIds: make([]string, 0, 10),
		name:    roomId,
	}

	e.rooms = append(e.rooms, room)

	e.server.Send("Room created!", connId)
}

func (e *Engine) delete(roomId string, connId string) {
	var room *Room
	var roomIndex int

	for i, r := range e.rooms {
		if r.name == roomId {
			room = r
			roomIndex = i
		}
	}

	if room == nil {
		e.server.Send("No room found!", connId)
		return
	}

	msg := fmt.Sprintf("Room deleted, you are no longer in room %s", room.name)

	e.broadcast(room, msg)

	for _, cid := range room.connIds {
		delete(e.connIdRoomMap, cid)
	}

	e.rooms = append(e.rooms[:roomIndex], e.rooms[roomIndex+1:]...)

	e.server.Send("Room deleted", connId)

}

func (e *Engine) join(roomId string, connId string) {
	existingRoom, ok := e.getRoomForConnId(connId)

	if ok {
		msg := fmt.Sprintf("You are already in existingRoom %s", existingRoom.name)
		e.server.Send(msg, connId)
		return
	}

	room, ok := e.getRoomById(roomId)

	if !ok {
		msg := fmt.Sprintf("No room found with name %s", roomId)
		e.server.Send(msg, connId)
		return
	}

	room.connIds = append(room.connIds, connId)

	e.connIdRoomMap[connId] = room.name

	e.server.Send("Joined room", connId)

}


func (e *Engine) leave(room *Room, connId string) {
	indexOfConnIdInRoom := -1

	for i, id := range room.connIds {
		if id == connId {
			indexOfConnIdInRoom = i
		}
	}

	if indexOfConnIdInRoom != -1 {
		room.connIds = append(room.connIds[:indexOfConnIdInRoom], room.connIds[indexOfConnIdInRoom+1:]...)
	}

	delete(e.connIdRoomMap, connId)

	msg := fmt.Sprintf("you left room %s", room.name)
	e.server.Send(msg, connId)
}

func (e *Engine) broadcast(room *Room, msg string) {
	for _, connId := range room.connIds {
		e.server.Send(msg, connId)
	}
}

func (e *Engine) notify(room *Room, msg string, target string) {
	for _, connId := range room.connIds {
		if connId == target {
			e.server.Send(msg, connId)
		}
	}
}

func (e *Engine) getRoomForConnId(connId string) (*Room, bool) {
	roomId, ok := e.connIdRoomMap[connId]

	if !ok {
		return nil, false
	}

	room, ok := e.getRoomById(roomId)

	return room, ok
}

func (e *Engine) getRoomById(roomId string) (*Room, bool) {
	var room *Room

	for _, r := range e.rooms {
		if r.name == roomId {
			room = r
		}
	}

	if room == nil {
		return nil, false
	}

	return room, true
}

