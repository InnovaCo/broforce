package bus

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/InnovaCo/broforce/config"
	"github.com/InnovaCo/broforce/logger"
)

func TestEventsBus(t *testing.T) {
	tmpfile, err := ioutil.TempFile("/tmp", "config_")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())
	cfg := config.New(tmpfile.Name(), config.YAMLAdapter)
	logger.New(cfg.Get("logger"))

	t.Run("SafeRun", func(t *testing.T) {
		Retry := uint32(0)
		f := func(ctx Context) error {
			Retry++
			return fmt.Errorf("Error number %d", Retry)
		}
		SafeRun(f, SafeParams{Retry: 3, Delay: 1})(Context{
			Name:   "test",
			Config: cfg.Get("test"),
			Bus:    &EventsBus{}})
		if Retry != 4 {
			t.Errorf("Retry %d != 4", Retry)
			t.Fail()
		}
	})
}
