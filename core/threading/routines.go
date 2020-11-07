package threading

import (
	"bytes"
	"runtime"
	"strconv"

	"github.com/tal-tech/go-zero/core/rescue"
)

func GoSafe(fn func()) {
	//
	// todo x: 很多人都写过类似实现. 啊哈哈...
	//
	go RunSafe(fn)
}

// Only for debug, never use it in production
func RoutineId() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	// if error, just return 0
	n, _ := strconv.ParseUint(string(b), 10, 64)

	return n
}


//
// todo x: 打印错误栈
//
func RunSafe(fn func()) {
	defer rescue.Recover()


	//
	// todo x: call fn()
	//
	fn()
}
