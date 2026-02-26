package config

import (
	"os"
	"sync"

	"github.com/spf13/viper"
)

// 线程安全的配置管理模块，从 YAML 文件加载应用配置

type ParamsConfig struct {
	ApiKey    string `mapstructure:"APIKey"`
	Endpoint  string `mapstructure:"Endpoint"`
	Embedding string `mapstructure:"Embedding"`
	ChatModel string `mapstructure:"ChatModel"`
	Redis     struct {
		Host string `mapstructure:"Host"`
		Port int    `mapstructure:"Port"`
		DB   int    `mapstructure:"DB"`
	} `mapstructure:"Redis"`
}

var once sync.Once

var c *ParamsConfig

// Map 获取配置
func Map() *ParamsConfig {
	// 单例模式
	once.Do(func() {
		v := viper.New()
		file, err := os.Open("config/config.yml")
		if err != nil {
			panic(err)
		}
		v.SetConfigType("yaml")
		err = v.ReadConfig(file)
		if err != nil {
			panic(err)
		}

		err = v.Unmarshal(&c)
		if err != nil {
			panic(err)
		}
		return
	})
	return c
}
