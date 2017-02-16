package config

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/ghodss/yaml"
)

func init() {
	registry(YAMLAdapter, Config(&defaultConfig{}))
}

type defaultConfig struct {
	data gabs.Container
	path string
}

func (p *defaultConfig) Init(path string) error {
	p.path = path
	data, err := ioutil.ReadFile(p.path)
	if err != nil {
		return fmt.Errorf("file `%s` not found: %v", p.path, err)
	}
	if jsonData, err := yaml.YAMLToJSON(data); err != nil {
		return fmt.Errorf("Error on parse file `%s`: %v!", p.path, err)
	} else {
		g, _ := gabs.ParseJSON(jsonData)
		p.data = *g
	}
	return nil
}

func (p *defaultConfig) Get(name string) ConfigData {
	if c := p.data.Search(name); c != nil {
		return ConfigData(&defaultConfigData{data: c})
	} else {
		return ConfigData(&defaultConfigData{data: &gabs.Container{}})
	}
}

type defaultConfigData struct {
	data *gabs.Container
}

func (p *defaultConfigData) String() string {
	return p.data.String()
}

func (p *defaultConfigData) Get(name string) ConfigData {
	if c := p.data.Search(name); c != nil {
		return ConfigData(&defaultConfigData{data: c})
	} else {
		return ConfigData(&defaultConfigData{data: &gabs.Container{}})
	}
}

func (p *defaultConfigData) Exist(path string) bool {
	return p.data.ExistsP(path)
}

func (p *defaultConfigData) GetString(path string) string {
	if len(strings.TrimSpace(path)) == 0 {
		return p.data.Data().(string)
	} else {
		return fmt.Sprintf("%v", p.data.Path(path).Data())
	}
}

func (p *defaultConfigData) Search(hierarchy ...string) string {
	return fmt.Sprintf("%v", p.data.Search(hierarchy...).Data())
}

func (p *defaultConfigData) GetStringOr(path string, defaultVal string) string {
	if p.data.ExistsP(path) {
		return p.GetString(path)
	} else {
		return defaultVal
	}
}

func (p *defaultConfigData) GetFloat(path string) float64 {
	if f, err := strconv.ParseFloat(p.GetString(path), 64); err != nil {
		fmt.Errorf("Error get float: %v", err)
		return 0
	} else {
		return f
	}
}

func (p *defaultConfigData) GetInt(path string) int {
	if i, err := strconv.Atoi(p.GetString(path)); err != nil {
		fmt.Errorf("Error get int: %v", err)
		return 0
	} else {
		return i
	}
}

func (p *defaultConfigData) GetIntOr(path string, defaultVal int) int {
	if p.data.ExistsP(path) {
		return p.GetInt(path)
	} else {
		return defaultVal
	}
}

func (p *defaultConfigData) GetBool(path string) bool {
	return strings.ToLower(p.GetString(path)) == "true"
}

func (p *defaultConfigData) GetArray(path string) []ConfigData {
	out := make([]ConfigData, 0)
	var data *gabs.Container

	if len(strings.TrimSpace(path)) == 0 {
		data = p.data
	} else {
		data = p.data.Path(path)
	}
	arr, err := data.Children()
	if err != nil {
		fmt.Errorf("Error get array `%v` from: %v", path, p.data.Path(path).Data())
		return out
	}
	for _, v := range arr {
		out = append(out, ConfigData(&defaultConfigData{data: v}))
	}
	return out
}

func (p *defaultConfigData) GetArrayString(path string) []string {
	out := make([]string, 0)
	var data *gabs.Container

	if len(strings.TrimSpace(path)) == 0 {
		data = p.data
	} else {
		data = p.data.Path(path)
	}
	arr, err := data.Children()
	if err != nil {
		fmt.Errorf("Error get array `%v` from: %v", path, p.data.Path(path).Data())
		return out
	}
	for _, v := range arr {
		out = append(out, v.Data().(string))
	}
	return out
}

func (p *defaultConfigData) GetMap(path string) map[string]ConfigData {
	out := make(map[string]ConfigData)
	var data *gabs.Container

	if len(strings.TrimSpace(path)) == 0 {
		data = p.data
	} else {
		data = p.data.Path(path)
	}
	mmap, err := data.ChildrenMap()
	if err != nil {
		fmt.Errorf("Error get map '%v' from: %v. Error: %s", path, p.data.Path(path).Data(), err)
		return out
	}

	for k, v := range mmap {
		out[k] = ConfigData(&defaultConfigData{data: v})
	}
	return out
}
