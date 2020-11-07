package main

import (
	"flag"
	"fmt"

	"greet/internal/config"
	"greet/internal/handler"
	"greet/internal/svc"

	"github.com/tal-tech/go-zero/core/conf"
	"github.com/tal-tech/go-zero/rest"
)

var configFile = flag.String("f", "etc/greet-api.yaml", "the config file")


/*

todo x:
	1. 完整的 http rest api 示例
	2. 通过 main.go, 可以找到关键模块:
		- rest.MustNewServer()
		- handler.RegisterHandlers()
		- server.Start()
	3. 通俗讲: web 框架, 最主要的部分:
		- server 对象,
		- router 对象,
		- start(), stop() 机制.
		- 中间件 hook 机制
	4. 其他部分, 都是非重点, 可不必花太多精力看. (80/20定理)


*/
func main() {
	// todo x: 1. 命令行参数解析
	flag.Parse()

	// todo x: 2. 配置文件解析
	var c config.Config
	conf.MustLoad(*configFile, &c)

	//
	// todo x: 3. 业务 ctx 对象.
	// 	- 这个不是 go.ctx, 是 业务 ctx. 目前看只是传了 全局 config.
	//
	ctx := svc.NewServiceContext(c)

	// todo x: 4. 创建 http server
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// todo x: 5. 注册 http 路由
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)

	// todo x 6. 启动 http server
	server.Start()
}
