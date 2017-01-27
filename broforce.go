package main

import (
	"fmt"
	"runtime"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/InnovaCo/broforce/bus"
	"github.com/InnovaCo/broforce/config"
	"github.com/InnovaCo/broforce/logger"
	"github.com/InnovaCo/broforce/tasks"
)

var version = "0.3.0"

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

	allowTasks := fmt.Sprintf(",%s,", *allow)
	c := config.New(*cfgPath, config.YAMLAdapter)
	logger.New(c.Get("logger"))
	b := bus.New()
	for n, s := range tasks.GetPool() {
		if strings.Index(allowTasks, n) != -1 {
			go s.Run(b, c.Get(n))
		}
	}

	runtime.Goexit()
}
