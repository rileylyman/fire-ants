package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
)

type Entity struct {
	Data  int
	Index int
}

type WorkerManager struct {
	elements      []Entity
	indices       chan int
	readyWorkers  map[*Worker]bool
	register      chan *Worker
	unregister    chan *Worker
	workerMutex   sync.Mutex
	elementsMutex sync.Mutex
}

func (manager *WorkerManager) getElement(index int) (entity Entity) {
	manager.elementsMutex.Lock()
	entity = manager.elements[index]
	manager.elementsMutex.Unlock()
	return
}

func (manager *WorkerManager) getWorker(worker *Worker) (workerStatus bool, ok bool) {
	manager.workerMutex.Lock()
	workerStatus, ok = manager.readyWorkers[worker]
	manager.workerMutex.Unlock()
	return
}

func (manager *WorkerManager) setWorker(worker *Worker, ready bool) {
	manager.workerMutex.Lock()
	manager.readyWorkers[worker] = ready
	manager.workerMutex.Unlock()
}

func (manager *WorkerManager) deleteWorker(worker *Worker) {
	manager.workerMutex.Lock()
	delete(manager.readyWorkers, worker)
	manager.workerMutex.Unlock()
}

func (manager *WorkerManager) has(worker *Worker) (ok bool) {
	manager.workerMutex.Lock()
	_, ok = manager.readyWorkers[worker]
	manager.workerMutex.Unlock()
	return
}

func (manager *WorkerManager) receive(worker *Worker) {
	for manager.has(worker) {
		data := make([]byte, 4096)
		length, err := worker.socket.Read(data)
		if err != nil {
			fmt.Println(err)
		}
		if length > 0 {
			data = data[:length]
			e := unmarshal(data)
			fmt.Println(e.Data)
			manager.elements[e.Index] = *e
		}
		manager.setWorker(worker, true)

	}
}

func (manager *WorkerManager) sendElement(currentIndex int, worker *Worker) {
	e := manager.getElement(currentIndex)
	bytes, err := json.Marshal(e)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(bytes))
	}
	worker.socket.Write(bytes)
}

func (manager *WorkerManager) sendData(worker *Worker) {
	for manager.has(worker) {
		if ready, ok := manager.getWorker(worker); ready && ok {
			select{
			case index := <-manager.indices:
				manager.sendElement(index, worker)
				fmt.Println("Sending: " + strconv.Itoa(index))
				manager.setWorker(worker, false)
			default:
				break
			}
		}
	}
	manager.unregister <- worker
}

func (manager *WorkerManager) manage() {
	for {
		select {
		case worker := <-manager.register:
			manager.setWorker(worker, true)
			go manager.receive(worker)
			go manager.sendData(worker)
		case worker := <-manager.unregister:
			if _, ok := manager.readyWorkers[worker]; ok {
				manager.deleteWorker(worker)
			}
		}
	}

}

func (manager *WorkerManager) processConnections(connection net.Listener) {
	for {
		conn, _ := connection.Accept()
		worker := &Worker{socket: conn, data: make(chan []byte)}
		manager.register <- worker
	}
}

func startWorkerManager(els []Entity) {
	idxs := make(chan int, len(els))
	for index := range els {
		idxs <- index
	}

	manager := &WorkerManager{
		elements:     els,
		indices :     idxs,
		readyWorkers: make(map[*Worker]bool),
		register:     make(chan *Worker),
		unregister:   make(chan *Worker),
	}
	connection, err := net.Listen("tcp", ":12345")
	if err != nil {
		fmt.Println(err)
	}
	go manager.processConnections(connection)
	manager.manage()
}

// GREAT MASTER AND WORKER DVIDE

type Worker struct {
	socket net.Conn
	data   chan []byte
}

func unmarshal(rawData []byte) *Entity {
	var entityFields map[string]interface{}
	if marshalError := json.Unmarshal(rawData, &entityFields); marshalError != nil {
		fmt.Println(marshalError)
	}
	entity := &Entity{}
	entity.Data = int(entityFields["Data"].(float64))
	entity.Index = int(entityFields["Index"].(float64))
	return entity
}

func (worker *Worker) process(rawData []byte) {
	entity := unmarshal(rawData)
	entityMap(entity)
	data, anotherMarshalError := json.Marshal(entity)
	if anotherMarshalError != nil {
		fmt.Println(anotherMarshalError)
	}
	worker.data <- data
	fmt.Println("Adding data: " + strconv.Itoa(entity.Data))
}

func (worker *Worker) stop() {
	worker.socket.Close()
	close(worker.data)
}

func (worker *Worker) send() {
	for {
		select {
		case data := <-worker.data:
			_, err := worker.socket.Write(data)
			if err != nil {
				fmt.Println(err)
				worker.stop()
			}
		}
	}
}

func startWorker() {

	fmt.Println("Starting Worker")

	connection, err := net.Dial("tcp", "localhost:12345")
	if err != nil {
		fmt.Println(err)
	}
	worker := &Worker{socket: connection, data: make(chan []byte)}

	go worker.send()

	for {
		rawData := make([]byte, 4096)
		length, err := connection.Read(rawData)
		fmt.Println("Length Recv: " + strconv.Itoa(length))
		fmt.Println("Received:" + string(rawData))
		rawData = rawData[:length]
		if err != nil {
			fmt.Println(err)
			break
		}
		if length > 0 {
			fmt.Println("Processing data ")
			go worker.process(rawData)
		} else {
			fmt.Println("zero length")
		}
	}
}

func entityMap(e *Entity) {
	e.Data += 1
}

func main() {
	flagMode := flag.String("mode", "server", "start in client or server mode")
	flag.Parse()

	if strings.ToLower(*flagMode) == "server" {
		array := make([]Entity, 100)
		for index := range array {
			array[index] = Entity{Data: 1, Index: index}
		}
		startWorkerManager(array)
	} else {
		startWorker()
	}

}
