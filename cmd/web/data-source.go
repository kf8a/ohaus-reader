package main

import (
	"encoding/json"
	ohaus "github.com/kf8a/ohaus"
	"log"
)

type dataSource struct {
	connections map[*connection]bool
	register    chan *connection
	unregister  chan *connection
}

func newDataSource() *dataSource {
	return &dataSource{
		connections: make(map[*connection]bool),
		register:    make(chan *connection),
		unregister:  make(chan *connection),
	}
}

// the one place where we talk to the insturment
func (q *dataSource) readData(cs chan string, test bool) {
	var data ohaus.Datum
	c := make(chan ohaus.Datum)
	scale := ohaus.Scale{PortName: "/dev/ttyUSB0"}
	if test {
		go scale.TestReader(c)
	} else {
		go scale.Reader(c)
	}

	for {
		data = <-c
		result, err := json.Marshal(data)
		if err != nil {
			log.Fatal(err)
		}
		cs <- string(result)
	}
}

func (q *dataSource) read(test bool) {

	cs := make(chan string)
	data := newDataSource()

	go data.readData(cs, test)

	for {
		select {
		case c := <-q.register:
			q.connections[c] = true
		case c := <-q.unregister:
			if q.connections[c] {
				q.connections[c] = false
				delete(q.connections, c)
				close(c.send)
			}
		default:
			data := <-cs
			for c := range q.connections {
				select {
				case c.send <- []byte(data):
				default:
					delete(q.connections, c)
					close(c.send)
				}
			}
		}
	}
}
