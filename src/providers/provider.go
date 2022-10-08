package providers

type Provider interface {
	Send(string) error
}