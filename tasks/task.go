package tasks

import (
	"fmt"
	"strings"

	"github.com/InnovaCo/broforce/bus"
)

var tasksPool = make(map[string]bus.Task)

func registry(name string, task bus.Task) {
	if _, ok := tasksPool[name]; ok {
		panic(fmt.Errorf("Task %s already registry", name))
	} else {
		tasksPool[name] = task
	}
}

func GetPool() map[string]bus.Task {
	return tasksPool
}

func GetPoolString() string {
	keys := make([]string, 0, len(tasksPool))
	for k := range tasksPool {
		keys = append(keys, k)
	}

	return strings.Join(keys, ",")
}
