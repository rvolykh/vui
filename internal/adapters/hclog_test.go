package adapters

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestHclogAdapter_Log(t *testing.T) {
	var ori = &logrus.Logger{
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.TraceLevel,
	}
	adapter := NewHclogAdapter(ori)

	table := []struct {
		name   string
		method func(string, ...interface{})
	}{
		{name: "trace", method: adapter.Trace},
		{name: "debug", method: adapter.Debug},
		{name: "info", method: adapter.Info},
		{name: "warn", method: adapter.Warn},
		{name: "error", method: adapter.Error},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ori.SetOutput(&buf)

			tt.method("this is test", "who", "programmer", "why", "testing")

			have := buf.String()

			assert.Contains(t, have, fmt.Sprintf("level=%s", tt.name))
			assert.Contains(t, have, `msg="this is test"`)
			assert.Contains(t, have, "who=programmer")
			assert.Contains(t, have, "why=testing")
		})
	}
}

func TestHclogAdapter_Level(t *testing.T) {
	var ori = &logrus.Logger{
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.TraceLevel,
	}
	adapter := NewHclogAdapter(ori)

	table := []struct {
		name      string
		level     hclog.Level
		wantTrace bool
		wantDebug bool
		wantInfo  bool
		wantWarn  bool
		wantError bool
	}{
		{name: "trace", level: hclog.Trace, wantTrace: true, wantDebug: true, wantInfo: true, wantWarn: true, wantError: true},
		{name: "debug", level: hclog.Debug, wantTrace: false, wantDebug: true, wantInfo: true, wantWarn: true, wantError: true},
		{name: "info", level: hclog.Info, wantTrace: false, wantDebug: false, wantInfo: true, wantWarn: true, wantError: true},
		{name: "warn", level: hclog.Warn, wantTrace: false, wantDebug: false, wantInfo: false, wantWarn: true, wantError: true},
		{name: "error", level: hclog.Error, wantTrace: false, wantDebug: false, wantInfo: false, wantWarn: false, wantError: true},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			adapter.SetLevel(tt.level)

			assert.Equal(t, tt.wantTrace, adapter.IsTrace())
			assert.Equal(t, tt.wantDebug, adapter.IsDebug())
			assert.Equal(t, tt.wantInfo, adapter.IsInfo())
			assert.Equal(t, tt.wantWarn, adapter.IsWarn())
			assert.Equal(t, tt.wantError, adapter.IsError())
		})

		t.Run("GetLevel", func(t *testing.T) {
			adapter.SetLevel(tt.level)

			assert.Equal(t, tt.level, adapter.GetLevel())
		})
	}
}

func TestHclogAdapter_Name(t *testing.T) {
	var ori = &logrus.Logger{
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.TraceLevel,
	}
	adapter := NewHclogAdapter(ori)

	t.Run("named", func(t *testing.T) {
		adapter.Named("test")

		assert.Equal(t, "test", adapter.Name())
	})

	t.Run("resetNamed", func(t *testing.T) {
		adapter.ResetNamed("test")

		assert.Equal(t, "logrusAdapter", adapter.Name())
	})
}

func TestHclogAdapter_With(t *testing.T) {
	var ori = &logrus.Logger{
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.TraceLevel,
	}
	adapter := NewHclogAdapter(ori)

	have := adapter.With("test", "test2").ImpliedArgs()

	assert.Equal(t, []interface{}{"test", "test2"}, have)
}

func TestHclogAdapter_StandardLogger(t *testing.T) {
	var ori = &logrus.Logger{
		Out:       &bytes.Buffer{},
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.TraceLevel,
	}
	adapter := NewHclogAdapter(ori)

	have := adapter.StandardLogger(nil)
	assert.NotNil(t, have)
	assert.Equal(t, ori.Out, have.Writer())
}
