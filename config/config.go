package config

import (
	"sync"
)

var instance Config
var once sync.Once
var configAdapters = make(map[string]Config)

func New(path string, adapter string) Config {
	once.Do(func() {
		if _, ok := configAdapters[adapter]; !ok {
			adapter = YAMLAdapter
		}
		instance = configAdapters[adapter]
		if err := instance.Init(path); err != nil {
			panic(err)
		}
	})
	return instance
}

type ConfigData interface {
	String() string
	Exist(path string) bool
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
