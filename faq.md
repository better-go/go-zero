

# go-zero annotated FAQ:


- version: `v1.0.25`

## 阅读指南:

### 1. 准备工作:

- 根据 cli 工具, 创建 demo, 根据 demo 按图索骥, 分析框架源码.

```bash

# 安装 cli: 
-> % GO111MODULE=on GOPROXY=https://goproxy.cn/,direct go get -u github.com/tal-tech/go-zero/tools/goctl


# 创建demo 项目:

cd ./tmp
goctl api new greet



```

- demo 项目结构: 


```bash

-> % tree ./greet -L 4
./greet
├── etc
│   └── greet-api.yaml
├── go.mod
├── greet.api
├── greet.go
└── internal
    ├── config
    │   └── config.go
    ├── handler
    │   ├── greethandler.go
    │   └── routes.go
    ├── logic
    │   └── greetlogic.go
    ├── svc
    │   └── servicecontext.go
    └── types
        └── types.go

7 directories, 10 files



```


- 根据提示 run 项目. curl api, 确定一切正常. 
- 开始分析源码.



### 2. 示例 阅读: 

- 通过阅读 [./tmp/greet](./tmp/greet), 找到框架总纲. 


### 3. 框架源码分析: 

- 根据示例项目, 找到  rest.MustNewServer(c.RestConf)


#### 3.1 rest.MustNewServer

- 代码位置: 
    - go-zero/rest/server.go#MustNewServer()


#### 3.2 engine.AddRoutes():

- 注册路由表+handler
- 代码位置:
    - go-zero/rest/server.go.Server().AddRoutes()

- 路由: 
    - go-zero/rest/router/patrouter.go:34
    
    
- 路由表定义位置: 
    - go-zero/rest/types.go:12


#### 3.3 engine.Start():



- start() 比较妙的地方: 
    - router 类型, 扩展的 interface. 
    
    
- go-zero/rest/engine.go:75


## ref:

### zrpc: 

- https://www.yuque.com/tal-tech/go-zero/rslrhx
- 这个是 grpc 的扩展, so, 那就好理解了. 









