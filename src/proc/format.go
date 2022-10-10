package proc

import (
	"bytes"
	"strings"
	"fmt"
)

type formatData struct {
	*Task
	Log string
}

func newFormatData(t *Task) *formatData {
	var strs []string

	for k, v := range t.RawLog {
		v := strings.TrimSpace(v)

		if k == "" {
			strs = append(strs, v)
		} else {
			strs = append(strs, fmt.Sprintf("%s: %s", k, v))
		}
	}

	return &formatData{ t, strings.Join(strs, ", "), }
}

func (f formatData) String() string {
	return fmt.Sprintf("Source: %s\nLog: %s", *f.Src, f.Log)
}

func Format(t *Task) (string, error) {
	var b bytes.Buffer

	if err := t.Src.Config.Template.Execute(&b, newFormatData(t)); err != nil {
		return "", err
	}

	return b.String(), nil
}