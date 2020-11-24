package app

import (
	"go-dfs/internal/pkg"
)

// 可支持的动态变量注入
var (
	// ServerType 服务类型
	ServiceType = ""
	// 对外服务ip
	ServiceIP = ""
	// 对外服务port
	ServicePort = ""
	//运行端口
	BindPort = ""
	// 存储目录
	StoragePath = ""
)

// Start start the server
func Start(config *pkg.DsfConfigType) {
	// init config file
	serverConfig := pkg.DsfConfigType{}
	if config == nil {
		pkg.InitConfig()
		serverConfig = pkg.DsfConfig
	} else {
		serverConfig = *config
	}
	// reveive the injection var from the ldflags
	// dynamic set the
	if ServiceType != "" {
		serverConfig.ServiceType = ServiceType
	}
	if ServiceIP != "" {
		serverConfig.ServiceIP = ServiceIP
	}
	if ServicePort != "" {
		serverConfig.ServicePort = ServicePort
	}
	if BindPort != "" {
		serverConfig.BindPort = BindPort
	}
	if StoragePath != "" {
		serverConfig.Storage.StoragePath = StoragePath
	}
	// start the given server type
	if serverConfig.ServiceType == "tracker" {
		tracker := NewTracker()
		tracker.Start(serverConfig)
	} else if serverConfig.ServiceType == "storage" {
		storage := NewStorage()
		storage.Start(serverConfig)
	} else {
		panic("error server type")
	}
}
