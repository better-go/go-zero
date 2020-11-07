// +build linux darwin

package proc

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/tal-tech/go-zero/core/logx"
)

const (
	wrapUpTime = time.Second
	// why we use 5500 milliseconds is because most of our queue are blocking mode with 5 seconds
	waitTime = 5500 * time.Millisecond // todo x: 默认最长等待时间
)

var (
	//
	// todo x: 注意这2个全局变量, 初始化的时机! ! ! 非常关键! ! !
	//
	wrapUpListeners          = new(listenerManager) // todo x: 这个是 http server 启动时, 依赖的全局变量
	shutdownListeners        = new(listenerManager) // todo x: 初始化的位置: ServiceGroup().Start()
	delayTimeBeforeForceQuit = waitTime             // todo x: 默认最长等待时间, 超时, 会 force kill
)

//
// todo x: 全局变量, 注意这个方法, 调用位置!
//		- 初始化位置在这里: go-zero/core/service/servicegroup.go
//			- ServiceGroup().Start()
//
func AddShutdownListener(fn func()) (waitForCalled func()) {
	return shutdownListeners.addListener(fn)
}

//
// todo x: 全局变量, 注意这个方法, 调用位置!
//
func AddWrapUpListener(fn func()) (waitForCalled func()) {
	//
	// todo x: 这里的 fn, 指的是 srv.Shutdown() 动作
	//
	return wrapUpListeners.addListener(fn)
}

func SetTimeToForceQuit(duration time.Duration) {
	delayTimeBeforeForceQuit = duration
}

//
// todo x: 上层调用是在 init() 方法里
//
func gracefulStop(signals chan os.Signal) {
	signal.Stop(signals)

	logx.Info("Got signal SIGTERM, shutting down...")

	//
	// todo x: 这个全局变量, 注意啥时候赋值的?
	// 		- 是在 engine.start() 启动的时候, 把 fn() 塞进来.
	//
	wrapUpListeners.notifyListeners()

	time.Sleep(wrapUpTime)

	//
	// todo x: 注意这个全局变量, 啥时候初始化的?
	//
	shutdownListeners.notifyListeners()

	// todo x: 默认最长等待时间, 超过会 force kill
	time.Sleep(delayTimeBeforeForceQuit - wrapUpTime)

	logx.Infof("Still alive after %v, going to force kill the process...", delayTimeBeforeForceQuit)
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
}

type listenerManager struct {
	lock      sync.Mutex
	waitGroup sync.WaitGroup
	listeners []func()
}

//
//
//
func (lm *listenerManager) addListener(fn func()) (waitForCalled func()) {
	lm.waitGroup.Add(1)

	lm.lock.Lock()

	lm.listeners = append(lm.listeners, func() {
		defer lm.waitGroup.Done()

		//
		// todo x: task fn() here!
		//
		fn()
	})
	lm.lock.Unlock()

	return func() {
		lm.waitGroup.Wait()
	}
}

func (lm *listenerManager) notifyListeners() {
	lm.lock.Lock()
	defer lm.lock.Unlock()

	for _, listener := range lm.listeners {
		listener()
	}
}
