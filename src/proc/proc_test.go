package proc

import (
	"testing"
	"context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/antibantique/pepe/src/providers"
)

func TestRun(t *testing.T) {
	ctx := context.Background()

	fp := providers.FakeProvider{}
	p := Proc{
		Provs: map[string]providers.P{ "": &fp },
		F:     func(_ *Task) (string, error) { return "", nil },
	}

	tCh := p.Run(ctx)
	defer close(tCh)

	require.NotNil(t, tCh)

	tCh <- &Task{}

	assert.True(t, fp.AcceptedCalled)
	assert.True(t, fp.SendCalled)
}