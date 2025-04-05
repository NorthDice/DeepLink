package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

type Config struct {
	Env      string   `yaml:"env"`
	Database Database `yaml:"database"`
	Kafka    KafkaConfig
}

type Database struct {
	MongoDB MongoConfig `yaml:"mongodb"`
}
type MongoConfig struct {
	Uri string `yaml:"uri" env-required:"true"`
}
type KafkaConfig struct {
	Brokers string `yaml:"brokers" env-required:"true"`
}

// fetchConfigPath fetches config path from command line flag or environment variable.
// Priority: flag > env > default.
// Default value is empty string.

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config file path is empty")
	}

	return MustLoadByPath(path)
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}

func MustLoadByPath(path string) *Config {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exist" + path)
	}

	var config Config

	if err := cleanenv.ReadConfig(path, &config); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &config
}
