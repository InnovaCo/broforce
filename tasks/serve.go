package tasks

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/InnovaCo/broforce/bus"
)

func init() {
	registry("serve", bus.Task(&serve{}))
}

//config section
//

type serve struct {
}

func (p *serve) serveRun(params *serveParams, event *bus.Event, w io.Writer) error {
	args := []string{params.Plugin}
	for k, v := range params.Vars {
		args = append(args, "--var", fmt.Sprintf("%s=%s", k, v))
	}
	switch true {
	case event.SubjectIs(bus.ServeCmdEvent):
		tmpfile, err := ioutil.TempFile("/tmp", "manifest_")
		if err != nil {
			return err
		}
		defer os.Remove(tmpfile.Name())
		w.Write(params.Manifest)
		if _, err := tmpfile.Write(params.Manifest); err != nil {
			return err
		}
		args = append(args, fmt.Sprintf("--manifest=%s", tmpfile.Name()))
	case event.SubjectIs(bus.ServeCmdWithDataEvent):
		args = append(args, fmt.Sprintf("--plugin-data=%s", strings.Replace(string(params.Manifest), "\n", "", -1)))
	}
	cmd := exec.Command("serve", args...)
	w.Write([]byte(strings.Join(cmd.Args, " ")))
	cmd.Stdout = w
	cmd.Stderr = w
	return cmd.Run()
}

func (p *serve) Run(ctx bus.Context) error {
	ctx.Bus.Subscribe(bus.ServeCmdEvent, bus.Context{
		Func:   p.handler,
		Name:   "ServeHandler",
		Bus:    ctx.Bus,
		Config: ctx.Config})
	ctx.Bus.Subscribe(bus.ServeCmdWithDataEvent, bus.Context{
		Func:   p.handler,
		Name:   "ServeHandler",
		Bus:    ctx.Bus,
		Config: ctx.Config})
	return nil
}

func (p *serve) handler(e bus.Event, ctx bus.Context) error {
	params := &serveParams{}
	if err := e.Unmarshal(params); err != nil {
		return err
	}
	buffer := bytes.NewBuffer(make([]byte, 0))
	err := p.serveRun(params, &e, io.Writer(buffer))
	ctx.Log.Info(buffer.String())
	return err
}
