package pkg

import (
	"fmt"

	"github.com/spf13/viper"
)

// DsfConfig dfs global config
var DsfConfig DsfConfigType

// DsfConfigType dfs配置
type DsfConfigType struct {
	ServerType  string `mapstructure:"server_type"`
	HTTPPort    string `mapstructure:"http_port"`
	DefaultLang string `mapstructure:"default_lang"`
	// NodeType,may be tracker server or storage server
	Tracker struct {
		NodeID int64 `mapstructure:"node_id"`
	} `mapstructure:"tracker"`
	Storage struct {
		// storage http scheme
		HTTPScheme string `mapstructure:"http_scheme"`
		// group
		Group string `mapstructure:"group"`
		// file size limit
		FileSizeLimit int64 `mapstructure:"file_size_limit"`
		// storagePath
		StoragePath string `mapstructure:"storage_path"`
		// trackerServers,can be one or more
		Tracker []string `mapstructure:"tracker"`
	} `mapstructure:"storage"`
}

// InitConfig 初始化配置
func InitConfig() {
	v := viper.New()
	v.AddConfigPath(".")
	v.AddConfigPath("./configs")
	v.AddConfigPath("../configs")
	v.SetConfigName("dfs")
	v.SetConfigType("yaml")
	err := v.ReadInConfig()
	if err != nil {
		panic("no config file")
	}
	err = v.Unmarshal(&DsfConfig)
	if err != nil {
		panic("parse config file error")
	}
	fmt.Printf("%+v", DsfConfig)
}
