# 进入应用入口
cd cmd
# 编译出linux可执行文件
# -w 禁止程序gdb调试，防止代码泄露
# -s 禁止调试信息输出，如panic异常信息等，避免保留敏感信息
SET GOOS=linux SET GOARCH=amd64 go build -ldflags "-w -s" -o dfs