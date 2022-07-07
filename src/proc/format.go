package proc

import (
	"strings"
	"fmt"
	"bytes"

	"github.com/antibantique/pepe/src/discovery"
)

type merged struct {
	*discovery.Service
	Log map[string]string
}

func format(svc *discovery.Service, logData map[string][]string) (*string, error) {
	log := make(map[string]string, len(logData))
	for k, v := range logData {
		log[k] = strings.Join(v, ", ")
	}

	m := merged{
		Service: svc,
		Log:     log,
	}

	if m.Config.Template == nil {
		return defaultFormat(m), nil
	} else {
		return customFormat(m)
	}
}

func defaultFormat(m merged) *string {
	var strs []string

	strs = append(strs, m.Name)
	for k, v := range m.Log {
		strs = append(strs, fmt.Sprintf("%s: %s", k, v))
	}

	msg := strings.Join(strs, "\n")

	return &msg
}

func customFormat(m merged) (*string, error) {
	var b bytes.Buffer

	if err := m.Config.Template.Execute(&b, m); err != nil {
		return nil, err
	}

	msg := b.String()

	return &msg, nil
}