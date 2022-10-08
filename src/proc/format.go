package proc

import (
	"bytes"
)

func Format(t *Task) (string, error) {
	var b bytes.Buffer

	if err := t.Src.Config.Template.Execute(&b, t); err != nil {
		return "", err
	}

	return b.String(), nil
}