package main

import "fmt"

type message struct {
	id          int
	neighborIds []int
}

type process struct {
	id            int
	started       bool
	channelMap    map[int]chan message
	state         map[int][]int
	numKnown      int
	numProcesses  int
	doneCollector chan<- struct{}
}

// Get the neighbor ids of this process.
func (p *process) neighbors() []int {
	return p.state[p.id]
}

// Get the channel for this process.
func (p *process) channel() chan message {
	return p.channelMap[p.id]
}

// Listen for and forward new information about the process graph.
func (p *process) listen() {
	for {
		received := <-p.channel()
		// if know whole graph
		if p.numKnown == p.numProcesses {
			// discard
			continue
		}
		// if haven't notified neighbors of myself yet
		if !p.started {
			p.start()
		}
		// if known
		if _, known := p.state[received.id]; known {
			// discard
			continue
		}
		// if new, forward
		fmt.Println(p.id, "learned about", received.id)
		p.numKnown++
		if p.numKnown == p.numProcesses {
			fmt.Println(p.id, "knows the whole graph")
			// signal that this process knows the complete graph
			go func() { p.doneCollector <- struct{}{} }()
		}
		p.state[received.id] = received.neighborIds
		for _, neighbor := range p.neighbors() {
			go func(channel chan message, m message, neighborId int) {
				channel <- m
				fmt.Println(p.id, "told", neighborId, "about", received.id)
			}(p.channelMap[neighbor], received, neighbor)
		}
	}
}

// Notify neighboring processes of this process.
func (p *process) start() {
	p.started = true
	for key, val := range p.state {
		for _, neighbor := range p.neighbors() {
			go func(channel chan message, m message, neighborId int) {
				channel <- m
				fmt.Println(p.id, "told", neighborId, "about", p.id)
			}(p.channelMap[neighbor], message{key, val}, neighbor)
		}
	}
}

func main() {
	channelMap := make(map[int]chan message)
	numProcesses := 5
	for i := 0; i < numProcesses; i++ {
		channelMap[i] = make(chan message, 0)
	}
	doneCollector := make(chan struct{}, 0)
	neighborMap := map[int][]int{
		0: {1},
		1: {0, 2},
		2: {1, 3, 4},
		3: {2, 4},
		4: {2, 3},
	}
	processes := make([]*process, numProcesses)
	for i := 0; i < numProcesses; i++ {
		processes[i] = &process{i, false, channelMap, map[int][]int{i: neighborMap[i]}, 1, numProcesses, doneCollector}
	}
	// make all processes listen
	for _, process := range processes {
		go process.listen()
	}
	// make process 0 kick things off
	go processes[0].start()
	// wait for all done signals
	for i := range processes {
		<-doneCollector
		fmt.Println(i+1, "processes know the whole graph")
	}
	fmt.Println("all nodes know the whole graph")
}
