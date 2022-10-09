package manager

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/antibantique/pepe/src/discovery/docker"
	"github.com/antibantique/pepe/src/proc"
	"github.com/antibantique/pepe/src/config"
)

func TestNew(t *testing.T) {
	d := docker.New("localhost", "")
	taskCh := make(chan *proc.Task)
	conf := config.C{}

	manager := New(d, taskCh, conf)

	assert.NotNil(t, manager, "should be valid pointer")
}