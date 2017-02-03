package config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig_Init(t *testing.T) {
	data := `timeSensor:
  interval: 10

task1:
  param1:
    - value1
    - value2

slackSensor:
  param1: value1
  param2: value2
`
	tmpfile, err := ioutil.TempFile("/tmp", "manifest_")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(data)); err != nil {
		t.Error(err)
		t.Fail()
	}
	tmpfile.Close()
	config := defaultConfig{}
	assert.NoError(t, config.Init(tmpfile.Name()))
}

func TestDefaultConfig_Get(t *testing.T) {
	data := `timeSensor:
  interval: 10

task1:
  param1:
    - value1
    - value2

slackSensor:
  param1: value1
  param2: value2
`
	tmpfile, err := ioutil.TempFile("/tmp", "manifest_")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(data)); err != nil {
		t.Error(err)
		t.Fail()
	}
	tmpfile.Close()
	config := defaultConfig{}
	config.Init(tmpfile.Name())
	configData := config.Get("slackSensor")

	assert.Equal(t, configData.GetString("param1"), "value1")
	assert.Equal(t, configData.GetString("param2"), "value2")
	assert.Equal(t, configData.Exist("param3"), false)
}

func TestDefaultConfig_GetNotExist(t *testing.T) {
	data := `timeSensor:
  interval: 10

task1:
  param1:
    - value1
    - value2

slackSensor:
  param1: value1
  param2: value2
`
	tmpfile, err := ioutil.TempFile("/tmp", "manifest_")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(data)); err != nil {
		t.Error(err)
		t.Fail()
	}
	tmpfile.Close()
	config := defaultConfig{}
	config.Init(tmpfile.Name())
	configData := config.Get("testTask")

	assert.Equal(t, configData.Exist("param3"), false)
}
