package providers

type Client interface {
	Send(string) error
}