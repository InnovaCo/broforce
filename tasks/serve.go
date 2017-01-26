package tasks

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/InnovaCo/broforce/bus"
	"github.com/InnovaCo/broforce/config"
	"github.com/InnovaCo/broforce/logger"
)

func init() {
	registry("serve", bus.Task(&serve{}))
}

type serve struct {
	perm string
}

func (p serve) serveRun(params serveParams, pType string) error {
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
		logger.Log.Debug("manifest\n", string(params.Manifest))

		if _, err := tmpfile.Write(params.Manifest); err != nil {
			return err
		}
		args = append(args, fmt.Sprintf("--manifest=%s", tmpfile.Name()))
	case strings.Compare(pType, bus.ServeCmdWithDataEvent) == 0:
		args = append(args, fmt.Sprintf("--plugin-data=%s", strings.Replace(string(params.Manifest), "\n", "", -1)))
	}

	cmd := exec.Command("serve", args...)

	logger.Log.Info(cmd.Args)

	cmd.Stdout = logger.Log.Out
	cmd.Stderr = logger.Log.Out
	return cmd.Run()
}

func (p *serve) Run(eventBus *bus.EventsBus, cfg config.ConfigData) error {
	logger.Log.Debug(cfg.String())

	eventBus.Subscribe(bus.ServeCmdEvent, p.handler)
	eventBus.Subscribe(bus.ServeCmdWithDataEvent, p.handler)
	return nil
}

func (p serve) handler(e bus.Event) error {
	params := serveParams{}
	if err := bus.Encoder(e.Data, &params, e.Coding); err != nil {
		return err
	}
	if err := p.serveRun(params, e.Subject); err != nil {
		return err
	}
	return nil
}
