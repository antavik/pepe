package proc

import (
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/antibantique/pepe/src/source"
	"github.com/antibantique/pepe/src/config"
)

func TestFormat_validTask(t *testing.T) {
	// valid template processing
	{
		template, err := template.New("test_template").Parse("{{ . }}")
		if err != nil {
			require.Error(t, err)
		}

		source := &source.S{
			Config: &config.C{
				Template: template,
			},
		}
		task := &Task{ Src: source, }

		result, err := Format(task)

		assert.NoError(t, err)
		assert.True(t, len(result) > 0, "should be non-empty sting")
	}
	// invalid template processing
	{
		template, err := template.New("test_template").Parse("{{ .test }}")
		if err != nil {
			require.Error(t, err)
		}

		source := &source.S{
			Config: &config.C{
				Template: template,
			},
		}
		task := &Task{ Src: source, }

		result, err := Format(task)

		assert.Error(t, err)
		assert.Empty(t, result, "should be empty sting")
	}
}