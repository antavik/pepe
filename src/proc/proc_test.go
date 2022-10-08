package proc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/antibantique/pepe/src/source"
)

func TestInternalRun(t *testing.T) {
	tasksCh := make(chan *Task)
	errorsCh := make(chan error)
	provs := map[string]*Provider{
		"telegrma": &Provider{
			Accept: func(*source.S) bool { return true },
		},
	}

	processor := Proc{ Providers: provs, }

	go processor.run(tasksCh, errorsCh)
	close(tasksCh)

	_, opened := <-errorsCh

	assert.Empty(t, errorsCh)
	assert.False(t, opened)
}