# go-terminator

## 概述
```
go程序的中断前兜底函数执行包，帮助go程序员在遇到可控中断信号时及时执行弥补方法。
```

## 使用方法
[示例代码 <https://github.com/biandoucheng/open-example/tree/main/go-terminator-example>](https://github.com/biandoucheng/open-example/tree/main/go-terminator-example)
```
package test

import (
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"

	terminator "github.com/biandoucheng/go-terminator"
)

// Counting service
type ACountServer struct {
	stop   bool
	StopAt time.Time

	Count  int
	Period time.Duration
}

func (a *ACountServer) Init(period time.Duration) {
	if period <= time.Second*1 {
		period = time.Second
	}
	a.Period = time.Second
	a.Count = 0
}

func (a *ACountServer) Add() {
	a.Count += 1
	fmt.Println("Current Count: ", a.Count)
}

func (a *ACountServer) Terminated() {
	a.stop = true
	a.StopAt = time.Now()
	fmt.Println("ACountServer is stopping: ", a.StopAt.String())
}

func (a *ACountServer) Run() {
	tricker := time.NewTicker(a.Period)

	for {
		if a.stop {
			fmt.Println("ACountServer is stopped .")
			break
		}

		<-tricker.C
		a.Add()
	}
}

func TestInterrupt(t *testing.T) {
	// Initialize interrupt listening
	T := terminator.TerminatedHandler{}
	T.Init([]os.Signal{syscall.SIGINT})
	T.Run()

	// Start counting service
	A := ACountServer{}
	A.Init(time.Second * 1)
	// Interrupt handling method of registration counting service
	T.Register("ACountServer.Terminated", 0, time.Second*10, A.Terminated)
	go A.Run()

	for {
		time.Sleep(time.Second * 1)
	}
}

// Enter the test directory and execute the following command: go test -v -run TestInterrupt
// Wait a few seconds
// Press "Control + C"
// output:
/*
=== RUN   TestInterrupt
Current Count:  1
Current Count:  2
Current Count:  3
Current Count:  4
^Cgo-terminator: Service interrupted by interrupt signal,and the service is exiting gracefully .
go-terminator: func ACountServer.Terminated[0] start running
.ACountServer is stopping:  2023-01-16 16:31:54.836863 +0800 CST m=+4.519889251
go-terminator: func ACountServer.Terminated[0] is done
.signal: interrupt
FAIL    github.com/biandoucheng/go-terminator/test      4.795s
*/

```