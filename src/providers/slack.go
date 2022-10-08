package providers

import (
	"time"
	"net/http"

	sl "github.com/slack-go/slack"
	log "github.com/go-pkgz/lgr"
)

type SlackProvider struct {
	Bot     *sl.Client
	Timeout time.Duration
	ChatId  string
}

func NewSlackProvider(token, apiUrl, chatId string, timeout time.Duration) *SlackProvider {
	log.Printf("[INFO] starting slack provider for %s", apiUrl)
	if timeout == 0 {
		timeout = time.Second * 60
	}

	if token == "" {
		return &SlackProvider{
			Bot:        nil,
			Timeout:    timeout,
			ChatId:     chatId,
		}
	}

	bot := sl.New(
		token,
		sl.OptionHTTPClient(&http.Client{Timeout: timeout}),
		sl.OptionAPIURL(apiUrl),
	)

	return &SlackProvider{
		Bot:     bot,
		Timeout: timeout,
		ChatId:  chatId,
	}
}

func (sp *SlackProvider) Send(msg string) error {
	if sp.Bot == nil || sp.ChatId == "" {
		return nil
	}

	if err := sp.sendText(sp.ChatId, msg); err != nil {
		return err
	}

	log.Printf("[DEBUG] message sent to slack")

	return nil
}

func (sp *SlackProvider) sendText(channelId, msg string) error {
	_, _, err := sp.Bot.PostMessage(
		channelId,
		sl.MsgOptionText(msg, false),
		sl.MsgOptionAsUser(true),
	)

	return err
}