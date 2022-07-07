package proc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/antibantique/pepe/src/discovery"
)

func TestInternalRun(t *testing.T) {
	tasksCh := make(chan *Task)
	errorsCh := make(chan error)

	processor := Processor{
		Services:  discovery.NewServiceManager(),
		Providers: []*Provider{
			&Provider{
				Supports: func(*discovery.Service) bool { return true },
			},
		},
	}

	go processor.run(tasksCh, errorsCh)
	close(tasksCh)

	_, opened := <-errorsCh

	assert.Empty(t, errorsCh)
	assert.False(t, opened)
}