package tasks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/mhanygin/go-gocd"

	"github.com/InnovaCo/broforce/bus"
	"github.com/InnovaCo/broforce/config"
)

func init() {
	registry("gocdSheduler", bus.Task(&gocdSheduler{}))
}

type gocdVars struct {
	Branch string `json:"variables[BRANCH]"`
	Sha    string `json:"variables[SHA]"`
}

type gocdSheduler struct {
	config   config.ConfigData
	login    string
	password string
	host     string
	times    int
	interval time.Duration
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

			ctx.Log.Debugf("%s: %s = %s",
				ref,
				fmt.Sprintf("pipelines.%s.ref", gitName),
				p.config.Search("pipelines", gitName, "ref"))

			if match, _ := regexp.MatchString(p.config.Search("pipelines", gitName, "ref"), ref); !match {
				ctx.Log.Debugf("%s not math %s", p.config.Search("pipelines", gitName, "ref"), ref)
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

			client := gocd.New(p.host, p.login, p.password)
			for i := 0; i < p.times; i++ {
				if err := client.SchedulePipeline(p.config.Search("pipelines", gitName, "pipeline"), d); err != nil {
					ctx.Log.Error(err)
					time.Sleep(p.interval * time.Second)
				} else {
					break
				}
			}
		}
	}
	return nil
}

func (p *gocdSheduler) Run(eventBus *bus.EventsBus, ctx bus.Context) error {
	p.config = ctx.Config

	p.host = p.config.GetString("host")
	p.times = ctx.Config.GetIntOr("times", 100)
	p.interval = time.Duration(ctx.Config.GetIntOr("interval", 10))

	if data, err := ioutil.ReadFile(p.config.GetString("access")); err == nil {
		cread := struct {
			Login    string `json:"login"`
			Password string `json:"password"`
		}{}
		if err := json.Unmarshal(data, &cread); err != nil {
			return err
		}
		p.login = cread.Login
		p.password = cread.Password
	} else {
		return err
	}
	eventBus.Subscribe(bus.GitlabHookEvent, bus.Context{Func: p.handler, Name: "GoCDShedulerHandler"})
	return nil
}
