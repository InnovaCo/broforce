package tasks

import (
	"math/rand"
	"regexp"

	"github.com/andygrunwald/go-jira"
	"github.com/valyala/fasttemplate"

	"github.com/InnovaCo/broforce/bus"
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

func (p *jiraResolver) Run(eventBus *bus.EventsBus, ctx bus.Context) error {
	p.bus = eventBus
	var err error
	if p.reg, err = regexp.Compile(ctx.Config.GetStringOr("input-template", "")); err != nil {
		return err
	}

	p.output = fasttemplate.New(ctx.Config.GetStringOr("output-template", ""), "{{", "}}")

	for _, t := range ctx.Config.GetArray("unknown-template") {
		p.unknown = append(p.unknown, fasttemplate.New(t.String(), "{{", "}}"))
	}

	p.host = ctx.Config.GetStringOr("jira-host", "")
	p.user = ctx.Config.GetStringOr("jira-user", "")
	p.password = ctx.Config.GetStringOr("jira-password", "")

	p.bus.Subscribe(bus.SlackMsgEvent, bus.Context{Func: p.handler, Name: "JiraResolverHandler"})

	return nil
}

func (p *jiraResolver) handler(e bus.Event, ctx bus.Context) error {
	msg := slackMessage{}
	if err := bus.Encoder(e.Data, &msg, e.Coding); err != nil {
		return err
	}

	event := bus.Event{
		Trace:   e.Trace,
		Subject: bus.SlackPostEvent,
		Coding:  bus.JsonCoding}

	jiraClient, err := jira.NewClient(nil, p.host)
	if err != nil {
		return err
	}

	if res, err := jiraClient.Authentication.AcquireSessionCookie(p.user, p.password); err != nil || res == false {
		return err
	}

	for _, s := range p.reg.FindAllString(msg.Text, -1) {
		ctx.Log.Debug("Get issue:", s)
		issue, _, err := jiraClient.Issue.Get(s)
		if err != nil {

			ctx.Log.Error(err)

			if len(p.unknown) > 0 {
				if err := bus.Coder(&event, slackMessage{
					Type:    msg.Type,
					Channel: msg.Channel,
					Text: p.unknown[rand.Intn(len(p.unknown)-1)].ExecuteString(map[string]interface{}{
						"key": s})}); err == nil {
					if err := p.bus.Publish(event); err != nil {
						ctx.Log.Error(err)
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
				ctx.Log.Error()
			}
		} else {
			return err
		}
	}

	return nil
}
