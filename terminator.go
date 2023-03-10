package goterminator

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

/*
..................................................................................................
. TerminatedHandler Program interrupt handling handle
. Implement a series of rescue methods before program interruption
.
. Init   	Initialize the interrupt handle and set the monitoring signal
. Register  Register the interrupt rescue method, set the method name (the previous one will be overwritten if it is repeated), the execution priority (the execution order is unpredictable if the same priority is the same), the execution timeout, and the execution method
. Remove	Remove a method from the registered map
. RunFuncs	Execute each method in order of priority
. Listen	Monitor the program interrupt signal and trigger the execution of the rescue method
. Run		Execute Listen in child thread
..................................................................................................
*/
type TerminatedHandler struct {
	sync.Mutex
	listenChan   chan os.Signal           // Listening channel
	stoped       bool                     // stop or not
	sigs         []os.Signal              // Target monitoring signal
	funcTimeMap  map[string]time.Duration // Map recording method execution timeout
	funcPriority map[string]int           // Dictionary that records the priority of method execution
	funcMap      map[string]func()        // A dictionary that records the execution function
}

var (
	// Default target monitoring signal
	defaultListenSigs = []os.Signal{syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2}
)

func (t *TerminatedHandler) Init(sigs []os.Signal) {
	t.Lock()
	if len(sigs) > 0 {
		t.sigs = sigs
	} else {
		t.sigs = defaultListenSigs
	}
	t.stoped = false
	t.listenChan = make(chan os.Signal, 1)
	t.funcTimeMap = map[string]time.Duration{}
	t.funcPriority = map[string]int{}
	t.funcMap = map[string]func(){}
	t.Unlock()
}

func (t *TerminatedHandler) Register(name string, priority int, timeout time.Duration, fun func()) bool {
	t.Lock()
	defer t.Unlock()

	if t.stoped {
		return false
	}

	if priority < 0 {
		priority = 0
	}

	t.funcPriority[name] = priority
	t.funcMap[name] = fun
	t.funcTimeMap[name] = timeout

	return true
}

func (t *TerminatedHandler) Remove(name string) bool {
	t.Lock()
	defer t.Unlock()

	if t.stoped {
		return false
	}

	delete(t.funcMap, name)
	delete(t.funcPriority, name)
	delete(t.funcTimeMap, name)

	return true
}

func (t *TerminatedHandler) RunFuncs() {
	t.Lock()
	t.stoped = true
	funcnames := sortFuncs(t.funcPriority)

	for idx, name := range funcnames {
		fun, has := t.funcMap[name]
		if !has {
			continue
		}

		tm, has := t.funcTimeMap[name]
		if !has {
			continue
		}

		fmt.Printf("%s: func %s[%d] start running \n.", terminatorName, name, idx)
		doneFlag := false
		ct := time.Now()
		ed := ct.Add(tm)
		go runFunc(fun, &doneFlag, name)

		for ct.Before(ed) {
			if doneFlag {
				fmt.Printf("%s: func %s[%d] is done \n.", terminatorName, name, idx)
				break
			}

			time.Sleep(time.Millisecond * 100)
			ct.Add(time.Millisecond * 100)
		}
	}
	t.Unlock()
}

func (t *TerminatedHandler) Listen() {
	// Interrupt signal monitoring
	signal.Notify(t.listenChan, t.sigs...)
	sig := <-t.listenChan
	fmt.Printf(terminatedInfo, sig.String())

	// Execute the method before system interruption
	t.RunFuncs()

	// Turn off listening and reissue the interrupt command
	signal.Stop(t.listenChan)
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
}

func (t *TerminatedHandler) Run() {
	go t.Listen()
}
