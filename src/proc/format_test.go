package proc

import (
	"testing"
	"text/template"
	"fmt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/antibantique/pepe/src/source"
	"github.com/antibantique/pepe/src/config"
)

func newTestTask(log map[string]string, templ string) (*Task, error) {
	template, err := template.New("test_template").Parse(templ)
	if err != nil {
		return nil, err
	}

	source := &source.S{
		Config: &config.C{
			Template: template,
		},
	}

	return &Task{ Src: source, RawLog: log }, nil
}

func TestNewFormatData(t *testing.T) {
	tests := []struct{
		log  map[string]string
		want string
	}{
		{
			log:  map[string]string{"": "test log"},
			want: "test log",
		},
		{
			log:  map[string]string{"42": "test log"},
			want: "42: test log",
		},
	}

	for _, test := range tests {
		task, err := newTestTask(test.log, "{{ . }}")
		if err != nil {
			require.Error(t, err)
		}

		testFormatData := newFormatData(task)

		assert.NotNil(t, testFormatData)
		assert.Equal(t, test.want, testFormatData.Log)
	}
}

func TestNewFormatDataStringer(t *testing.T) {
	task, err := newTestTask(map[string]string{"": "test log"}, "{{ . }}")
	if err != nil {
		require.Error(t, err)
	}

	assert.Equal(t, "Source: \nLog: test log", fmt.Sprint(newFormatData(task)))
}

func TestFormat(t *testing.T) {
	// valid template processing
	{
		task, err := newTestTask(map[string]string{}, "{{ . }}")
		if err != nil {
			require.Error(t, err)
		}

		formatted, err := Format(task)

		assert.NoError(t, err)
		assert.True(t, len(formatted) > 0, "should be non-empty sting")
	}
	// invalid template processing
	{
		task, err := newTestTask(map[string]string{}, "{{ .test }}")
		if err != nil {
			require.Error(t, err)
		}

		formatted, err := Format(task)

		assert.Error(t, err)
		assert.Empty(t, formatted, "should be empty sting")
	}
}