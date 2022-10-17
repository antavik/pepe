package providers

import (
	"github.com/antibantique/pepe/src/source"
)

type P interface {
	Send(string) error
	Accepted(*source.S) bool
}