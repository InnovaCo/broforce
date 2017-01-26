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

var version = "0.3.0"

func main() {
	cfgPath := kingpin.Flag("config", "Path to config.yml file.").Default("config.yml").String()
	logPath := kingpin.Flag("log", "Path for log file.").Default("broforce.log").String()
	show := kingpin.Flag("show", "Show all task names.").Bool()
	lvl := kingpin.Flag("log-level", "log level").Default("info").String()
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
	fmt.Println(allowTasks)

	f, err := os.OpenFile(*logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Errorf("Error opening file: %v", err)
		return
	}
	logger.New(f, *lvl)
	b := bus.New()
	c := config.New(*cfgPath, config.YAMLAdapter)
	for n, s := range tasks.GetPool() {
		if strings.Index(allowTasks, n) != -1 {
			go s.Run(b, c.Get(n))
		}
	}

	runtime.Goexit()
	defer f.Close()
}
