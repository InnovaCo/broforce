package tasks

import (
	"math/rand"
	"regexp"

	"github.com/andygrunwald/go-jira"
	"github.com/valyala/fasttemplate"

	"github.com/InnovaCo/broforce/bus"
	"github.com/InnovaCo/broforce/config"
	"github.com/InnovaCo/broforce/logger"
)

func init() {
	registry("jiraResolver", bus.Task(&jiraResolver{}))
}

type jiraResolver struct {
	bus      *bus.EventsBus
	host     string
	user     string
	password string
	reg      *regexp.Regexp
	output   *fasttemplate.Template
	unknown  []*fasttemplate.Template
}

func (p *jiraResolver) Run(eventBus *bus.EventsBus, cfg config.ConfigData) error {
	logger.Log.Debug(cfg.String())

	p.bus = eventBus
	var err error
	if p.reg, err = regexp.Compile(cfg.GetStringOr("input-template", "")); err != nil {
		return err
	}

	p.output = fasttemplate.New(cfg.GetStringOr("output-template", ""), "{{", "}}")

	for _, t := range cfg.GetArray("unknown-template") {
		p.unknown = append(p.unknown, fasttemplate.New(t.String(), "{{", "}}"))
	}

	p.host = cfg.GetStringOr("jira-host", "")
	p.user = cfg.GetStringOr("jira-user", "")
	p.password = cfg.GetStringOr("jira-password", "")

	p.bus.Subscribe(bus.SlackMsgEvent, p.handler)

	return nil
}

func (p *jiraResolver) handler(e bus.Event) error {
	msg := slackMessage{}
	if err := bus.Encoder(e.Data, &msg, e.Coding); err != nil {
		return err
	}

	event := bus.Event{Subject: bus.SlackPostEvent, Coding: bus.JsonCoding}

	jiraClient, err := jira.NewClient(nil, p.host)
	if err != nil {
		return err
	}

	if res, err := jiraClient.Authentication.AcquireSessionCookie(p.user, p.password); err != nil || res == false {
		return err
	}

	for _, s := range p.reg.FindAllString(msg.Text, -1) {
		logger.Log.Debug("Get issue:", s)
		issue, _, err := jiraClient.Issue.Get(s)
		if err != nil {

			logger.Log.Error(err)

			if len(p.unknown) > 0 {
				if err := bus.Coder(&event, slackMessage{
					Type:    msg.Type,
					Channel: msg.Channel,
					Text: p.unknown[rand.Intn(len(p.unknown)-1)].ExecuteString(map[string]interface{}{
						"key": s})}); err == nil {
					if err := p.bus.Publish(event); err != nil {
						logger.Log.Error(err)
					}
				} else {
					return err
				}
			}
			continue
		}

		if err := bus.Coder(&event, slackMessage{
			Type:    msg.Type,
			Channel: msg.Channel,
			Text: p.output.ExecuteString(map[string]interface{}{
				"key":     issue.Key,
				"url":     issue.Self,
				"summary": issue.Fields.Summary,
				"status":  issue.Fields.Status.Name})}); err == nil {
			if err := p.bus.Publish(event); err != nil {
				logger.Log.Error()
			}
		} else {
			return err
		}
	}

	return nil
}
