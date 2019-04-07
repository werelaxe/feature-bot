package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

type Config struct {
	StoragePath string
	Token       string
	AdminId     int
}

func ParseConfig(configPath string) (*Config, error) {
	var config Config
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, errors.New("can not parse config")
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, errors.New("can not parse config")
	}
	return &config, nil
}
