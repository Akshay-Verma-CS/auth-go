package configuration

import (
	"log"
	"os"
	"sync"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Cache struct {
		Redis struct {
			Url            string
			Address        string
			Username       string
			Password       string
			DB             int
			Expiration     int
			ClusterAddress []string
		}
	}
	Database struct {
		MySQL struct {
			Host     string
			Username string
			Password string
			DB       string
		}
	}
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}
		err := LoadConfig("config.yaml", instance)
		if err != nil {
			log.Fatalf("Error loading config: %v", err)
		}
	})
	return instance
}

func LoadConfig(filename string, config *Config) error {
	configFile, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(configFile, config)
	if err != nil {
		return err
	}
	return nil
}
