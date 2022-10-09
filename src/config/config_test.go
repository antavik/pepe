package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeContainerConfig(t *testing.T) {
	commonConf := C{}
	labels := map[string]string{
		"pepe.template":        "test",
		"pepe.telegram":      "false",
		"pepe.slack":         "false",
		"pepe.stdout":        "true",
		"pepe.stderr":        "true",
		"pepe.regex":         "INFO",
		"test.test":          "test",
	}

	c := MakeContainerConfig("svc", commonConf, labels)

	assert.NotNil(t, c.Template)
	assert.False(t, c.TgEnabled)
	assert.False(t, c.SlEnabled)
	assert.True(t, c.Stdout)
	assert.True(t, c.Stderr)
	assert.True(t, c.Re.Match([]byte("INFO string")))
}