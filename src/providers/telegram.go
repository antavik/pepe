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

type TelegramClient struct {
	Bot     *tb.Bot
	Timeout time.Duration
	ChatId string

	limiter   chan time.Time
}

func NewTelegramClient(ctx context.Context, token, apiUrl, chatId string, timeout time.Duration) (*TelegramClient, error) {
	log.Printf("[INFO] starting telegram client for %s", apiUrl)
	if timeout == 0 {
		timeout = time.Second * 60
	}

	limiter := runLimiter(ctx)

	if token == "" {
		return &TelegramClient{
			Bot:     nil,
			Timeout: timeout,
			ChatId: chatId,

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

	client := TelegramClient{
		Bot:     bot,
		Timeout: timeout,
		ChatId:  chatId,

		limiter: limiter,
	}

	return &client, nil
}

func (tc TelegramClient) Send(msg string) error {
	if tc.Bot == nil || tc.ChatId == "" {
		return nil
	}

	_, opened := <-tc.limiter
	if !opened {
		return fmt.Errorf("limiter not working")
	}

	if err := tc.sendText(tc.ChatId, msg); err != nil {
		return err
	}

	log.Printf("[DEBUG] message sent to telegram")

	return nil
}

func (tc TelegramClient) sendText(channelId, msg string) error {
	_, err := tc.Bot.Send(
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
		ticker := time.NewTicker(time.Second)
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