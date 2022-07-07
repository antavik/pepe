package proc

import (
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"

	"github.com/antibantique/pepe/src/discovery"
)

func TestFormat(t *testing.T) {
	templ := template.Must(template.New("testTemplate").Parse("{{ .Name }}\n{{ index .Log `testField`}}"))
	tests := []struct{
		service *discovery.Service
		logData map[string][]string
		want    string
	}{
		// default formatting
		{
			&discovery.Service{"testId", "0.0.0.0", "testName", &discovery.Config{Template: nil,}},
			map[string][]string{"testField": {"testValue"}},
			"testName\ntestField: testValue",
		},
		// custom formatting
		{
			&discovery.Service{"testId", "0.0.0.0", "testName", &discovery.Config{Template: templ,}},
			map[string][]string{"testField": {"testValue"}},
			"testName\ntestValue",
		},
	}

	for _, tt := range tests {
		msg, err := format(tt.service, tt.logData)

		assert.NoError(t, err)
		assert.Equal(t, tt.want, *msg)
	}
}