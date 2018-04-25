package main

import (
	"time"
	"fmt"
)

// 线程间通信的消息队列
// 使用无锁的循环队列实现一个消费者着一个生产者的模型
type queue struct {
	arr  []string
	head int
	tail int
	size int
}

func (q *queue) Init(size int) {
	q.size = size + 1
	q.arr = make([]string, q.size)
	q.head, q.tail = 0, 0
}

func (q *queue) Len() int {
	return (q.head + q.size - q.tail) % q.size
}

func (q *queue) Push(x string) {
	if !q.IsFull() {
		q.arr[q.head] = x
		q.head = (q.head + 1) % q.size
	}
}

func (q *queue) Pop() (x string) {
	if !q.IsEmpty() {
		x = q.arr[q.tail]
		q.tail = (q.tail + 1) % q.size
	}
	return
}

func (q *queue) IsEmpty() bool {
	return q.Len() == 0
}

func (q *queue) IsFull() bool {
	return q.Len() == q.size-1
}

func producer() {
	ticker := time.NewTicker(2 * time.Second)
	for i := 0; i < MAXSIZE; i++ {
		<-ticker.C
		q.Push(fmt.Sprintf("msg_%d\n", i))
		if i == 0 {
		}
	}
	producerExit <- true
}

func consumer() {
	ticker := time.NewTicker(3 * time.Second)
	flag := true
	for flag || !q.IsEmpty() {
		<-ticker.C
		select {
		case <-producerExit:
			flag = false
		default:

		}
		x := q.Pop()
		fmt.Print(x)
	}
	consumerExit<-true
}

var q queue

const MAXSIZE = 10

var producerExit = make(chan bool)
var consumerExit = make(chan bool)

func init() {
	q = queue{}
	q.Init(MAXSIZE)
}

func main() {
	go producer()
	go consumer()
	<-consumerExit
}
