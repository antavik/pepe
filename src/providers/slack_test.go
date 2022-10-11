package providers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	sl "github.com/slack-go/slack"
)

func TestNewSlackProvider(t *testing.T) {
	c := NewSlackProvider("", "", "", 0)

	assert.Nil(t, c.Bot)
}

func TestSlackSend(t *testing.T) {
	tests := []SlackProvider{
		SlackProvider{nil,          42, ""},
		SlackProvider{&sl.Client{}, 42, ""},
	}

	for _, tt := range tests {
		err := tt.Send("")

		assert.Nil(t, err)
	}
}