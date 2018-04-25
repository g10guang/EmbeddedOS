package main

import (
	"errors"
	"fmt"
	"time"
	"sync"
	"os"
	"strconv"
	"strings"
)

/*
动态内存分配
*/

type MemBlock struct {
	Arr       []byte
	Size      int
	allocTime time.Time
	next      *MemBlock
	link      *MemLink
	isAlloc   bool
	no        int // 记录这是第几块，用于输出链表
}

type MemLink struct {
	size    int
	freeNum int
	head    *MemBlock
	// 记录所有的指针，用于日志记录
	data []*MemBlock
	mux  sync.Mutex
}

/*
申请分配内存
*/
func (s *MemLink) alloc() (block *MemBlock) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.freeNum > 0 {
		block = s.head
		s.freeNum--
		if s.freeNum == 0 {
			s.head = nil
		} else {
			s.head = s.head.next
		}
		// here remove the relationship from the link
		block.next = nil
		block.allocTime = time.Now()
		block.isAlloc = true
	}
	return
}

/*
内存回收
*/
func (s *MemBlock) free() {
	s.link.mux.Lock()
	defer s.link.mux.Unlock()
	var link *MemLink
	switch s.link {
	case l30:
		link = l30
	case l60:
		link = l60
	default:
		panic(errors.New("unknown mem link"))
	}
	s.Size = link.size
	s.isAlloc = false
	// clear the content
	for i := 0; i < s.Size; i++ {
		s.Arr[i] = 0
	}
	// insert the free node to the head of link.
	if link.freeNum != 0 {
		s.next = link.head
	}
	link.head = s
	link.freeNum++
}

/*
初始化时整理队列
*/
func (s *MemLink) makeLink() {
	s.head = new(MemBlock)
	s.head.link = s
	s.data = make([]*MemBlock, s.freeNum)
	s.head.Arr = make([]byte, s.size)
	s.head.Size = s.size
	s.head.no = 0
	s.data[0] = s.head
	for i, t := 1, s.head; i < s.freeNum; i++ {
		nb := new(MemBlock)
		nb.link = s
		nb.Arr = make([]byte, s.size)
		nb.Size = s.size
		nb.no = i
		s.data[i] = nb
		t.next = nb
		t = nb
	}
}

func OS_malloc(size int) (block *MemBlock) {
	var link *MemLink
	switch {
	case size <= 30:
		link = l30
	case size <= 60:
		link = l60
	default:
		panic(errors.New(fmt.Sprintf("Cannot malloc Size=%d which >= 60\n", size)))
	}
	block = link.alloc()
	block.Size = size
	return
}

func OS_free(block *MemBlock) {
	block.free()
}

type Blocks struct {
	blocks []*MemBlock
	mux    sync.Mutex
}

func (b *Blocks) addBlock(n *MemBlock) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.blocks = append(b.blocks, n)
}

func (b *Blocks) freeBlock() (n *MemBlock) {
	b.mux.Lock()
	defer b.mux.Unlock()
	if len(b.blocks) > 0 {
		// remove the first. because it is the first block to be malloced.
		n = b.blocks[0]
		b.blocks = b.blocks[1:]
	}
	return
}

/*
测试释放入口
*/
func testFree() {
	flag := false
	// malloc not exit or some block has not been free.
	for !flag || len(blockAlloc.blocks) > 0 {
		for len(blockAlloc.blocks) > 0 && time.Now().Sub(blockAlloc.blocks[0].allocTime) > 3*time.Second {
			// 释放第一块
			OS_free(blockAlloc.freeBlock())
			logChan<-buildLogMsg("释放一个内存块")
		}
		select {
		case <-allocExitSignal:
			flag = true
		default:

		}
	}
	freeExitSignal <- true
}

/*
测试分配入口
*/
func testMalloc() {
	suits := []int{18, 46, 16, 35, 13, 38, 22, 32, 25, 43, 20, 39, 11, 41}
	for i, size := range suits {
		<-mallocTicker.C
		b := OS_malloc(size)
		blockAlloc.addBlock(b)
		// set the content to be i is not the real requirements.
		for k := 0; k < b.Size; k++ {
			b.Arr[k] = byte(i)
		}
		logChan<-buildLogMsg("分配一块内存")
	}
	allocExitSignal <- true
}

var l30 *MemLink

var l60 *MemLink

var mallocTicker *time.Ticker

var blockAlloc *Blocks

var allocExitSignal chan bool

var freeExitSignal chan bool

var logFile *os.File

var logChan chan *strings.Builder

/*
初始化内存区块
*/
func init() {
	l30 = new(MemLink)
	l30.size = 30
	l30.freeNum = 10
	l30.makeLink()
	l60 = new(MemLink)
	l60.size = 60
	l60.freeNum = 10
	l60.makeLink()
	// init ticker for malloc and free
	mallocTicker = time.NewTicker(time.Second * 2)

	blockAlloc = new(Blocks)

	allocExitSignal = make(chan bool)
	freeExitSignal = make(chan bool)

	var err error

	logFile, err = os.Create("./dynamicMemAlloc.log")
	if err != nil {
		panic(err)
	}
	logChan = make(chan *strings.Builder, 10)
	logChan<-buildLogMsg("init-初始化")
}

/*
打印日志用于判断分配是否正确
*/
func log() {
	for bp := range logChan {
		logFile.WriteString(bp.String())
	}
}

func buildLogMsg(msg string) (bp *strings.Builder) {
	bp = &strings.Builder{}
	fmt.Fprint(bp, "\n\n")
	fmt.Fprintln(bp, msg)
	bp = showBlockContent(l30, bp)
	bp = showBlockContent(l60, bp)
	return
}

func showBlockContent(link *MemLink, bp *strings.Builder) *strings.Builder {
	link.mux.Lock()
	defer link.mux.Unlock()
	fmt.Fprintf(bp, "\n%d*10bytes 区间已分配 %d 块，剩余 %d 块\n", link.size, len(link.data)-link.freeNum, link.freeNum)
	fmt.Fprintln(bp, "分配信息:")
	for i, block := range link.data {
		if block.isAlloc {
			fmt.Fprintf(bp, "No.%d allocated size=%d\n", i, block.Size)
		} else {
			fmt.Fprintf(bp, "No.%d free size=%d\n", i, block.Size)
		}
		fmt.Fprintf(bp, "content is: %v\n", block.Arr)
	}
	// 输出链表
	t := link.head
	linkOrder := make([]string, 0)
	for t != nil {
		linkOrder = append(linkOrder, strconv.Itoa(t.no))
		t = t.next
	}
	fmt.Fprintln(bp, strings.Join(linkOrder, "-->"))
	return bp
}

func main() {
	go testMalloc()
	go testFree()
	go log()
	<-freeExitSignal
	time.Sleep(time.Second)
	close(logChan)
}
