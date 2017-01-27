package bus

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/InnovaCo/broforce/config"
	"github.com/InnovaCo/broforce/logger"
)

func TestSimple(t *testing.T) {
	tmpfile, err := ioutil.TempFile("/tmp", "config_")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	cfg := config.New(tmpfile.Name(), config.YAMLAdapter)
	logger.New(cfg.Get("logger"))
	ctx := Context{Log: logger.Logger4Handler("test", "")}

	t.Run("SafeHandler", func(t *testing.T) {
		Retry := uint32(0)

		f := func(e Event, ctx Context) error {
			Retry++
			return fmt.Errorf("Error number %d", Retry)
		}

		SafeHandler(f, SafeParams{Retry: 3, Delay: 1})(Event{}, ctx)

		if Retry != 4 {
			t.Errorf("Retry %d != 4", Retry)
			t.Fail()
		}
	})

	t.Run("PubSub", func(t *testing.T) {
		got := false
		a := &simpleAdapter{}
		a.Run()

		handler := func(e Event, ctx Context) error {
			got = true
			return nil
		}

		a.Subscribe(UnknownEvent, Context{Func: handler, Name: "TestHandler"})
		a.Publish(Event{Subject: UnknownEvent, Data: []byte(""), Coding: JsonCoding})

		time.Sleep(1 * time.Second)

		if !got {
			t.Error("Not got")
			t.Fail()
		}
	})
}
