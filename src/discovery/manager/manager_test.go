package manager

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/antibantique/pepe/src/discovery"
	"github.com/antibantique/pepe/src/proc"
	"github.com/antibantique/pepe/src/config"
)

func TestNew(t *testing.T) {
	docker := discovery.NewDocker("localhost", "")
	taskCh := make(chan *proc.Task)
	conf := config.C{}

	manager := New(docker, taskCh, conf)

	assert.NotNil(t, manager, "should be valid pointer")
}