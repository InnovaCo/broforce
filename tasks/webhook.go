package tasks

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/Jeffail/gabs"

	"github.com/InnovaCo/broforce/bus"
	"github.com/InnovaCo/broforce/config"
	"github.com/InnovaCo/broforce/logger"
)

func init() {
	registry("hookSensor", bus.Task(&hookSensor{}))
}

const (
	maxRepry = 10
)

type hookSensor struct {
	gitParams map[string]string
	bus       *bus.EventsBus
}

func (p hookSensor) selector(body []byte) (string, error) {
	g, err := gabs.ParseJSON(body)
	if err != nil {
		return bus.UnknownEvent, err
	}

	logger.Log.Info(string(body))

	val, ok := g.Search("repository", "url").Data().(string)
	if !ok {
		return bus.UnknownEvent, fmt.Errorf("Key %s not found", "repository.url")
	}

	logger.Log.Debugf("Repo %v", val)

	switch true {
	case strings.Index(val, "gitlab.") != -1:
		return bus.GitlabHookEvent, nil
	case strings.Index(val, "github.") != -1:
		return bus.GithubHookEvent, nil
	default:
		return bus.UnknownEvent, fmt.Errorf("detect %s", bus.UnknownEvent)
	}
}

func (p *hookSensor) git(w http.ResponseWriter, r *http.Request) {
	logger.Log.Debug(r.Header, r.ContentLength)

	if strings.Compare(p.gitParams["AuthKeyValue"], r.FormValue(p.gitParams["AuthKeyName"])) != 0 {
		logger.Log.Debugf("not valid %v: \"%v\"!=\"%v\"",
			p.gitParams["AuthKeyName"],
			p.gitParams["AuthKeyValue"],
			r.FormValue("api-key"))
		return
	}
	defer r.Body.Close()

	if body, err := ioutil.ReadAll(r.Body); err != nil {
		logger.Log.Error(err)
	} else {
		if g, err := p.selector(body); err != nil {
			logger.Log.Error(err)
		} else {
			if err := p.bus.Publish(bus.Event{
				Trace:   bus.NewUUID(),
				Subject: g,
				Coding:  bus.JsonCoding,
				Data:    body}); err != nil {
				logger.Log.Error(err)
			}
		}
	}
	return
}

func (p *hookSensor) Run(eventBus *bus.EventsBus, cfg config.ConfigData) error {
	p.bus = eventBus

	if cfg.Exist("git") {
		logger.Log.Debugf("add git handler with params: %v", cfg.GetMap("git"))
		p.gitParams = make(map[string]string)

		p.gitParams["AuthKeyName"] = cfg.GetStringOr("git.auth-key-name", "")
		p.gitParams["AuthKeyValue"] = cfg.GetStringOr("git.auth-key-value", "")
		http.HandleFunc(cfg.GetStringOr("git.url", "/git"), p.git)
	}

	logger.Log.Debug("Run")
	var err error
	i := 0
	delay := time.Duration(cfg.GetIntOr("delay", 10))
	for {
		if err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.GetIntOr("port", 8080)), nil); err != nil {
			logger.Log.Debug(err)
			time.Sleep(delay * time.Second)
			i++
			if i >= maxRepry {
				i = 0
				delay = 2 * delay
				logger.Log.Infof("new delay %v", delay)
			}
		} else {
			break
		}
	}
	logger.Log.Debug("Complete")
	return err
}
