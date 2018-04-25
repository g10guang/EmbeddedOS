# 嵌入式操作系统实验

[实验指导书](./document/嵌入式操作系统指导书.pdf)

实验主要有四部分内容：

1. [事件驱动调度](./eventDriven/eventDriven.go)
2. [固定大小块的分配](./dynamicMemAlloc/dynamicMemAlloc.go)
3. [线程间通信](./threadCommunicate/threadCommunicate.go)
4. 嵌入式操作系统的代码移植

由于本实验采用的是 Go 编程语言，所以实验三和实验四实现起来没有难度，都有 Go 的原生支持。

实验三线程间通信，没有采用 channel 实现，在一个生产者一个消费者的场景下，使用了无锁的循环队列实现