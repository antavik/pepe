package manager

import (
	"context"

	"github.com/antibantique/pepe/src/source"
)

type Harvester struct {
	Source source.S
	Cancel context.CancelFunc
}