

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

## 项目入口:


- 



## core: 






