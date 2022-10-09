package proc

import (
	"github.com/antibantique/pepe/src/source"
)

type Task struct {
	Src    *source.S
	RawLog map[string]string
}