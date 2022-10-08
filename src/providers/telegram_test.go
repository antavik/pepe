package providers

import (
	"testing"
	"context"

	"github.com/stretchr/testify/assert"
	tb "gopkg.in/tucnak/telebot.v2"
)

func TestNewTelegramProvider(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := NewTelegramProvider(ctx, "", "", "", 0)

	assert.NoError(t, err)
	assert.Nil(t, client.Bot, "empty token, bot should be nil")
	assert.Len(t, client.limiter, 1)
}

func TestTelegramSend(t *testing.T) {
	tests := []TelegramProvider{
		TelegramProvider{},
		TelegramProvider{Bot: &tb.Bot{}},
	}

	for _, tt := range tests {
		err := tt.Send("")

		assert.Nil(t, err, "bot nil or empty chanId, send return nil")
	}
}

func TestTelegramSend_closeLimiter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	limiter := runLimiter(ctx)

	client := TelegramProvider{
		Bot:    &tb.Bot{},
		ChatId: "test",

		limiter: limiter,
	}

	// close and empty limiter
	cancel()
	<-limiter

	err := client.Send("")

	assert.Error(t, err)
}

func TestRecipient(t *testing.T) {
	tests := []struct {
		rec  recipient
		want string
	}{
		{recipient{"nickname"} , "@nickname"},  // nickname
		{recipient{"@nickname"}, "@nickname"},  // nickname with @
		{recipient{"-10042" }  , "-10042" },    // channel
	}

	for _, tt := range tests {
		r := tt.rec.Recipient()

		assert.Equal(t, tt.want, r)
	}
}

func TestRunLimiter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	limiter := runLimiter(ctx)

	assert.Len(t, limiter, 1)

	// close and empty limiter
	cancel()
	<- limiter

	_, opened := <-limiter

	assert.False(t, opened)
}