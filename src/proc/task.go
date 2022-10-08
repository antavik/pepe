package proc

import (
	"strings"
	"fmt"

	"github.com/antibantique/pepe/src/source"
)

type Task struct {
	Src *source.S
	Log map[string]string
}

func (t Task) String() string {
	var strs []string
	for k, v := range t.Log {
		strs = append(strs, fmt.Sprintf("%s: %s", k, strings.TrimSpace(v)))
	}

	var src string
	switch {
	case t.Src.Name != "":
		src = t.Src.Name
	case t.Src.Id != "":
		src = t.Src.Id
	default:
		src = t.Src.Ip
	}

	return fmt.Sprintf("Source: %s\nLog: %s", src, strings.Join(strs, ", "))
}