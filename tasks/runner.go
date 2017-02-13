package tasks

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/InnovaCo/broforce/bus"
)

func init() {
	registry("runner", bus.Task(&runner{}))
}

//config section
//
//runnner:
//  path: path to scripts
//  map:
//   event_type:
//     - script_name
//     - scrip_name
//

type runner struct {
	interval time.Duration
}

func (p *runner) handler(script string) bus.Handler {
	return func(e bus.Event, ctx bus.Context) error {
		buffer := bytes.NewBuffer(make([]byte, 0))
		w := io.Writer(buffer)
		cmd := exec.Command("env", fmt.Sprintf("BROFORCE_EVENT='%s'", string(e.Data)), script)
		ctx.Log.Debug(strings.Join(cmd.Args, " "))
		cmd.Stdout = w
		cmd.Stderr = w
		err := cmd.Run()
		ctx.Log.Info(buffer)
		return err
	}
}

func (p *runner) Run(ctx bus.Context) error {
	path := ctx.Config.GetStringOr("path", "./")
	for k, _ := range ctx.Config.GetMap("map") {
		ctx.Log.Debug("KEY: ", k)
		for _, script := range ctx.Config.GetArrayString(fmt.Sprintf("map.%s", k)) {
			ctx.Log.Debug(filepath.Join(path, script))
			ctx.Bus.Subscribe(k, bus.Context{
				Func:   p.handler(filepath.Join(path, script)),
				Name:   fmt.Sprintf("%s%c%s", path, os.PathSeparator, script),
				Bus:    ctx.Bus,
				Config: ctx.Config})
		}
	}
	return nil
}
