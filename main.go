package main

import (
	"GameSocketServer/server/engine"
	log "github.com/sirupsen/logrus"
)

func main() {

	eng := &engine.Engine{}
	cnf := engine.Config{Port: 9100}

	err := eng.Start(cnf)

	if err != nil {
		log.Errorf("failed to start the engine, error: %s", err)
	}
}
