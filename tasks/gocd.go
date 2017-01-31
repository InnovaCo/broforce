package tasks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/Jeffail/gabs"

	"github.com/InnovaCo/broforce/bus"
	"github.com/InnovaCo/broforce/config"
	"github.com/InnovaCo/broforce/logger"
)

func init() {
	registry("gocdSheduler", bus.Task(&gocdSheduler{}))
}

type goCdCredents struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type gocdVars struct {
	Branch string `json:"variables[BRANCH]"`
	Sha    string `json:"variables[SHA]"`
}

type gocdSheduler struct {
	config   config.ConfigData
	credents goCdCredents
}

func (p *gocdSheduler) handler(e bus.Event, ctx bus.Context) error {
	if e.Coding != bus.JsonCoding {
		return nil
	}
	g, err := gabs.ParseJSON(e.Data)
	if err != nil {
		return err
	}
	git, ok := g.Path("repository.git_ssh_url").Data().(string)
	if !ok {
		return fmt.Errorf("Key %s not found", "repository.git_ssh_url")
	}
	ref, ok := g.Path("ref").Data().(string)
	if !ok {
		return fmt.Errorf("Key %s not found", "ref")
	}
	for gitName := range p.config.GetMap("pipelines") {
		if strings.Compare(gitName, git) == 0 {
			if match, _ := regexp.MatchString(p.config.GetString(fmt.Sprintf("pipelines.%s.ref", gitName)), ref); !match {
				ctx.Log.Debugf("%s not math %s", p.config.GetString(fmt.Sprintf("pipelines.%s.ref", gitName)), ref)
				return nil
			}
			if before, ok := g.Path("before").Data().(string); ok && strings.Compare(before, defaultSHA) == 0 {
				ctx.Log.Debugf("before == %s", g.Path("before").Data().(string))
				return nil
			}
			v := gocdVars{}
			if v.Sha, ok = g.Path("ref").Data().(string); !ok {
				return fmt.Errorf("Key %s not found", "body.ref")
			}
			s := strings.Split(ref, "/")
			v.Branch = s[len(s)-1]
			d, _ := json.Marshal(v)
			resp, err := p.goCdRequest("POST",
				fmt.Sprintf("%s/go/api/pipelines/%s/schedule",
					p.config.GetString("host"),
					p.config.GetString(fmt.Sprintf("pipelines.%s.pipeline", gitName))),
				string(d),
				map[string]string{"Confirm": "true"})

			switch true {
			case err != nil:
				return err
			case resp.StatusCode != http.StatusOK:
				return fmt.Errorf("Operation error: %s", resp.Status)
			default:
				break
			}
		}
	}
	return nil
}

func (p *gocdSheduler) Run(eventBus *bus.EventsBus, ctx bus.Context) error {
	p.config = ctx.Config

	if data, err := ioutil.ReadFile(p.config.GetString("access")); err == nil {
		p.credents = goCdCredents{}
		json.Unmarshal(data, &p.credents)
	} else {
		return err
	}

	eventBus.Subscribe(bus.GitlabHookEvent, bus.Context{Func: p.handler, Name: "GoCDShedulerHandler"})
	return nil
}

func (p gocdSheduler) goCdRequest(method string, resource string, body string, headers map[string]string) (*http.Response, error) {
	req, _ := http.NewRequest(method, resource, bytes.NewReader([]byte(body)))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", "application/json")

	req.SetBasicAuth(p.credents.Login, p.credents.Password)

	logger.Log.Debugf(" --> %s %s:\n%s\n%s\n\n", method, resource, req.Header, body)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	} else {
		logger.Log.Debugf("<-- %s\n", resp.Status)
	}
	return resp, nil
}
