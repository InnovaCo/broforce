package tasks

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/andygrunwald/go-jira"
	"github.com/nlopes/slack"
	"github.com/valyala/fasttemplate"

	"github.com/InnovaCo/broforce/bus"
)

func init() {
	registry("jiraResolver", bus.Task(&jiraResolver{}))
	registry("jiraCommenter", bus.Task(&jiraCommenter{}))
}

func createLink(issue *jira.Issue) string {
	return strings.Replace(strings.Replace(issue.Self, "/rest/api/2/issue", "/browse", -1), issue.ID, issue.Key, -1)
}

type jiraResolver struct {
	host     string
	user     string
	password string
	reg      *regexp.Regexp
	output   *fasttemplate.Template
	unknown  []*fasttemplate.Template
}

func (p *jiraResolver) Run(ctx bus.Context) error {
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

	ctx.Bus.Subscribe(bus.SlackMsgEvent, bus.Context{
		Func:   p.handler,
		Name:   "JiraResolverHandler",
		Bus:    ctx.Bus,
		Config: ctx.Config})

	return nil
}

func (p *jiraResolver) handler(e bus.Event, ctx bus.Context) error {
	msg := slackMessage{}
	if err := e.Unmarshal(&msg); err != nil {
		return err
	}

	jiraClient, err := jira.NewClient(nil, p.host)
	if err != nil {
		return err
	}
	if res, err := jiraClient.Authentication.AcquireSessionCookie(p.user, p.password); err != nil || res == false {
		return err
	}

	event := bus.NewEvent(e.Trace, bus.SlackPostEvent, bus.JsonCoding)

	set := make(map[string]bool)
	for _, s := range p.reg.FindAllString(msg.Text, -1) {
		if _, found := set[s]; found {
			continue
		} else {
			set[s] = true
		}

		ctx.Log.Debug("Get issue:", s)
		issue, _, err := jiraClient.Issue.Get(s, nil)
		if err != nil {

			ctx.Log.Error(err)

			if len(p.unknown) > 0 {
				if err := event.Marshal(slackMessage{
					Type:    msg.Type,
					Channel: msg.Channel,
					Text: p.unknown[rand.Intn(len(p.unknown)-1)].ExecuteString(map[string]interface{}{
						"key": s})}); err != nil {
					return err
				}
				if err := ctx.Bus.Publish(*event); err != nil {
					return err
				}
			}
			continue
		}

		if err := event.Marshal(slackMessage{
			Type:    msg.Type,
			Channel: msg.Channel,
			Text: p.output.ExecuteString(map[string]interface{}{
				"key":     issue.Key,
				"url":     createLink(issue),
				"summary": issue.Fields.Summary,
				"status":  issue.Fields.Status.Name})}); err != nil {
			return err
		}
		if err := ctx.Bus.Publish(*event); err != nil {
			return err
		}
	}
	return nil
}

type jiraCommenter struct {
	output  *fasttemplate.Template
	channel string
}

func (p *jiraCommenter) Run(ctx bus.Context) error {
	p.channel = ctx.Config.GetStringOr("channel", "")
	p.output = fasttemplate.New(ctx.Config.GetStringOr("output-template", ""), "{{", "}}")
	ctx.Bus.Subscribe(bus.JiraHookEvent, bus.Context{
		Func:   p.handler,
		Name:   "JiraCommentHandler",
		Bus:    ctx.Bus,
		Config: ctx.Config})

	return nil
}

func (p *jiraCommenter) handler(e bus.Event, ctx bus.Context) error {
	g, err := gabs.ParseJSON(e.Data)
	if err != nil {
		return err
	}

	if !g.ExistsP("comment") {
		return fmt.Errorf("no comment tag")
	}
	issue := jira.Issue{}
	if err := json.Unmarshal(g.Path("issue").Bytes(), &issue); err != nil {
		return err
	}
	comment := jira.Comment{}
	if err := json.Unmarshal(g.Path("comment").Bytes(), &comment); err != nil {
		return err
	}

	msg := slackMessage{
		Text: p.output.ExecuteString(map[string]interface{}{
			"key":     issue.Key,
			"url":     createLink(&issue),
			"summary": issue.Fields.Summary,
			"status":  issue.Fields.Status.Name}),
		Channel: p.channel,
		Attachments: []slack.Attachment{
			slack.Attachment{
				Color:      "#008000",
				Text:       "Комментарий",
				MarkdownIn: []string{"title", "fields", "text"},
				Fields: []slack.AttachmentField{
					slack.AttachmentField{
						Title: fmt.Sprintf("от %s:", comment.Author.DisplayName),
						Value: comment.Body,
						Short: false}}}}}

	if event, err := bus.NewEventWithData(e.Trace, bus.JsonCoding, bus.SlackPostEvent, msg); err != nil {
		return err
	} else {
		return ctx.Bus.Publish(*event)
	}
}
