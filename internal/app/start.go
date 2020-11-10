package app

import "go-dfs/internal/pkg"

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
	// start the given server type
	if serverConfig.ServerType == "tracker" {
		tracker := NewTracker()
		tracker.Start(serverConfig)
	} else if serverConfig.ServerType == "storage" {
		storage := NewStorage()
		storage.Start(serverConfig)
	} else {
		panic("error server type")
	}
}
