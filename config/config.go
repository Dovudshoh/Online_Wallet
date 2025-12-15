package config

import (
	"encoding/json"
	"os"
)


type Config struct {
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
	DBName     string `json:"db_name"`
}

func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg := &Config{}
	err = json.NewDecoder(file).Decode(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
