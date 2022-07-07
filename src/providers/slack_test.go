package providers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	sl "github.com/slack-go/slack"
)

func TestNewSlackClient_EmptyToken(t *testing.T) {
	c := NewSlackClient("", "", "", 0)

	assert.Nil(t, c.Bot)
}

func TestSlackSend_BotNilOrEmptyChannelId(t *testing.T) {
	tests := []SlackClient{
		SlackClient{nil,          42, ""},
		SlackClient{&sl.Client{}, 42, ""},
	}

	for _, tt := range tests {
		err := tt.Send("")

		assert.Nil(t, err)
	}
}