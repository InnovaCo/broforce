package tasks

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/InnovaCo/broforce/bus"
)

func TestGetPool(t *testing.T) {
	tasksPool = make(map[string]bus.Task)
	registry("test", bus.Task(nil))

	assert.Equal(t, len(GetPool()), 1)
}
