# go-dfs
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
# 接口说明
- 上传
  - api: /upload
  - method: post
  - 参数:file
- 下载
  - api: /完整文件路径
  - method: get
- 删除
  - api: /delete
  - method: post
  - 参数: file_name(注意为完整路径)
  
# 项目工具
- gin，高效的golang web框架
- leveldb，基于golang的kv数据库
