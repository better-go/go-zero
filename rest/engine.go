package rest

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/justinas/alice"
	"github.com/tal-tech/go-zero/core/codec"
	"github.com/tal-tech/go-zero/core/load"
	"github.com/tal-tech/go-zero/core/stat"
	"github.com/tal-tech/go-zero/rest/handler"
	"github.com/tal-tech/go-zero/rest/httpx"
	"github.com/tal-tech/go-zero/rest/internal"
	"github.com/tal-tech/go-zero/rest/router"
)

// use 1000m to represent 100%
const topCpuUsage = 1000

var ErrSignatureConfig = errors.New("bad config for Signature")

//
// todo x: 框架引擎对象:
//
type engine struct {
	conf                 RestConf                     // todo x: 配置
	routes               []featuredRoutes             // 路由
	unauthorizedCallback handler.UnauthorizedCallback //
	unsignedCallback     handler.UnsignedCallback     //
	middlewares          []Middleware                 // todo x: 预留 hook 口子, 用于扩展插件列表 Server.Use() 使用
	shedder              load.Shedder                 //
	priorityShedder      load.Shedder                 //
}

//
//
//
func newEngine(c RestConf) *engine {
	srv := &engine{
		conf: c,
	}

	// todo x: 用来做降载, default=900, range=[0:1000]
	if c.CpuThreshold > 0 {
		//
		// todo x:
		//
		srv.shedder = load.NewAdaptiveShedder(load.WithCpuThreshold(c.CpuThreshold))
		//
		//
		srv.priorityShedder = load.NewAdaptiveShedder(load.WithCpuThreshold(
			(c.CpuThreshold + topCpuUsage) >> 1))
	}

	return srv
}

func (s *engine) AddRoutes(r featuredRoutes) {
	s.routes = append(s.routes, r)
}

func (s *engine) SetUnauthorizedCallback(callback handler.UnauthorizedCallback) {
	s.unauthorizedCallback = callback
}

func (s *engine) SetUnsignedCallback(callback handler.UnsignedCallback) {
	s.unsignedCallback = callback
}

////////////////////////////////////////////////////////////////////////////////////////////////////////

//
// todo x: 启动 http + 路由
//
func (s *engine) Start() error {

	//
	// todo x: 特别注意
	//		- 1. router.NewRouter()
	//			- 返回值类型是: httpx.Router, 这是扩展的 	http.Handler interface 类型
	//			- 这个是很妙的地方. 要 follow 这条线.
	//
	return s.StartWithRouter(router.NewRouter())
}

//
// todo: 启动路由, 自动 http/https
//		- 接上, 注意参数类型(interface 类型)
//		- http server 支持 优雅退出
//
func (s *engine) StartWithRouter(router httpx.Router) error {
	//
	//
	// todo x: 1. 路由绑定
	//		- 注意内部一些实现细节
	//			- 特别有价值的部分! 重要事情说3遍!
	//			- 特别有价值的部分! 重要事情说3遍!
	//			- 特别有价值的部分! 重要事情说3遍!
	//		- 默认集成很多有用插件
	//		- server.Use() 开了 hook 口子, 可以挂自定义 Middleware
	//
	if err := s.bindRoutes(router); err != nil {
		return err
	}

	//-------------------------------------------------------------------------------------

	// todo x: 启动 http server
	if len(s.conf.CertFile) == 0 && len(s.conf.KeyFile) == 0 {
		//
		// todo x: HTTP server, 内部支持进程优雅退出, 赞!
		//
		return internal.StartHttp(s.conf.Host, s.conf.Port, router)
	}

	return internal.StartHttps(s.conf.Host, s.conf.Port, s.conf.CertFile, s.conf.KeyFile, router)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////

func (s *engine) appendAuthHandler(fr featuredRoutes, chain alice.Chain,
	verifier func(alice.Chain) alice.Chain) alice.Chain {

	//
	// todo x: jwt check
	//
	if fr.jwt.enabled {
		if len(fr.jwt.prevSecret) == 0 {
			chain = chain.Append(handler.Authorize(fr.jwt.secret,
				handler.WithUnauthorizedCallback(s.unauthorizedCallback)))
		} else {
			chain = chain.Append(handler.Authorize(fr.jwt.secret,
				handler.WithPrevSecret(fr.jwt.prevSecret),
				handler.WithUnauthorizedCallback(s.unauthorizedCallback)))
		}
	}

	// todo x: 闭包调用
	return verifier(chain)
}

func (s *engine) bindFeaturedRoutes(router httpx.Router, fr featuredRoutes, metrics *stat.Metrics) error {
	//
	// todo x: 注意, 返回值是个函数!
	//
	verifier, err := s.signatureVerifier(fr.signature)
	if err != nil {
		return err
	}

	for _, route := range fr.routes {
		//
		// todo x: 路由绑定(这里注册了很多强大的插件)
		//
		if err := s.bindRoute(fr, router, metrics, route, verifier); err != nil {
			return err
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////

//
// todo x: supper awesome here!
//
func (s *engine) bindRoute(fr featuredRoutes, router httpx.Router, metrics *stat.Metrics,
	route Route, verifier func(chain alice.Chain) alice.Chain) error {
	//
	// todo x: 默认集成的插件列表: 初始化了一堆有用的中间件
	//		- 特别注意看一下 alice.New() 这个包的实现. 蛮有用的lib
	//
	chain := alice.New(
		//
		// todo x
		//
		handler.TracingHandler,
		s.getLogHandler(),
		handler.MaxConns(s.conf.MaxConns),
		handler.BreakerHandler(route.Method, route.Path, metrics),
		handler.SheddingHandler(s.getShedder(fr.priority), metrics),
		handler.TimeoutHandler(time.Duration(s.conf.Timeout)*time.Millisecond),
		handler.RecoverHandler,
		handler.MetricHandler(metrics),
		//
		// todo x:
		//
		handler.PromethousHandler(route.Path),
		handler.MaxBytesHandler(s.conf.MaxBytes),
		handler.GunzipHandler,
	)

	//-------------------------------------------------------------------------------------

	//
	// todo x: auth check, 这里有 jwt 支持
	//
	chain = s.appendAuthHandler(fr, chain, verifier)

	//-------------------------------------------------------------------------------------

	//
	// todo x: 自定义的插件列表
	//
	for _, middleware := range s.middlewares {
		//
		// todo x: 这里有个 interface 接口适配 技巧写法
		//
		chain = chain.Append(convertMiddleware(middleware))
	}

	//
	// todo x: 第三方 lib
	//
	handle := chain.ThenFunc(route.Handler)

	//
	// todo x: router 是 interface 类型, 注意这个使用技巧
	//
	return router.Handle(route.Method, route.Path, handle)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////

func (s *engine) bindRoutes(router httpx.Router) error {
	metrics := s.createMetrics()

	for _, fr := range s.routes {
		if err := s.bindFeaturedRoutes(router, fr, metrics); err != nil {
			return err
		}
	}

	return nil
}

func (s *engine) createMetrics() *stat.Metrics {
	var metrics *stat.Metrics

	if len(s.conf.Name) > 0 {
		metrics = stat.NewMetrics(s.conf.Name)
	} else {
		metrics = stat.NewMetrics(fmt.Sprintf("%s:%d", s.conf.Host, s.conf.Port))
	}

	return metrics
}

func (s *engine) getLogHandler() func(http.Handler) http.Handler {
	if s.conf.Verbose {
		return handler.DetailedLogHandler
	} else {
		return handler.LogHandler
	}
}

func (s *engine) getShedder(priority bool) load.Shedder {
	if priority && s.priorityShedder != nil {
		return s.priorityShedder
	}
	return s.shedder
}

//
// todo x: 注意: 返回值, 是一个函数, 用于闭包调用
//
func (s *engine) signatureVerifier(signature signatureSetting) (func(chain alice.Chain) alice.Chain, error) {
	if !signature.enabled {
		return func(chain alice.Chain) alice.Chain {
			return chain
		}, nil
	}

	if len(signature.PrivateKeys) == 0 {
		if signature.Strict {
			return nil, ErrSignatureConfig
		} else {
			return func(chain alice.Chain) alice.Chain {
				return chain
			}, nil
		}
	}

	decrypters := make(map[string]codec.RsaDecrypter)
	for _, key := range signature.PrivateKeys {
		fingerprint := key.Fingerprint
		file := key.KeyFile
		decrypter, err := codec.NewRsaDecrypter(file)
		if err != nil {
			return nil, err
		}

		decrypters[fingerprint] = decrypter
	}

	return func(chain alice.Chain) alice.Chain {
		if s.unsignedCallback != nil {
			return chain.Append(handler.ContentSecurityHandler(
				decrypters, signature.Expiry, signature.Strict, s.unsignedCallback))
		} else {
			return chain.Append(handler.ContentSecurityHandler(
				decrypters, signature.Expiry, signature.Strict))
		}
	}, nil
}

//
//
//
func (s *engine) use(middleware Middleware) {
	s.middlewares = append(s.middlewares, middleware)
}

//
// todo x: 技巧代码, 这里返回 http.Handler 接口类型
//		- 这是类型适配技巧. 很赞!
//
func convertMiddleware(ware Middleware) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		//
		//
		//
		return ware(next.ServeHTTP)
	}
}
