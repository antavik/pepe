package providers

import (
	"github.com/antibantique/pepe/src/source"
)

type FakeProvider struct {
	SendCalled     bool
	AcceptedCalled bool
}

func (fp *FakeProvider) Send(_ string) error {
	fp.SendCalled = true

	return nil
}

func (fp *FakeProvider) Accepted(_ *source.S) bool {
	fp.AcceptedCalled = true

	return true
}