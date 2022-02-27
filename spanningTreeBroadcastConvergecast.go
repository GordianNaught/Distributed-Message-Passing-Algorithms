package main

import "fmt"

type message interface{}

type startMessage struct{}
type id int

func (i id) toString() string {
	return fmt.Sprintf("%q", int(i))
}

type goMessage struct {
	sender id
}
type backMessage struct {
	val_set []id
}

type process struct {
	receivedFirstMessage bool
	parent               id
	self                 id
	neighbors            []id
	children             []id
	channelMap           map[id]chan message
	expectedMessages     int
	done                 chan struct{}
}

func (p *process) channel() chan message {
	return p.channelMap[p.self]
}

func (p *process) toString() string {
	return p.self.toString()
}

func (p *process) send(target id, m message) {
	p.channelMap[target] <- m
	fmt.Printf("%v sent %#v to %v\n", p.self, m, target)
}
func (p *process) execute() {
	for m := range p.channel() {
		switch m := m.(type) {
		case startMessage:
			p.parent = p.self
			p.expectedMessages = len(p.neighbors)
			for _, neighbor := range p.neighbors {
				go p.send(neighbor, goMessage{p.self})
			}
		case goMessage:
			if !p.receivedFirstMessage {
				p.receivedFirstMessage = true
				p.parent = m.sender
				p.expectedMessages = len(p.neighbors) - 1
				if p.expectedMessages == 0 {
					go p.send(m.sender, backMessage{[]id{p.self}})
					continue
				}
				for _, neighbor := range p.neighbors {
					if neighbor == m.sender {
						continue
					}
					go p.send(neighbor, goMessage{p.self})
				}
				continue
			}
			go p.send(m.sender, backMessage{[]id{}})
		case backMessage:
			p.children = append(p.children, m.val_set...)
			p.expectedMessages--
			if p.expectedMessages == 0 {
				if p.parent == p.self {
					p.done <- struct{}{}
				} else {
					go p.send(p.parent, backMessage{append(p.children, p.self)})
				}
			}
		}
	}
}

func main() {
	numProcesses := 4
	channelMap := make(map[id]chan message)
	for i := 1; i < numProcesses+1; i++ {
		channelMap[id(i)] = make(chan message)
	}
	processes := make([]*process, numProcesses)
	doneCollector := make(chan struct{})
	processes[0] = &process{
		receivedFirstMessage: false,
		parent:               id(1),
		self:                 id(1),
		neighbors:            []id{id(2), id(3)},
		channelMap:           channelMap,
		expectedMessages:     0,
		done:                 doneCollector,
	}
	processes[1] = &process{
		receivedFirstMessage: false,
		parent:               id(2),
		self:                 id(2),
		neighbors:            []id{id(1), id(3)},
		channelMap:           channelMap,
		expectedMessages:     0,
		done:                 doneCollector,
	}
	processes[2] = &process{
		receivedFirstMessage: false,
		parent:               id(3),
		self:                 id(3),
		neighbors:            []id{id(1), id(2), id(4)},
		channelMap:           channelMap,
		expectedMessages:     0,
		done:                 doneCollector,
	}
	processes[3] = &process{
		receivedFirstMessage: false,
		parent:               id(4),
		self:                 id(4),
		neighbors:            []id{id(3)},
		channelMap:           channelMap,
		expectedMessages:     0,
		done:                 doneCollector,
	}
	for _, p := range processes {
		go p.execute()
	}
	go processes[0].send(1, startMessage{})
	<-doneCollector
	fmt.Println("complete, p1 has collected:", processes[0].children)
}
