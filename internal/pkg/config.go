package pkg

import (
	"fmt"

	"github.com/spf13/viper"
)

// DsfConfig dfs global config
var DsfConfig DsfConfigType

// DsfConfigType dfs配置
type DsfConfigType struct {
	// 对外提供服务的信息
	ServiceScheme string `mapstructure:"service_scheme"`
	ServiceIP     string `mapstructure:"service_ip"`
	ServicePort   string `mapstructure:"service_port"`
	ServiceType   string `mapstructure:"service_type"`
	BindPort      string `mapstructure:"bind_port"`
	DefaultLang   string `mapstructure:"default_lang"`
	// NodeType,may be tracker server or storage server
	Tracker struct {
		// 节点id，用于雪花算法生成唯一文件名称
		NodeID int64 `mapstructure:"node_id"`
		//启用临时文件功能
		EnableTempFile bool `mapstructure:"enable_temp_file"`
	} `mapstructure:"tracker"`
	Storage struct {
		// group
		Group string `mapstructure:"group"`
		// file size limit
		FileSizeLimit int64 `mapstructure:"file_size_limit"`
		// storagePath
		StoragePath string `mapstructure:"storage_path"`
		// trackerServers,can be one or more
		Trackers []string `mapstructure:"trackers"`
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
