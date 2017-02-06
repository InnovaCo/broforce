package tasks

import (
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

type serve struct {
}

func (p *serve) serveRun(params serveParams, pType string, ctx *bus.Context, w io.Writer) error {
	args := []string{params.Plugin}
	for k, v := range params.Vars {
		args = append(args, "--var", fmt.Sprintf("%s=%s", k, v))
	}
	switch true {
	case strings.Compare(pType, bus.ServeCmdEvent) == 0:
		tmpfile, err := ioutil.TempFile("/tmp", "manifest_")
		if err != nil {
			return err
		}
		defer os.Remove(tmpfile.Name())
		ctx.Log.Debug("manifest\n", string(params.Manifest))

		if _, err := tmpfile.Write(params.Manifest); err != nil {
			return err
		}
		args = append(args, fmt.Sprintf("--manifest=%s", tmpfile.Name()))
	case strings.Compare(pType, bus.ServeCmdWithDataEvent) == 0:
		args = append(args, fmt.Sprintf("--plugin-data=%s", strings.Replace(string(params.Manifest), "\n", "", -1)))
	}

	cmd := exec.Command("serve", args...)

	ctx.Log.Info(cmd.Args)

	cmd.Stdout = w
	cmd.Stderr = w
	return cmd.Run()
}

func (p *serve) Run(ctx bus.Context) error {
	ctx.Bus.Subscribe(bus.ServeCmdEvent, bus.Context{Func: p.handler, Name: "ServeHandler"})
	ctx.Bus.Subscribe(bus.ServeCmdWithDataEvent, bus.Context{Func: p.handler, Name: "ServeHandler"})
	return nil
}

func (p *serve) handler(e bus.Event, ctx bus.Context) error {
	params := serveParams{}
	if err := e.Unmarshal(&params); err != nil {
		return err
	}
	w := newCmdWrite()
	defer ctx.Log.Info(string(w.Buffer))

	return p.serveRun(params, e.Subject, &ctx, io.Writer(w))
}

func newCmdWrite() *cmdWrite {
	return &cmdWrite{Buffer: make([]byte, 0)}
}

type cmdWrite struct {
	Buffer []byte
}

func (p *cmdWrite) Write(d []byte) (int, error) {
	p.Buffer = append(p.Buffer, d...)
	return len(d), nil
}
