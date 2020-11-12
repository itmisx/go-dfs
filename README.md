# easy-dfs
简单的分布式文件系统

[![LICENSE](https://raw.githubusercontent.com/smally84/go-dfs/4bffdfa2b020b98ccd8d618a53ac7c0294d26786/assets/mit.svg)](https://github.com/syndtr/goleveldb)
[![LANGUAGE](https://raw.githubusercontent.com/smally84/go-dfs/28deed0c494ef141109a5e141c54b752c20923b0/assets/language.svg)]()
[![DB](https://raw.githubusercontent.com/smally84/go-dfs/6851b5a0b570278c52135f77e812810278c2b898/assets/leveldb.svg)]()
[![gopsutil](https://raw.githubusercontent.com/smally84/go-dfs/28deed0c494ef141109a5e141c54b752c20923b0/assets/gopsutil.svg)]()

# 项目说明
本项目参考fastdfs逻辑进行简单实现，主要功能包括：
- 文件上传
- 文件删除
- 文件下载
- 文件多副本同步保存和删除
- tracker自动容量均衡到不同的存储组

# 使用说明
>特别说明，无论是tracker还是storage，使用的都是同个编译输出文件，仅仅配置不同而已
## 一、源码安装
- 1.clone源代码,编译出二进制文件
```
cd cmd
go build main.go -o dfs
./dfs
```
- 2.配置文件
将 configs/dsf.yml放到dfs可执行文件目录
```#服务类型tracker or storage
server_type: "storage"
#http_port
http_port: 9000
#默认语言
default_lang: zh_cn
#跟踪服务器的配置
tracker:
  node_id: 1
# 存储服务器的配置    
storage:
  #http_scheme
  http_scheme: http
  #存储服务所属的group
  group: group1 
  #文件大小限制,单位字节
  file_size_limit: 100000
  #存储目录
  storage_path: ./
  #跟踪服务器，可以有多个
  tracker: 
    - http://127.0.0.1:9000
```
服务的类型：用server_type来定义。
最小系统，要配置一个tracker，一个storage
## docker安装
参考docker-compose.yml配置
```
version: "3.7"
services:
  #tracker
  tracker:
    image: golang:1.15
    container_name: "dfs_tracker"
    expose:
      - 9000
    volumes:
      - ./go-dfs/tracker:/app  
    working_dir: /app
    command: /app/tracker_app
    restart: "always"
    networks:
      soa: 
        ipv4_address: 172.20.0.2
    logging:
      options:
        max-size: "10M"
        max-file: "5"
  storage_1:
    image: golang:1.15
    container_name: "dfs_storage_1"
    expose:
      - 9000
    volumes:
      - ./go-dfs/storage-1:/app
    working_dir: /app
    command: /app/storage_app
    restart: "always"
    networks:
      soa: 
        ipv4_address: 172.20.0.3
    logging:
      options:
        max-size: "10M"
        max-file: "5"
  storage_2:
    image: golang:1.15
    container_name: "dfs_storage_2"
    expose:
      - 9000
    volumes:
      - ./go-dfs/storage-2:/app
    working_dir: /app
    command: /app/storage_app
    restart: "always"
    networks:
      soa: 
        ipv4_address: 172.20.0.4
    logging:
      options:
        max-size: "10M"
        max-file: "5"
# 网络配置
networks:
  dfs:
    name: dfs
    ipam:
      driver: default
      config:
        -
          subnet: "172.20.0.0/16"
```
目录结构
```
-go-dfs
 |_ _tracker
 |      |_ _ configs
 |      |       |_ _ dfs.yml
 |      |_ _ tracker
 |_ _storage-1
 |      |_ _ configs
 |      |       |_ _ dfs.yml
 |      |_ _ storage
 |_ _storage-2
        |_ _ configs
        |       |_ _ dfs.yml
        |_ _ storage
```
dfs.yml的配置请参考configs/dfs.yml
其中：serverType要配置对应的服务类型，跟踪服务器为`tracker`,存储服务器为`storage`
另外存储服务器时，要配置tracker服务器的host地址

# 接口说明
统一返回格式：
  ```
  {
    "code": 0,
    "data": {},
    "msg": ""
  }
  ```
  code为0表示操作成功，非0表示有错误，msg为错误信息
- 上传
  - api: /upload
  - method: post
  - 参数:file
  返回结果示例：
  ```
  {
    "code": 0,
    "data": {
        "url": "/group1/2020/11/12/7/1326785062468919296.png"
    },
    "msg": ""
  }
  ```
- 下载
  - api: /完整文件路径
  - method: get
- 删除
  - api: /delete
  - method: post
  - 参数: file(值为完整的文件路径)
  返回结果示例:
  ```
  {
    "code": 0,
    "msg": ""
  }
  ```
# 测试用例  
参考目录: easy-dfs/internal/app/dfs_test.go
```go
// 启动tracker
config1 := pkg.DsfConfigType{
  ServerType:  "tracker",
  HTTPPort:    "9000",
  DefaultLang: "zh_cn",
}
config1.Tracker.NodeID = 1
go Start(&config1)
// 启动storage
config2 := pkg.DsfConfigType{
  ServerType:  "storage",
  HTTPPort:    "9001",
  DefaultLang: "zh_cn",
}
config2.Storage.HTTPScheme = "http"
config2.Storage.Group = "group1"
config2.Storage.StoragePath = "./dfs/1"
config2.Storage.Tracker = []string{"http://127.0.0.1:9000"}
go Start(&config2)

// 启动storage
config3 := pkg.DsfConfigType{
  ServerType:  "storage",
  HTTPPort:    "9002",
  DefaultLang: "zh_cn",
}
config3.Storage.HTTPScheme = "http"
config3.Storage.Group = "group1"
config3.Storage.StoragePath = "./dfs/2"
config3.Storage.Tracker = []string{"http://127.0.0.1:9000"}
go Start(&config3)
```
# 项目工具
- gin，高效的golang web框架
- leveldb，基于golang的kv数据库

# License
MIT
