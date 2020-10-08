package p0partA

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/cmu440/p0partA/kvstore"
	"net"
	"strconv"
)

const MaxBufferSize = 500

type keyValueServer struct {
	activeNumber       int            //number of active connection clients
	droppedNumber      int            //number of dropped connection clients
	reportCounterChan  chan bool      //report a count for a newly added(true) or closed(false) connection
	requestCounterChan chan bool      //request a counter result of activeNumber(true) of  droppedNumber(false)
	replyCounterChan   chan int       //reply counter result
	operationChan      chan operation //channel for sending operation requests
	store              kvstore.KVStore
	listener           net.Listener
}

// the struct to represent a operation request from a client
type operation struct {
	action          string
	key             string
	value           []byte
	responseChannel chan []byte // a dedicated channel for sending responses to a specific connection
}

// New creates and returns (but does not start) a new KeyValueServer.
func New(store kvstore.KVStore) KeyValueServer {
	return &keyValueServer{
		0, 0, make(chan bool), make(chan bool), make(chan int), make(chan operation), store, nil,
	}
}

func (kvs *keyValueServer) Start(port int) error {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	kvs.listener = listener
	go kvs.acceptConnections()
	go kvs.storeManager()
	return nil
}

func (kvs *keyValueServer) acceptConnections() {
	go kvs.counterRoutine() //start the goroutine for counting clients
	for {
		connection, err := kvs.listener.Accept()
		if err != nil {
			return
		}
		kvs.reportCounterChan <- true
		go kvs.handleConnection(connection)
	}
}

func (kvs *keyValueServer) counterRoutine() {
	for {
		select {
		case report := <-kvs.reportCounterChan:
			if report {
				kvs.activeNumber++
			} else {
				kvs.droppedNumber++
				kvs.activeNumber--
			}
		case req := <-kvs.requestCounterChan:
			if req {
				kvs.replyCounterChan <- kvs.activeNumber
			} else {
				kvs.replyCounterChan <- kvs.droppedNumber
			}
		}
	}
}

func (kvs *keyValueServer) handleConnection(connection net.Conn) {
	defer connection.Close()
	//the channel for receiving response for this connection
	responseChan := make(chan []byte, MaxBufferSize)
	reader := bufio.NewReader(connection)
	//start the writer routine for this connection
	go writeToConnection(connection, responseChan)

	for {
		payload, err := reader.ReadBytes('\n')
		if err != nil { // error when io.EOF is spotted
			kvs.reportCounterChan <- false //report client dropped
			return
		}
		elements := bytes.Split(bytes.TrimSpace(payload), []byte(":"))
		switch string(elements[0]) {
		case "Put":
			kvs.operationChan <- operation{"Put", string(elements[1]), elements[2], nil}
		case "Delete":
			kvs.operationChan <- operation{"Delete", string(elements[1]), nil, nil}
		case "Get":
			kvs.operationChan <- operation{"Get", string(elements[1]), nil, responseChan}
		}
	}
}

func writeToConnection(connection net.Conn, channel chan []byte) {
	for body := range channel {
		connection.Write(body)
	}
}

//the function to deal with all storage operations
func (kvs *keyValueServer) storeManager() {
	for {
		select {
		case input := <-kvs.operationChan:
			switch input.action {
			case "Put":
				kvs.store.Put(input.key, input.value)
			case "Delete":
				kvs.store.Clear(input.key)
			case "Get":
				responses := kvs.store.Get(input.key)
				for i := 0; i < len(responses); i++ {
					if len(input.responseChannel) < MaxBufferSize {
						input.responseChannel <- []byte(fmt.Sprintf("%s:%s\n", input.key, responses[i]))
					}
				}
			}
		}
	}
}

func (kvs *keyValueServer) Close() {
	kvs.listener.Close()
}

func (kvs *keyValueServer) CountActive() int {
	kvs.requestCounterChan <- true
	return <-kvs.replyCounterChan
}

func (kvs *keyValueServer) CountDropped() int {
	kvs.requestCounterChan <- false
	return <-kvs.replyCounterChan
}
