package providers

import (
	"time"
	"net/http"

	sl "github.com/slack-go/slack"
	log "github.com/go-pkgz/lgr"
)

type SlackClient struct {
	Bot     *sl.Client
	Timeout time.Duration
	ChatId  string
}

func NewSlackClient(token, apiUrl, chatId string, timeout time.Duration) *SlackClient {
	log.Printf("[INFO] starting slack client for %s", apiUrl)
	if timeout == 0 {
		timeout = time.Second * 60
	}

	if token == "" {
		return &SlackClient{
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

	return &SlackClient{
		Bot:     bot,
		Timeout: timeout,
		ChatId:  chatId,
	}
}

func (sc *SlackClient) Send(msg string) error {
	if sc.Bot == nil || sc.ChatId == "" {
		return nil
	}

	if err := sc.sendText(sc.ChatId, msg); err != nil {
		return err
	}

	log.Printf("[DEBUG] message sent to slack")

	return nil
}

func (sc *SlackClient) sendText(channelId, msg string) error {
	_, _, err := sc.Bot.PostMessage(
		channelId,
		sl.MsgOptionText(msg, false),
		sl.MsgOptionAsUser(true),
	)

	return err
}