package main

import "fmt"

type message interface{}

type goMessage struct {
}

type backMessage struct {
	ids []rune
}

type process struct {
	id               rune
	channelMap       map[rune]chan message
	parent           rune
	children         []rune
	expectedMessages int
	collected        []rune
	doneChannel      chan struct{}
}

func (p *process) channel() chan message {
	return p.channelMap[p.id]
}

func (p *process) execute() {
	for received := range p.channel() {
		switch m := received.(type) {
		case goMessage:
			fmt.Printf("%c received go message\n", p.id)
			// if not root and no children
			if p.id != p.parent && len(p.children) == 0 {
				go func(parentChannel chan message, m backMessage) {
					parentChannel <- m
				}(p.channelMap[p.parent], backMessage{p.collected})
				fmt.Printf("%c completed with |%s|\n", p.id, string(p.collected))
				return
			}
			for _, child := range p.children {
				go func(childChannel chan message, m goMessage) {
					childChannel <- m
				}(p.channelMap[child], goMessage{})
			}
		case backMessage:
			fmt.Printf("%c received back message of |%s|\n", p.id, string(m.ids))
			p.collected = append(p.collected, m.ids...)
			p.expectedMessages--
			if p.expectedMessages == 0 {
				// if root
				if p.id == p.parent {
					p.doneChannel <- struct{}{}
				} else {
					go func(parentChannel chan message, m backMessage) {
						parentChannel <- m
					}(p.channelMap[p.parent], backMessage{p.collected})
				}
				fmt.Printf("%c completed with |%s|\n", p.id, string(p.collected))
				return
			}
		default:
			fmt.Println("unhandled message:", m)
		}
	}
}

func main() {
	channelMap := make(map[rune]chan message)
	numProcesses := 8
	for _, name := range "abcfgijk" {
		channelMap[name] = make(chan message, 0)
	}
	doneCollector := make(chan struct{})
	processes := make([]*process, numProcesses)
	processes[0] = &process{'a', channelMap, 'a', []rune{'i', 'b'}, 2, []rune{'a'}, doneCollector}
	processes[1] = &process{'b', channelMap, 'a', []rune{'c'}, 1, []rune{'b'}, doneCollector}
	processes[2] = &process{'c', channelMap, 'b', []rune{}, 0, []rune{'c'}, doneCollector}
	processes[3] = &process{'f', channelMap, 'k', []rune{}, 0, []rune{'f'}, doneCollector}
	processes[4] = &process{'g', channelMap, 'k', []rune{}, 0, []rune{'g'}, doneCollector}
	processes[5] = &process{'i', channelMap, 'a', []rune{'k', 'j'}, 2, []rune{'i'}, doneCollector}
	processes[6] = &process{'j', channelMap, 'i', []rune{}, 0, []rune{'j'}, doneCollector}
	processes[7] = &process{'k', channelMap, 'i', []rune{'g', 'f'}, 2, []rune{'k'}, doneCollector}
	for _, p := range processes {
		go p.execute()
	}
	channelMap['a'] <- goMessage{}
	<-doneCollector
}
