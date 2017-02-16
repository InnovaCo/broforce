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

func TestDefaultConfigData_GetMap(t *testing.T) {
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
	configData := config.Get("slackSensor").GetMap("   ")

	v, ok := configData["param1"]

	assert.Equal(t, len(configData), 2)
	assert.Equal(t, ok, true)
	assert.Equal(t, v.GetString(""), "value1")
}

func TestDefaultConfigData_GetArrayString(t *testing.T) {
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
	configData := config.Get("task1").GetArrayString("param1")

	assert.Equal(t, len(configData), 2)
	assert.Equal(t, configData, []string{"value1", "value2"})
}

func TestDefaultConfigData_GetArray(t *testing.T) {
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
	configData := config.Get("task1").GetArray("param1")

	assert.Equal(t, len(configData), 2)
	assert.Equal(t, configData[0].GetString(""), "value1")
}

func TestDefaultConfigData_Get(t *testing.T) {
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
	configData := config.Get("task1").Get("param1")
	assert.Equal(t, configData.GetArrayString(""), []string{"value1", "value2"})
}
