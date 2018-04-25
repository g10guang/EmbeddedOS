package main

/*
事件驱动调度
*/

import (
	"fmt"
	"time"
	"sync"
)

type eventMap struct {
	index  int
	length int
	bitmap []int
	ticker *time.Ticker
	mux    sync.Mutex
}

func (s *eventMap) set(index int) {
	word, bit := index/32, uint(index%32)
	s.lock()
	s.bitmap[word] |= 1 << bit
}

func (s *eventMap) clear(index int) {
	word, bit := index/32, uint(index%32)
	s.lock()
	s.bitmap[word] &= ^(1 << bit)
}

func (s *eventMap) lock() {
	scheduler.mux.Lock()
	defer scheduler.mux.Unlock()
}

func (s *eventMap) has(index int) bool {
	word, bit := index/32, uint(index%32)
	s.lock()
	return word < len(s.bitmap) && s.bitmap[word]&(1<<bit) != 0
}

var scheduler eventMap

func interrupt(signal chan bool) {
	for {
		<-scheduler.ticker.C
		scheduler.index--
		if scheduler.index <= 0 {
			break
		}
		scheduler.set(scheduler.index)
	}
	// interrupt exit
	signal <- true
}

func init() {
	scheduler = eventMap{index: 7, length: 7}
	scheduler.bitmap = make([]int, scheduler.index/32+1)
	scheduler.ticker = time.NewTicker(time.Second)
}

func eventScheduler(mainWait chan bool, interruptExit chan bool) {
	flag := true
	exitFlag := false
	for !exitFlag || flag {
		select {
		case <-interruptExit:
			exitFlag = true
			//	if not default and no channel in cases can read or write, select will block.
		default:
			break
		}
		flag = false
		for i := 1; i < scheduler.length; i++ {
			if scheduler.has(i) {
				flag = true
				ticker := time.NewTicker(time.Millisecond * 600)
				for t := 0; t < i; t++ {
					<-ticker.C
					fmt.Printf("task_%d is running\n", i)
				}
				// clear this event
				scheduler.clear(i)
				i = 0
			}
		}
	}
	// event scheduler exit
	mainWait <- true
}

func main() {
	// create a ticker to do something after 1s
	waitChan := make(chan bool)
	go interrupt(waitChan)
	mainWait := make(chan bool)
	go eventScheduler(mainWait, waitChan)
	<-mainWait
}
