// +build linux darwin

package proc

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/tal-tech/go-zero/core/logx"
)

const timeFormat = "0102150405"


//
// todo x: 信号控制逻辑:
//		- go main 启动, 调用到 init(), go 一个 routine, 用于监听 os.Signal.
//		- 当有 syscall.SIGTERM() 信号, 执行 gracefulStop() 方法,
//			- 在 http server 中, 会触发 fn() 执行, 这里的 fn 指的是 srv.Shutdown(), 通知 http server 退出.
//			- 等待一个 max 退出时间, 若超时, 执行 force kill 进程.
//
func init() {

	//
	// todo x: 特别注意 ! ! !
	//		- 这个实现方式, 挺特别!
	//
	go func() {
		var profiler Stopper

		// https://golang.org/pkg/os/signal/#Notify
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGTERM)

		for {
			v := <-signals
			switch v {
			case syscall.SIGUSR1:
				dumpGoroutines()
			case syscall.SIGUSR2:
				if profiler == nil {
					profiler = StartProfile()
				} else {
					profiler.Stop()
					profiler = nil
				}
			case syscall.SIGTERM:
				//
				// todo x: 优雅退出方式
				//
				gracefulStop(signals)
			default:
				logx.Error("Got unregistered signal:", v)
			}
		}
	}()
}
