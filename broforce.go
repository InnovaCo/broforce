package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/InnovaCo/broforce/bus"
	"github.com/InnovaCo/broforce/config"
	"github.com/InnovaCo/broforce/logger"
	"github.com/InnovaCo/broforce/tasks"
)

var version = "0.5.0"

func main() {
	cfgPath := kingpin.Flag("config", "Path to config.yml file.").Default("config.yml").String()
	show := kingpin.Flag("show", "Show all task names.").Bool()
	allow := kingpin.Flag("allow", "list of allowed tasks").Default(tasks.GetPoolString()).String()

	kingpin.Version(version)
	kingpin.Parse()

	if *show {
		fmt.Println("task names:")
		for n := range tasks.GetPool() {
			fmt.Println(" - ", n)
		}
		return
	}

	if _, err := os.Stat(*cfgPath); os.IsNotExist(err) {
		fmt.Errorf("%v", err)
		return
	}
	allowTasks := fmt.Sprintf(",%s,", *allow)
	c := config.New(*cfgPath, config.YAMLAdapter)
	if c == nil {
		fmt.Println("Error: config not create")
		return
	}
	logger.New(c.Get("logger"))
	b := bus.New()
	for n, s := range tasks.GetPool() {
		if strings.Index(allowTasks, fmt.Sprintf(",%s,", n)) != -1 {

			logger.Log.Debugf("Config for %s: %v", n, c.Get(n))

			go s.Run(bus.Context{Name: n, Config: c.Get(n), Log: logger.Logger4Handler(n, ""), Bus: b})
		}
	}

	runtime.Goexit()
}
