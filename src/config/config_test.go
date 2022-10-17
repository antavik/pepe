package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeContainerConfig(t *testing.T) {
	commonConf := C{}
	labels := map[string]string{
		"pepe.template": "test",
		"pepe.telegram": "false",
		"pepe.slack":    "false",
		"pepe.stdout":   "true",
		"pepe.stderr":   "true",
		"pepe.regex":    "INFO",
		"test.test":     "test",
	}

	c := MakeContainerConfig("svc", commonConf, labels)

	assert.NotNil(t, c.Template)
	assert.False(t, c.TgEnabled)
	assert.False(t, c.SlEnabled)
	assert.True(t, c.Stdout)
	assert.True(t, c.Stderr)
	assert.True(t, c.Re.Match([]byte("INFO string")))
}

func TestMap(t *testing.T) {
	commonConf := C{}
	labels := map[string]string{
		"pepe.template": "test",
		"pepe.telegram": "false",
		"pepe.slack":    "false",
		"pepe.stdout":   "true",
		"pepe.stderr":   "true",
		"pepe.regex":    "INFO",
	}

	c := MakeContainerConfig("svc", commonConf, labels)
	m := c.Map()

	assert.Equal(t, c.TemplateRaw, m["template"])
	assert.Equal(t, c.TgEnabled,   m["telegram"])
	assert.Equal(t, c.SlEnabled,   m["slack"])
	assert.Equal(t, c.Stdout,      m["stdout"])
	assert.Equal(t, c.Stderr,      m["stderr"])
	assert.Equal(t, c.ReRaw,       m["regex"])
}