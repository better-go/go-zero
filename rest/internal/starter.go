package internal

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/tal-tech/go-zero/core/proc"
)

// todo x: HTTP server
func StartHttp(host string, port int, handler http.Handler) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	server := buildHttpServer(addr, handler)

	//
	// todo x: 进程优雅退出, 赞! ! !
	//
	gracefulOnShutdown(server)
	return server.ListenAndServe()
}

//
// todo x: HTTPS server
//
func StartHttps(host string, port int, certFile, keyFile string, handler http.Handler) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	if server, err := buildHttpsServer(addr, handler, certFile, keyFile); err != nil {
		return err
	} else {
		//
		// todo x: 进程优雅退出, 赞! ! !
		//
		gracefulOnShutdown(server)
		// certFile and keyFile are set in buildHttpsServer
		return server.ListenAndServeTLS("", "")
	}
}

func buildHttpServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{Addr: addr, Handler: handler}
}

func buildHttpsServer(addr string, handler http.Handler, certFile, keyFile string) (*http.Server, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	config := tls.Config{Certificates: []tls.Certificate{cert}}
	return &http.Server{
		Addr:      addr,
		Handler:   handler,
		TLSConfig: &config,
	}, nil
}


//
// todo x: 进程优雅退出
//
func gracefulOnShutdown(srv *http.Server) {

	//
	// todo x: 注意内部的 fn 指的是: srv.Shutdown() 动作
	//
	proc.AddWrapUpListener(func() {
		//
		// todo x: 标准库的 http 退出, 在另外一个 go routine 监控 os.Signal 中, 触发调用
		//
		srv.Shutdown(context.Background())
	})
}
