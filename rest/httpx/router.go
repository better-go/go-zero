package httpx

import "net/http"

//
// todo x: 扩展了 go 标准库 http.Handler 这个 interface
//
type Router interface {
	http.Handler

	//
	// todo x: 扩展3个新方法
	//
	Handle(method string, path string, handler http.Handler) error
	SetNotFoundHandler(handler http.Handler)
	SetNotAllowedHandler(handler http.Handler)
}
