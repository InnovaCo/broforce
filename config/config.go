package config

import (
	"fmt"
	"sync"
)

var instance Config
var once sync.Once
var configAdapters = make(map[string]Config)

func registry(name string, cfg Config) {
	configAdapters[name] = cfg
}

func New(path string, adapter string) Config {
	once.Do(func() {
		instance = configAdapters[adapter]
		if instance == nil {
			return
		}
		if err := instance.Init(path); err != nil {
			fmt.Errorf("Error: %v", err)
			instance = nil
			return
		}
	})
	return instance
}

type ConfigData interface {
	String() string
	Exist(path string) bool
	Get(path string) ConfigData
	Search(hierarchy ...string) string
	GetString(path string) string
	GetStringOr(path string, defaultVal string) string
	GetFloat(path string) float64
	GetInt(path string) int
	GetIntOr(path string, defaultVal int) int
	GetArray(path string) []ConfigData
	GetArrayString(path string) []string
	GetMap(path string) map[string]ConfigData
}

type Config interface {
	Init(path string) error
	Get(name string) ConfigData
}
