package proc

import (
	"github.com/antibantique/pepe/src/providers"
	"github.com/antibantique/pepe/src/source"
)

type Provider struct {
	Client providers.Provider
	Accept func(*source.S) bool
}