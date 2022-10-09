package proc

import (
	"bytes"
	"strings"
	"fmt"
)

type facade struct {
	*Task
	Log string
}

func NewFacade(t *Task) *facade {
	var strs []string
	var log string

	for k, v := range t.RawLog {
		v := strings.TrimSpace(v)

		if k == "" {
			log = v
		} else {
			log = fmt.Sprintf("%s: %s", k, v)
		}

		strs = append(strs, log)
	}

	return &facade{ t, strings.Join(strs, ", "), }
}

func (f facade) String() string {
	return fmt.Sprintf("Source: %s\nLog: %s", *f.Src, f.Log)
}

func Format(t *Task) (string, error) {
	var b bytes.Buffer

	if err := t.Src.Config.Template.Execute(&b, NewFacade(t)); err != nil {
		return "", err
	}

	return b.String(), nil
}