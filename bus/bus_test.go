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

	logger.New(os.Stderr, "panic")
	cfg := config.New(tmpfile.Name(), config.YAMLAdapter)

	t.Run("SafeRun", func(t *testing.T) {
		Retry := uint32(0)

		f := func(eventBus *EventsBus, cfg config.ConfigData) error {
			Retry++
			return fmt.Errorf("Error number %d", Retry)
		}

		SafeRun(f, SafeParams{Retry: 3, Delay: 1})(&EventsBus{}, cfg.Get("test"))

		if Retry != 4 {
			t.Errorf("Retry %d != 4", Retry)
			t.Fail()
		}
	})
}
