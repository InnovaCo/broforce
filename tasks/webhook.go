package tasks

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/Jeffail/gabs"

	"github.com/InnovaCo/broforce/bus"
	"os"
	"strconv"
)

func init() {
	registry("hookSensor", bus.Task(&hookSensor{}))
}

const (
	maxRetry       = 10
	envWebhookPort = "BROFORSE_WEBHOOK_PORT"
	defaultDelay   = 10
	defaultPort    = 8080
)

type hookSensor struct {
	gitParams  map[string]string
	jiraParams map[string]string
	ctx        *bus.Context
}

func (p hookSensor) selector(body []byte) (string, error) {
	g, err := gabs.ParseJSON(body)
	if err != nil {
		return bus.UnknownEvent, err
	}

	val, ok := g.Search("repository", "url").Data().(string)
	if !ok {
		return bus.UnknownEvent, fmt.Errorf("Key %s not found", "repository.url")
	}

	p.ctx.Log.Debugf("Repo %v", val)
	p.ctx.Log.Debugf("Hook: %s", g.String())

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
	p.ctx.Log.Debug(r.Header, r.ContentLength)
	defer r.Body.Close()

	if strings.Compare(p.gitParams["AuthKeyValue"], r.FormValue(p.gitParams["AuthKeyName"])) != 0 {
		p.ctx.Log.Debugf("not valid %v: \"%v\"!=\"%v\"",
			p.gitParams["AuthKeyName"],
			p.gitParams["AuthKeyValue"],
			r.FormValue("api-key"))
		return
	}

	if body, err := ioutil.ReadAll(r.Body); err != nil {
		p.ctx.Log.Error(err)
	} else {
		if gitType, err := p.selector(body); err != nil {
			p.ctx.Log.Error(err)
		} else {
			uuid := bus.NewUUID()

			p.ctx.Log.Debugf("Push: %s", uuid)

			if err := p.ctx.Bus.Publish(bus.Event{
				Trace:   uuid,
				Subject: gitType,
				Coding:  bus.JsonCoding,
				Data:    body}); err != nil {
				p.ctx.Log.Error(err)
			}
		}
	}
	return
}

func (p *hookSensor) jira(w http.ResponseWriter, r *http.Request) {
	p.ctx.Log.Debug(r.Header, r.ContentLength)
	defer r.Body.Close()

	if strings.Compare(p.jiraParams["AuthKeyValue"], r.FormValue(p.jiraParams["AuthKeyName"])) != 0 {
		p.ctx.Log.Debugf("not valid %v: \"%v\"!=\"%v\"",
			p.jiraParams["AuthKeyName"],
			p.jiraParams["AuthKeyValue"],
			r.FormValue("api-key"))
		return
	}

	if body, err := ioutil.ReadAll(r.Body); err != nil {
		p.ctx.Log.Error(err)
	} else {
		uuid := bus.NewUUID()

		p.ctx.Log.Debugf("Push: %s", uuid)

		if err := p.ctx.Bus.Publish(bus.Event{
			Trace:   uuid,
			Subject: bus.JiraHookEvent,
			Coding:  bus.JsonCoding,
			Data:    body}); err != nil {
			p.ctx.Log.Error(err)
		}
	}
	return
}

func (p *hookSensor) Run(ctx bus.Context) error {
	p.ctx = &ctx

	if p.ctx.Config.Exist("git") {
		p.ctx.Log.Debugf("add git handler with params: %v", p.ctx.Config.GetMap("git"))

		p.gitParams = make(map[string]string)
		p.gitParams["AuthKeyName"] = p.ctx.Config.GetStringOr("git.auth-key-name", "")
		p.gitParams["AuthKeyValue"] = p.ctx.Config.GetStringOr("git.auth-key-value", "")
		http.HandleFunc(p.ctx.Config.GetStringOr("git.url", "/git"), p.git)
	}

	if p.ctx.Config.Exist("jira") {
		p.ctx.Log.Debugf("add jira handler with params: %v", p.ctx.Config.GetMap("jira"))

		p.jiraParams = make(map[string]string)
		p.jiraParams["AuthKeyName"] = p.ctx.Config.GetStringOr("jira.auth-key-name", "")
		p.jiraParams["AuthKeyValue"] = p.ctx.Config.GetStringOr("jira.auth-key-value", "")
		http.HandleFunc(p.ctx.Config.GetStringOr("jira.url", "/jira"), p.jira)
	}

	p.ctx.Log.Debug("Run")

	delay := time.Duration(p.ctx.Config.GetIntOr("delay", defaultDelay))
	port, err := strconv.Atoi(os.Getenv(envWebhookPort))
	if err != nil {
		p.ctx.Log.Error(err)
		port = p.ctx.Config.GetIntOr("port", defaultPort)
	}

	p.ctx.Log.Debugf("PORT: %d", port)

	i := 0
	for {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			p.ctx.Log.Debug(err)

			time.Sleep(delay * time.Second)
			i++
			if i >= maxRetry {
				i = 0
				delay = 2 * delay
				p.ctx.Log.Infof("new delay %v", delay)
			}
		} else {
			break
		}
	}

	p.ctx.Log.Debug("Complete")

	return nil
}
