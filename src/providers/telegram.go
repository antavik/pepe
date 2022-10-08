package providers

import (
	"time"
	"net/http"
	"strings"
	"context"
	"fmt"

	tb "gopkg.in/tucnak/telebot.v2"
	log "github.com/go-pkgz/lgr"
)

type TelegramProvider struct {
	Bot     *tb.Bot
	Timeout time.Duration
	ChatId  string

	limiter chan time.Time
}

func NewTelegramProvider(ctx context.Context, token, apiUrl, chatId string, timeout time.Duration) (*TelegramProvider, error) {
	log.Printf("[INFO] starting telegram provider for %s", apiUrl)

	if timeout == 0 {
		timeout = time.Second * 10
	}

	limiter := runLimiter(ctx)

	if token == "" {
		return &TelegramProvider{
			Bot:     nil,
			Timeout: timeout,
			ChatId:  chatId,

			limiter: limiter,
		}, nil
	}

	bot, err := tb.NewBot(tb.Settings{
		URL:    apiUrl,
		Token:  token,
		Client: &http.Client{Timeout: timeout},
	})
	if err != nil {
		return nil, err
	}

	return &TelegramProvider{
		Bot:     bot,
		Timeout: timeout,
		ChatId:  chatId,

		limiter: limiter,
	}, nil
}

func (tp TelegramProvider) Send(msg string) error {
	if tp.Bot == nil || tp.ChatId == "" {
		return nil
	}

	_, opened := <-tp.limiter
	if !opened {
		return fmt.Errorf("limiter not working")
	}

	if err := tp.sendText(tp.ChatId, msg); err != nil {
		return err
	}

	log.Printf("[DEBUG] message sent to telegram")

	return nil
}

func (tp TelegramProvider) sendText(channelId, msg string) error {
	_, err := tp.Bot.Send(
		recipient{chatId: channelId},
		msg,
		tb.ModeDefault,
		tb.NoPreview,
	)

	return err
}

type recipient struct {
	chatId string
}

func (r recipient) Recipient() string {
	if !strings.HasPrefix(r.chatId, "-100") && !strings.HasPrefix(r.chatId, "@") {
		return "@" + r.chatId
	}

	return r.chatId
}

func runLimiter(ctx context.Context) chan time.Time {
	limiter := make(chan time.Time, 1)
	limiter <- time.Now()

	go func() {
		ticker := time.NewTicker(time.Second * 2)
		defer ticker.Stop()
		defer close(limiter)

		for {
			select {
			case t := <-ticker.C:
				limiter <- t
			case <-ctx.Done():
				return
			}
		}
	}()

	return limiter
}