package config

import (
	"fmt"
	"github.com/Vladislav747/minio-project/pkg/logging"
	"github.com/ilyakaznacheev/cleanenv"
	"sync"
)

type Config struct {
	IsDebug bool `yaml:"is_debug"`
	Listen  struct {
		Type   string `yaml:"type" env-default:"port"`
		BindIP string `json:"bind_ip" env-default:"localhost"`
		Port   string `json:"port" env-default:"10002"`
	}
	MinIO struct {
		Endpoint  string `json:"endpoint" env-required:"true"`
		AccessKey string `json:"access_key" env-default:"minio"`
		SecretKey string `json:"secret_key" env-default:"minio123"`
	} `yaml:"minio"`
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		logger := logging.GetLogger()
		logger.Info("read application config")
		instance = &Config{}
		if err := cleanenv.ReadConfig("config.yml", instance); err != nil {
			fmt.Println(instance, "instance")
			help, _ := cleanenv.GetDescription(instance, nil)
			logger.Info(help)
			logger.Fatal(err)
		}
	})
	return instance
}
